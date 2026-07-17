package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/switchboard/switchboard/internal/audit"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/catalog"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/jobs"
	"github.com/switchboard/switchboard/internal/notifications"
	"github.com/switchboard/switchboard/internal/rbac"
	"github.com/switchboard/switchboard/internal/security"
	"github.com/switchboard/switchboard/internal/settings"
	"github.com/switchboard/switchboard/internal/setup"
	"github.com/switchboard/switchboard/internal/static"
	"github.com/switchboard/switchboard/internal/users"
	"github.com/switchboard/switchboard/internal/webhooks"
)

type Server struct {
	cfg    config.Config
	router http.Handler
	pool   *pgxpool.Pool
	asynq  *asynq.Server
}

func New(ctx context.Context, cfg config.Config) (*Server, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	enforcer, err := rbac.New(pool)
	if err != nil {
		return nil, err
	}

	queries := db.New(pool)
	tokens := auth.NewTokenService(cfg)
	sessions := auth.NewSessionService(queries, tokens)
	auditLog := audit.New(queries)
	authMW := auth.NewMiddleware(tokens, queries)
	localAuth := auth.NewLocalHandler(queries, sessions, tokens, auditLog)
	oidcAuth := auth.NewOIDCHandler(cfg, queries, sessions, tokens, auditLog)

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	asynqClient := asynq.NewClient(redisOpt)
	notifySvc := notifications.NewService(queries, cfg, asynqClient)
	jobProcessor := jobs.NewProcessor(pool, cfg, notifySvc)

	asynqServer := asynq.NewServer(redisOpt, asynq.Config{Concurrency: 5})
	mux := asynq.NewServeMux()
	jobProcessor.Register(mux)

	catalogH := catalog.NewHandler(queries)
	usersH := users.NewHandler(queries, enforcer, sessions, cfg, localAuth, auditLog)
	securityH := security.NewHandler(queries)
	notifyH := notifications.NewHandler(queries, cfg)
	webhookH := webhooks.NewHandler(asynqClient, cfg, queries)
	auditH := audit.NewHandler(queries)
	settingsH := settings.NewHandler(queries)
	setupH := setup.NewHandler(pool, queries, localAuth)
	setupGate := setup.BlockIfIncomplete(queries)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AppBaseURL, "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/api/settings/theme", settingsH.GetTheme)

	r.Route("/api/setup", func(r chi.Router) {
		r.Get("/status", setupH.Status)
		r.Post("/", setupH.Complete)
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.With(setupGate).Get("/providers", oidcAuth.ListProviders)
		r.With(setupGate).Post("/login", localAuth.Login)
		r.With(setupGate).Post("/refresh", localAuth.Refresh)
		r.Post("/logout", localAuth.Logout)
		r.With(setupGate).Get("/invite", usersH.PreviewInvite)
		r.With(setupGate).Post("/invite/accept", usersH.AcceptInvite)
		r.With(setupGate).Get("/oidc/{provider}/login", oidcAuth.Login)
		r.With(setupGate).Get("/oidc/{provider}/callback", oidcAuth.Callback)
		r.With(setupGate, authMW.RequireAuth).Get("/me", auth.MeHandler(localAuth))
	})

	r.Group(func(r chi.Router) {
		r.Use(setupGate)

	r.Route("/api/catalog", func(r chi.Router) {
		r.Get("/public", catalogH.ListPublic)
		r.Get("/sections", catalogH.ListSections)
		r.With(authMW.RequireAuth, rbac.RequirePermission(enforcer, auth.GetRolesFromRequest, "catalog", "read")).
			Get("/", catalogH.ListForUser)
	})

	r.Route("/api/security", func(r chi.Router) {
		r.Use(authMW.RequireAuth, rbac.RequirePermission(enforcer, auth.GetRolesFromRequest, "security", "read"))
		r.Get("/cves", securityH.ListCVEs)
		r.Get("/reports", securityH.ListReports)
		r.Get("/reports/{id}", securityH.GetReport)
	})

	r.Route("/api/notifications", func(r chi.Router) {
		r.Use(authMW.RequireAuth, rbac.RequirePermission(enforcer, auth.GetRolesFromRequest, "notifications", "read"))
		r.Get("/", notifyH.List)
		r.Patch("/read-all", notifyH.MarkAllRead)
		r.Patch("/{id}/read", notifyH.MarkRead)
	})

	r.Route("/api/profile", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Get("/notification-preferences", notifyH.GetPreferences)
		r.Patch("/notification-preferences", notifyH.UpdatePreferences)
		r.Get("/sessions", usersH.ListProfileLoginHistory)
		r.Get("/login-history", usersH.ListProfileLoginHistory)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(authMW.RequireAuth, rbac.RequirePermission(enforcer, auth.GetRolesFromRequest, "admin", "*"))
		r.Get("/applications", catalogH.ListAll)
		r.Get("/catalog/preview", catalogH.PreviewForRole)
		r.Get("/catalog-sections", catalogH.ListSections)
		r.Post("/catalog-sections", catalogH.CreateSection)
		r.Patch("/catalog-sections/{id}", catalogH.UpdateSection)
		r.Delete("/catalog-sections/{id}", catalogH.DeleteSection)
		r.Post("/applications", catalogH.Create)
		r.Patch("/applications/{id}", catalogH.Update)
		r.Delete("/applications/{id}", catalogH.Delete)
		r.Put("/applications/{id}/roles", catalogH.SetRoles)
		r.Get("/applications/{id}/roles", catalogH.GetRoles)

		r.Get("/roles", usersH.ListRoles)
		r.Post("/roles", usersH.CreateRole)
		r.Patch("/roles/{id}", usersH.UpdateRole)
		r.Delete("/roles/{id}", usersH.DeleteRole)

		r.Get("/users", usersH.ListUsers)
		r.Post("/users", usersH.CreateUser)
		r.Post("/users/invite", usersH.InviteUser)
		r.Get("/invitations", usersH.ListInvitations)
		r.Get("/users/{id}/roles", usersH.GetUserRoles)
		r.Patch("/users/{id}", usersH.UpdateUser)
		r.Put("/users/{id}/roles", usersH.SetUserRoles)
		r.Post("/users/{id}/force-logout", usersH.ForceLogout)
		r.Get("/users/{id}/sessions", usersH.ListSessions)
		r.Delete("/sessions/{id}", usersH.RevokeSession)

		r.Get("/oidc-providers", usersH.ListOIDCProviders)
		r.Post("/oidc-providers", usersH.CreateOIDCProvider)
		r.Patch("/oidc-providers/{id}", usersH.UpdateOIDCProvider)
		r.Delete("/oidc-providers/{id}", usersH.DeleteOIDCProvider)

		r.Get("/teams-webhooks", notifyH.ListTeamsWebhooks)
		r.Post("/teams-webhooks", notifyH.CreateTeamsWebhook)
		r.Patch("/teams-webhooks/{id}", notifyH.UpdateTeamsWebhook)
		r.Delete("/teams-webhooks/{id}", notifyH.DeleteTeamsWebhook)
		r.Get("/smtp-status", notifyH.SMTPStatus)
		r.Get("/webhook-endpoints", webhookH.Endpoints)
		r.Get("/webhook-events", webhookH.ListEvents)
		r.Get("/webhook-events/{id}", webhookH.GetEvent)
		r.Get("/audit-logs", auditH.List)
		r.Put("/settings/theme", settingsH.UpdateTheme)
	})

	}) // setupGate group

	r.Post("/webhooks/harbor", webhookH.Harbor)
	r.Post("/webhooks/trivy", webhookH.Trivy)

	r.Handle("/*", static.Handler())

	go func() {
		if err := asynqServer.Run(mux); err != nil {
			log.Printf("asynq server: %v", err)
		}
	}()

	return &Server{cfg: cfg, router: r, pool: pool, asynq: asynqServer}, nil
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Server) Close() {
	s.pool.Close()
	s.asynq.Shutdown()
}
