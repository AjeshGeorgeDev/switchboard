package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/integrations/harbor"
	"github.com/switchboard/switchboard/internal/integrations/trivy"
	"github.com/switchboard/switchboard/internal/notifications"
	"github.com/switchboard/switchboard/internal/settings"
)

const (
	TypeProcessHarborWebhook = "webhook:harbor"
	TypeProcessTrivyWebhook  = "webhook:trivy"
	TypeCVEPull              = "cve:pull"
	TypeNotifyDigest         = "notify:digest"
	TypeSendTeams            = "notify:teams"
	TypeSendEmail            = "notify:email"
	TypeRetentionCleanup     = "maintenance:retention"
)

type taskEnvelope struct {
	EventID string          `json:"event_id"`
	Body    json.RawMessage `json:"body"`
}

type Processor struct {
	queries *db.Queries
	cfg     config.Config
	notify  *notifications.Service
	trivy   *trivy.Client
}

func NewProcessor(pool *pgxpool.Pool, cfg config.Config, notify *notifications.Service) *Processor {
	return &Processor{
		queries: db.New(pool),
		cfg:     cfg,
		notify:  notify,
		trivy:   trivy.NewClient(cfg),
	}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeProcessHarborWebhook, p.handleHarborWebhook)
	mux.HandleFunc(TypeProcessTrivyWebhook, p.handleTrivyWebhook)
	mux.HandleFunc(TypeCVEPull, p.handleCVEPull)
	mux.HandleFunc(TypeNotifyDigest, p.handleNotifyDigest)
	mux.HandleFunc(TypeSendTeams, p.handleSendTeams)
	mux.HandleFunc(TypeSendEmail, p.handleSendEmail)
	mux.HandleFunc(TypeRetentionCleanup, p.handleRetentionCleanup)
}

func (p *Processor) handleHarborWebhook(ctx context.Context, t *asynq.Task) (err error) {
	_, body, eventID := unwrapTask(t.Payload())
	var cveNote string
	defer func() {
		p.markWebhookEvent(ctx, eventID, err)
		if err == nil && cveNote != "" && eventID != uuid.Nil {
			_ = p.queries.UpdateWebhookEventStatus(ctx, db.UpdateWebhookEventStatusParams{
				ID:           eventID,
				Status:       db.WebhookEventStatusProcessed,
				ErrorMessage: pgtype.Text{String: cveNote, Valid: true},
			})
		}
	}()

	harborCfg := settings.ResolveHarbor(ctx, p.queries, p.cfg)
	reports, parseErr := harbor.ParseDeploymentReports(body, harborCfg.URL)
	if parseErr != nil {
		return parseErr
	}

	var notes []string
	for _, report := range reports {
		hash := reportHash(body, report.DedupKey)
		if existing, err := p.queries.GetDeploymentReportByPayloadHash(ctx, pgtype.Text{String: hash, Valid: true}); err == nil && existing.ID.String() != "" {
			if note := p.ingestHarborCVEs(ctx, report); note != "" {
				notes = append(notes, note)
			}
			continue
		}

		_, err = p.queries.CreateDeploymentReport(ctx, db.CreateDeploymentReportParams{
			AppName:        report.AppName,
			ImageName:      report.ImageName,
			ImageTag:       report.ImageTag,
			TriggeredBy:    pgtype.Text{String: report.TriggeredBy, Valid: report.TriggeredBy != ""},
			Status:         string(report.Status),
			CriticalCount:  report.CriticalCount,
			HighCount:      report.HighCount,
			MediumCount:    report.MediumCount,
			LowCount:       report.LowCount,
			ReportUrl:      pgtype.Text{String: report.ReportURL, Valid: report.ReportURL != ""},
			RawPayload:     body,
			PayloadHash:    pgtype.Text{String: hash, Valid: true},
		})
		if err != nil {
			return err
		}
		_ = p.notify.Notify(ctx, notifications.Event{
			Type:     "deployment_report",
			Title:    fmt.Sprintf("Deployment: %s:%s", report.ImageName, report.ImageTag),
			Body:     fmt.Sprintf("Status: %s", report.Status),
			Severity: "info",
		})
		if note := p.ingestHarborCVEs(ctx, report); note != "" {
			notes = append(notes, note)
		}
	}
	if len(notes) > 0 {
		cveNote = strings.Join(notes, "; ")
	}
	return nil
}

// ingestHarborCVEs fetches per-CVE details from Harbor. Returns a non-empty note when
// enrichment was skipped or failed (webhook still succeeds).
func (p *Processor) ingestHarborCVEs(ctx context.Context, report harbor.DeploymentReportInput) string {
	harborCfg := settings.ResolveHarbor(ctx, p.queries, p.cfg)
	client := harbor.NewClient(harborCfg)
	if !client.Configured() {
		log.Printf("harbor CVE ingest skipped for %s:%s: Harbor API credentials not configured", report.ImageName, report.ImageTag)
		return "CVE ingest skipped: configure Harbor under Admin → Configuration"
	}
	if report.Digest == "" {
		log.Printf("harbor CVE ingest skipped for %s:%s: no artifact digest in webhook", report.ImageName, report.ImageTag)
		return "CVE ingest skipped: webhook payload has no artifact digest (need SCANNING_COMPLETED)"
	}
	project := report.Project
	repository := report.Repository
	if project == "" || repository == "" {
		project, repository = harbor.SplitRepoFullName(report.AppName)
	}
	findings, err := client.FetchArtifactVulnerabilities(ctx, project, repository, report.Digest)
	if err != nil {
		log.Printf("harbor CVE ingest failed for %s/%s@%s: %v", project, repository, report.Digest, err)
		return "CVE ingest failed: " + err.Error()
	}
	if len(findings) == 0 {
		log.Printf("harbor CVE ingest: no findings for %s/%s@%s", project, repository, report.Digest)
		return "CVE ingest: Harbor returned 0 vulnerabilities (check scan completed and robot has artifact-addition read)"
	}
	critical := false
	for _, f := range findings {
		sev := normalizeSeverity(strings.ToUpper(f.Severity))
		_, _ = p.queries.UpsertCVEFinding(ctx, db.UpsertCVEFindingParams{
			ImageName:        report.ImageName,
			ImageTag:         report.ImageTag,
			CveID:            f.CVEID,
			Severity:         sev,
			PackageName:      pgtype.Text{String: f.Package, Valid: f.Package != ""},
			InstalledVersion: pgtype.Text{String: f.InstalledVersion, Valid: f.InstalledVersion != ""},
			FixedVersion:     pgtype.Text{String: f.FixedVersion, Valid: f.FixedVersion != ""},
			Source:           "webhook",
			ScanDate:         time.Now(),
			RawPayload:       f.Raw,
		})
		if sev == "critical" {
			critical = true
		}
	}
	log.Printf("harbor CVE ingest: upserted %d findings for %s:%s", len(findings), report.ImageName, report.ImageTag)
	if critical {
		_ = p.notify.Notify(ctx, notifications.Event{
			Type:     "critical_cve",
			Title:    fmt.Sprintf("Critical CVE in %s:%s", report.ImageName, report.ImageTag),
			Body:     "Critical vulnerabilities detected via Harbor scan",
			Severity: "critical",
		})
	}
	return ""
}

func (p *Processor) handleTrivyWebhook(ctx context.Context, t *asynq.Task) (err error) {
	_, body, eventID := unwrapTask(t.Payload())
	defer func() { p.markWebhookEvent(ctx, eventID, err) }()

	var payload struct {
		ArtifactName string `json:"artifact_name"`
		Results      []struct {
			Vulnerabilities []struct {
				VulnerabilityID  string `json:"VulnerabilityID"`
				Severity         string `json:"Severity"`
				PkgName          string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion     string `json:"FixedVersion"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	imageName, imageTag := splitImage(payload.ArtifactName)
	critical := false
	for _, res := range payload.Results {
		for _, v := range res.Vulnerabilities {
			raw, _ := json.Marshal(v)
			_, _ = p.queries.UpsertCVEFinding(ctx, db.UpsertCVEFindingParams{
				ImageName:        imageName,
				ImageTag:         imageTag,
				CveID:            v.VulnerabilityID,
				Severity:         normalizeSeverity(v.Severity),
				PackageName:      pgtype.Text{String: v.PkgName, Valid: v.PkgName != ""},
				InstalledVersion: pgtype.Text{String: v.InstalledVersion, Valid: v.InstalledVersion != ""},
				FixedVersion:     pgtype.Text{String: v.FixedVersion, Valid: v.FixedVersion != ""},
				Source:           "webhook",
				ScanDate:         time.Now(),
				RawPayload:       raw,
			})
			if v.Severity == "CRITICAL" {
				critical = true
			}
		}
	}

	if critical {
		return p.notify.Notify(ctx, notifications.Event{
			Type:     "critical_cve",
			Title:    fmt.Sprintf("Critical CVE in %s:%s", imageName, imageTag),
			Body:     "Critical vulnerabilities detected in deployment scan",
			Severity: "critical",
		})
	}
	return nil
}

func (p *Processor) handleCVEPull(ctx context.Context, _ *asynq.Task) error {
	if !p.cfg.CVEPullEnabled {
		return nil
	}
	findings, err := p.trivy.FetchAllCVEs(ctx)
	if err != nil {
		return err
	}
	for _, f := range findings {
		raw, _ := json.Marshal(f)
		_, _ = p.queries.UpsertCVEFinding(ctx, db.UpsertCVEFindingParams{
			ImageName:        f.ImageName,
			ImageTag:         f.ImageTag,
			CveID:            f.CVEID,
			Severity:         f.Severity,
			PackageName:      pgtype.Text{String: f.Package, Valid: f.Package != ""},
			InstalledVersion: pgtype.Text{String: f.Installed, Valid: f.Installed != ""},
			FixedVersion:     pgtype.Text{String: f.Fixed, Valid: f.Fixed != ""},
			Source:           "weekly_pull",
			ScanDate:         time.Now(),
			RawPayload:       raw,
		})
	}
	if len(findings) == 0 {
		return nil
	}
	return p.notify.Notify(ctx, notifications.Event{
		Type:     "weekly_digest",
		Title:    "Weekly CVE Digest",
		Body:     fmt.Sprintf("CVE pull completed with %d findings processed", len(findings)),
		Severity: "info",
	})
}

func (p *Processor) handleNotifyDigest(ctx context.Context, _ *asynq.Task) error {
	return p.notify.Notify(ctx, notifications.Event{
		Type:     "weekly_digest",
		Title:    "Weekly CVE Digest",
		Body:     "Your weekly security digest is ready",
		Severity: "info",
	})
}

func (p *Processor) handleSendTeams(ctx context.Context, t *asynq.Task) error {
	return p.notify.SendTeamsPayload(ctx, t.Payload())
}

func (p *Processor) handleSendEmail(ctx context.Context, t *asynq.Task) error {
	return p.notify.SendEmailPayload(ctx, t.Payload())
}

func (p *Processor) handleRetentionCleanup(ctx context.Context, _ *asynq.Task) error {
	notifCutoff := time.Now().AddDate(0, 0, -p.cfg.NotificationRetention)
	cveCutoff := time.Now().AddDate(0, -p.cfg.CVERetentionMonths, 0)
	_ = p.queries.DeleteOldNotifications(ctx, notifCutoff)
	_ = p.queries.DeleteOldCVEFindings(ctx, cveCutoff)
	_ = p.queries.DeleteOldDeploymentReports(ctx, cveCutoff)
	_ = p.queries.DeleteOldWebhookEvents(ctx, cveCutoff)
	return nil
}

func unwrapTask(payload []byte) (taskEnvelope, []byte, uuid.UUID) {
	var env taskEnvelope
	if err := json.Unmarshal(payload, &env); err == nil && len(env.Body) > 0 {
		id, _ := uuid.Parse(env.EventID)
		return env, env.Body, id
	}
	return taskEnvelope{}, payload, uuid.Nil
}

func (p *Processor) markWebhookEvent(ctx context.Context, eventID uuid.UUID, procErr error) {
	if eventID == uuid.Nil {
		return
	}
	status := db.WebhookEventStatusProcessed
	var errMsg pgtype.Text
	if procErr != nil {
		status = db.WebhookEventStatusFailed
		errMsg = pgtype.Text{String: procErr.Error(), Valid: true}
	}
	_ = p.queries.UpdateWebhookEventStatus(ctx, db.UpdateWebhookEventStatusParams{
		ID:           eventID,
		Status:       status,
		ErrorMessage: errMsg,
	})
}

func reportHash(body []byte, dedupKey string) string {
	if dedupKey != "" {
		sum := sha256.Sum256([]byte(dedupKey))
		return hex.EncodeToString(sum[:])
	}
	return payloadHash(body)
}

func payloadHash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func splitImage(artifact string) (string, string) {
	for i := len(artifact) - 1; i >= 0; i-- {
		if artifact[i] == ':' {
			return artifact[:i], artifact[i+1:]
		}
	}
	return artifact, "latest"
}

func normalizeSeverity(s string) string {
	switch s {
	case "CRITICAL":
		return "critical"
	case "HIGH":
		return "high"
	case "MEDIUM":
		return "medium"
	case "LOW":
		return "low"
	default:
		return "unknown"
	}
}
