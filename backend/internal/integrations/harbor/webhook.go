package harbor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/switchboard/switchboard/internal/db"
)

// WebhookEvent is the native Harbor v2 webhook envelope.
type WebhookEvent struct {
	Type      string    `json:"type"`
	OccurAt   int64     `json:"occur_at"`
	Operator  string    `json:"operator"`
	EventData EventData `json:"event_data"`
}

type EventData struct {
	Resources  []Resource `json:"resources"`
	Repository Repository `json:"repository"`
}

type Resource struct {
	ResourceURL  string                       `json:"resource_url"`
	Tag          string                       `json:"tag"`
	Digest       string                       `json:"digest"`
	ScanOverview map[string]json.RawMessage   `json:"scan_overview"`
}

type Repository struct {
	Name         string `json:"name"`
	RepoFullName string `json:"repo_full_name"`
	Namespace    string `json:"namespace"`
}

type scanOverviewEntry struct {
	ScanStatus string `json:"scan_status"`
	ReportID   string `json:"report_id"`
	Summary    struct {
		Summary map[string]int `json:"summary"`
	} `json:"summary"`
}

// DeploymentReportInput is the normalized internal model for a deployment report row.
type DeploymentReportInput struct {
	AppName       string
	ImageName     string
	ImageTag      string
	TriggeredBy   string
	Status        db.DeployStatus
	CriticalCount int32
	HighCount     int32
	MediumCount   int32
	LowCount      int32
	ReportURL     string
	DedupKey      string
}

// ParseDeploymentReports parses Harbor v2 native events or legacy flat JSON payloads.
func ParseDeploymentReports(payload []byte, harborBaseURL string) ([]DeploymentReportInput, error) {
	var native WebhookEvent
	if err := json.Unmarshal(payload, &native); err != nil {
		return nil, err
	}
	if native.Type != "" && (len(native.EventData.Resources) > 0 || native.EventData.Repository.RepoFullName != "") {
		return parseNativeEvent(native, harborBaseURL), nil
	}
	return parseLegacyPayload(payload)
}

func parseNativeEvent(ev WebhookEvent, harborBaseURL string) []DeploymentReportInput {
	repoName := ev.EventData.Repository.RepoFullName
	if repoName == "" {
		repoName = ev.EventData.Repository.Name
	}
	if repoName == "" {
		repoName = "unknown"
	}

	operator := ev.Operator
	if operator == "" {
		operator = "webhook"
	}

	resources := ev.EventData.Resources
	if len(resources) == 0 {
		resources = []Resource{{}}
	}

	out := make([]DeploymentReportInput, 0, len(resources))
	for _, res := range resources {
		imageName, imageTag := resolveImage(res, repoName)
		status, crit, high, med, low := statusFromEvent(ev.Type, res)
		reportURL := ""
		if rid := scanReportID(res); rid != "" && harborBaseURL != "" {
			reportURL = strings.TrimRight(harborBaseURL, "/") + "/harbor/projects/" + strings.Split(repoName, "/")[0] + "/repositories/" + strings.TrimPrefix(repoName, strings.Split(repoName, "/")[0]+"/") + "/artifacts-tab/scan/overview"
			_ = reportURL // harbor UI paths vary; store report_id in dedup key instead
			reportURL = strings.TrimRight(harborBaseURL, "/") + "/harbor/repositories/" + repoName + "/artifacts-tab/scan/overview"
		}
		if rid := scanReportID(res); rid != "" {
			reportURL = strings.TrimRight(harborBaseURL, "/") + "/api/v2.0/projects/" + strings.Split(repoName, "/")[0] + "/repositories/" + pathAfterFirst(repoName) + "/artifacts/" + res.Digest + "/additions/vulnerabilities"
			if harborBaseURL == "" {
				reportURL = "harbor-report:" + rid
			}
		}

		dedup := fmt.Sprintf("%s:%s:%s:%s:%d", ev.Type, imageName, imageTag, res.Digest, ev.OccurAt)
		out = append(out, DeploymentReportInput{
			AppName:       repoName,
			ImageName:     imageName,
			ImageTag:      imageTag,
			TriggeredBy:   operator,
			Status:        status,
			CriticalCount: crit,
			HighCount:     high,
			MediumCount:   med,
			LowCount:      low,
			ReportURL:     reportURL,
			DedupKey:      dedup,
		})
	}
	return out
}

func pathAfterFirst(s string) string {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) < 2 {
		return parts[0]
	}
	return parts[1]
}

func resolveImage(res Resource, repoName string) (imageName, imageTag string) {
	if res.ResourceURL != "" {
		imageName, imageTag = ParseResourceURL(res.ResourceURL)
	}
	if imageName == "" {
		imageName = repoName
	}
	if res.Tag != "" {
		imageTag = res.Tag
	}
	if imageTag == "" {
		if res.Digest != "" {
			imageTag = "@" + strings.TrimPrefix(res.Digest, "sha256:")
			if len(imageTag) > 20 {
				imageTag = imageTag[:20]
			}
		} else {
			imageTag = "latest"
		}
	}
	return imageName, imageTag
}

func scanReportID(res Resource) string {
	for _, raw := range res.ScanOverview {
		var entry scanOverviewEntry
		if err := json.Unmarshal(raw, &entry); err == nil && entry.ReportID != "" {
			return entry.ReportID
		}
	}
	return ""
}

func statusFromEvent(eventType string, res Resource) (status db.DeployStatus, crit, high, med, low int32) {
	switch eventType {
	case "SCANNING_FAILED":
		return db.DeployStatusFailed, 0, 0, 0, 0
	case "PUSH_ARTIFACT", "PULL_ARTIFACT", "DELETE_ARTIFACT":
		return db.DeployStatusSuccess, 0, 0, 0, 0
	case "SCANNING_COMPLETED":
		scanStatus := ""
		for _, raw := range res.ScanOverview {
			var entry scanOverviewEntry
			if err := json.Unmarshal(raw, &entry); err == nil {
				scanStatus = entry.ScanStatus
				crit, high, med, low = severityCounts(entry.Summary.Summary)
				break
			}
		}
		return NormalizeDeployStatus(scanStatus, crit > 0 || high > 0), crit, high, med, low
	default:
		return db.DeployStatusSuccess, 0, 0, 0, 0
	}
}

func severityCounts(summary map[string]int) (crit, high, med, low int32) {
	if summary == nil {
		return 0, 0, 0, 0
	}
	return int32(summary["Critical"]), int32(summary["High"]), int32(summary["Medium"]), int32(summary["Low"])
}

// NormalizeDeployStatus maps Harbor scan_status strings to deploy_status enum values.
func NormalizeDeployStatus(scanStatus string, hasHighSeverity bool) db.DeployStatus {
	switch strings.ToLower(strings.TrimSpace(scanStatus)) {
	case "success", "completed", "done":
		if hasHighSeverity {
			return db.DeployStatusPartial
		}
		return db.DeployStatusSuccess
	case "error", "failed", "failure":
		return db.DeployStatusFailed
	case "running", "pending", "scheduled":
		return db.DeployStatusPartial
	default:
		if scanStatus == "" {
			return db.DeployStatusSuccess
		}
		if hasHighSeverity {
			return db.DeployStatusPartial
		}
		return db.DeployStatusSuccess
	}
}

// ParseResourceURL extracts image name and tag from a Harbor resource_url.
func ParseResourceURL(url string) (imageName, imageTag string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", ""
	}
	// Strip scheme if present
	if idx := strings.Index(url, "://"); idx >= 0 {
		url = url[idx+3:]
	}
	// host/project/repo:tag or host/project/repo@sha256:digest
	if atIdx := strings.Index(url, "@"); atIdx > 0 {
		return url[:atIdx], url[atIdx:]
	}
	colonIdx := strings.LastIndex(url, ":")
	if colonIdx > 0 {
		slashAfterColon := strings.Contains(url[colonIdx:], "/")
		if !slashAfterColon {
			return url[:colonIdx], url[colonIdx+1:]
		}
	}
	return url, ""
}

func parseLegacyPayload(payload []byte) ([]DeploymentReportInput, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	appName := strField(m, "app_name", "repository_name", "unknown")
	imageName := strField(m, "image_name", "resource_url", appName)
	imageTag := strField(m, "image_tag", "tag", "latest")
	triggeredBy := strField(m, "triggered_by", "operator", "webhook")
	rawStatus := strField(m, "status", "result", "success")
	status := NormalizeDeployStatus(rawStatus, intField(m, "critical_count") > 0 || intField(m, "high_count") > 0)
	if rawStatus == "failed" || rawStatus == "failure" {
		status = db.DeployStatusFailed
	} else if rawStatus == "partial" {
		status = db.DeployStatusPartial
	} else if rawStatus == "success" {
		if intField(m, "critical_count") > 0 || intField(m, "high_count") > 0 {
			status = db.DeployStatusPartial
		} else {
			status = db.DeployStatusSuccess
		}
	}

	return []DeploymentReportInput{{
		AppName:       appName,
		ImageName:     imageName,
		ImageTag:      imageTag,
		TriggeredBy:   triggeredBy,
		Status:        status,
		CriticalCount: intField(m, "critical_count"),
		HighCount:     intField(m, "high_count"),
		MediumCount:   intField(m, "medium_count"),
		LowCount:      intField(m, "low_count"),
		ReportURL:     strField(m, "report_url"),
		DedupKey:      "",
	}}, nil
}

func strField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func intField(m map[string]interface{}, key string) int32 {
	if v, ok := m[key].(float64); ok {
		return int32(v)
	}
	return 0
}
