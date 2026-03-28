package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/query"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/infrastructure/postgres"
	redisInfra "github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/infrastructure/redis"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/interfaces/http/handler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/identity?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	if err := runMigrations(pool); err != nil {
		log.Printf("Warning: migration error: %v", err)
	}

	redisClient, err := redisInfra.NewRedisClient(ctx)
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}
	_ = redisClient

	userRepo := postgres.NewUserPostgresRepository(pool)
	grantRepo := postgres.NewAuditorAccessPostgresRepository(pool)
	knownCompanyRepo := postgres.NewKnownCompanyPostgresRepository(pool)
	_ = knownCompanyRepo

	tokenSvc := handler.NewJWTTokenService()
	publisher := &mockPublisher{}

	loginHandler := command.NewLoginHandler(userRepo, tokenSvc)
	refreshHandler := command.NewRefreshTokenHandler(tokenSvc)
	logoutHandler := command.NewLogoutHandler(tokenSvc)
	changePwHandler := command.NewChangePasswordHandler(userRepo, tokenSvc, publisher)
	forgotPwHandler := command.NewForgotPasswordHandler(userRepo, publisher)
	resetPwHandler := command.NewResetPasswordHandler(userRepo, publisher)
	createUserHandler := command.NewCreateUserHandler(userRepo, publisher)
	updateCertHandler := command.NewUpdateAuditorCertificateHandler(userRepo)
	grantHandler := command.NewGrantAuditorAccessHandler(grantRepo, publisher, userRepo)
	revokeHandler := command.NewRevokeAuditorAccessHandler(grantRepo, publisher)

	currentUserQuery := query.NewGetCurrentUserHandler(userRepo)
	listAccessQuery := query.NewListAuditorAccessHandler(grantRepo)

	authHandler := handler.NewAuthHandler(
		loginHandler, refreshHandler, logoutHandler,
		changePwHandler, forgotPwHandler, resetPwHandler,
		currentUserQuery,
	)
	userHandler := handler.NewUserHandler(createUserHandler, updateCertHandler)
	accessHandler := handler.NewAuditorAccessHandler(grantHandler, revokeHandler, listAccessQuery)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"openapi":"3.1.0","info":{"title":"Identity Service","version":"1.0.0"}}`))
	})

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("POST /v1/auth/change-password", authHandler.ChangePassword)
	mux.HandleFunc("GET /v1/users/me", authHandler.GetCurrentUser)
	mux.HandleFunc("POST /v1/users/system-admins", userHandler.CreateSystemAdmin)
	mux.HandleFunc("POST /v1/users/company-admins", userHandler.CreateCompanyAdmin)
	mux.HandleFunc("POST /v1/users/employees", userHandler.CreateEmployee)
	mux.HandleFunc("POST /v1/users/auditors", userHandler.CreateAuditor)
	mux.HandleFunc("PUT /v1/users/auditors/{id}/certificate", userHandler.UpdateAuditorCertificate)
	mux.HandleFunc("POST /v1/auditor-access", accessHandler.Grant)
	mux.HandleFunc("DELETE /v1/auditor-access/{grantId}", accessHandler.Revoke)
	mux.HandleFunc("GET /v1/auditor-access", accessHandler.List)

	loggingMux := loggingMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMux,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("Shutting down gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Identity Service starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func runMigrations(pool *pgxpool.Pool) error {
	migration := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(64) PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		role VARCHAR(32) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		must_change_password BOOLEAN NOT NULL DEFAULT true,
		company_id VARCHAR(64),
		branch_id VARCHAR(64),
		onboarding_completed BOOLEAN DEFAULT false,
		icp_certificate_chain TEXT,
		icp_certificate_serial VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS auditor_access_grants (
		id VARCHAR(64) PRIMARY KEY,
		auditor_id VARCHAR(64) NOT NULL,
		scope VARCHAR(32) NOT NULL,
		inventory_id VARCHAR(64),
		company_branch_id VARCHAR(64),
		company_id VARCHAR(64),
		granted_by VARCHAR(64) NOT NULL,
		granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS known_companies (
		id VARCHAR(64) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS device_tokens (
		user_id VARCHAR(64) NOT NULL,
		token VARCHAR(512) NOT NULL,
		platform VARCHAR(16) NOT NULL,
		registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, token)
	);

	CREATE TABLE IF NOT EXISTS outbox (
		id BIGSERIAL PRIMARY KEY,
		event_type VARCHAR(64) NOT NULL,
		payload JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		sent BOOLEAN NOT NULL DEFAULT false,
		sent_at TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);
	CREATE INDEX IF NOT EXISTS idx_auditor_access_auditor ON auditor_access_grants(auditor_id);
	CREATE INDEX IF NOT EXISTS idx_outbox_unsent ON outbox(sent, created_at);
	`
	_, err := pool.Exec(context.Background(), migration)
	return err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed: %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

type mockPublisher struct{}

func (m *mockPublisher) Publish(ctx context.Context, event user.DomainEvent) error {
	log.Printf("[PubSub] Event: %s Type: %s", event.EventID, event.EventType)
	return nil
}
