package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"file_sys/backend/config"
	"file_sys/backend/internal/handler"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/router"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/storage"
	"file_sys/backend/internal/util"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Database (required)
	pool, err := util.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		log.Printf("Warning: migration error: %v", err)
	}

	// Redis (optional)
	if cfg.RedisEnabled {
		log.Printf("Redis enabled at %s", cfg.RedisAddr)
		// TODO: connect redis client when needed
	} else {
		log.Printf("Redis not configured, skipping (permission cache disabled)")
	}

	// Storage (required)
	var store storage.Storage
	if cfg.StorageDriver == "minio" {
		minioStore, err := storage.NewMinioStorage(cfg.MinioEndpoint, cfg.MinioAccessKey,
			cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
		if err != nil {
			log.Fatalf("Failed to connect to MinIO: %v", err)
		}
		store = minioStore
		log.Printf("MinIO storage enabled at %s (bucket: %s)", cfg.MinioEndpoint, cfg.MinioBucket)
	} else {
		store = storage.NewLocalStorage(cfg.StoragePath)
		log.Printf("Local storage enabled at %s", cfg.StoragePath)
	}

	// Repositories
	userRepo := repository.NewUserRepo(pool)
	tokenRepo := repository.NewRefreshTokenRepo(pool)
	folderRepo := repository.NewFolderRepo(pool)
	fileRepo := repository.NewFileRepo(pool)
	versionRepo := repository.NewVersionRepo(pool)
	permRepo := repository.NewPermissionRepo(pool)
	teamRepo := repository.NewTeamRepo(pool)

	// Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	folderService := service.NewFolderService(folderRepo, permRepo)
	fileService := service.NewFileService(fileRepo, versionRepo, folderRepo, permRepo, userRepo, store)
	teamService := service.NewTeamService(teamRepo, userRepo)
	searchService := service.NewSearchService(pool)

	// OnlyOffice (optional but enabled by default)
	var ooService *service.OnlyOfficeService
	if cfg.OnlyOfficeEnabled {
		ooService = service.NewOnlyOfficeService(
			fileService, 
			cfg.OnlyOfficeJWTSecret,
			cfg.OnlyOfficeDSURL, 
			cfg.OnlyOfficeCallbackURL, 
			pool, 
			cfg.OnlyOfficeTheme,
			cfg.OnlyOfficeJWTExpireHours,
			cfg.OnlyOfficeDocCache,
			cfg.OnlyOfficeLargeFileThresholdMB,
		)
		log.Printf("OnlyOffice enabled at %s (theme: %s, JWT expire: %dh, large file threshold: %dMB)", 
			cfg.OnlyOfficeDSURL, 
			cfg.OnlyOfficeTheme,
			cfg.OnlyOfficeJWTExpireHours,
			cfg.OnlyOfficeLargeFileThresholdMB,
		)
	} else {
		log.Printf("OnlyOffice not configured, online editing disabled")
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	authHandler.SetUserRepo(userRepo)
	fileHandler := handler.NewFileHandler(fileService, permRepo)
	folderHandler := handler.NewFolderHandler(folderService)
	versionHandler := handler.NewVersionHandler(fileService)
	trashHandler := handler.NewTrashHandler(fileService, folderService)
	teamHandler := handler.NewTeamHandler(teamService)
	ooHandler := handler.NewOnlyOfficeHandler(ooService)
	searchHandler := handler.NewSearchHandler(searchService)
	permHandler := handler.NewPermissionHandler(permRepo, fileService, folderService)

	// Rate limiter: 5 login attempts per minute per IP
	loginLimiter := middleware.NewRateLimiter(5, 1*time.Minute)
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			loginLimiter.Cleanup(10 * time.Minute)
		}
	}()

	// Setup router
	r := router.Setup(cfg, authHandler, fileHandler, folderHandler, teamHandler,
		ooHandler, versionHandler, searchHandler, trashHandler, permHandler, loginLimiter)

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
		log.Printf("Health check: http://localhost:%s/health", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS "ltree"`,
		`CREATE EXTENSION IF NOT EXISTS "pg_trgm"`,
		migrationUsers,
		migrationTeams,
		migrationFolders,
		migrationFiles,
		`ALTER TABLE files ADD COLUMN IF NOT EXISTS team_id UUID`,
		`CREATE INDEX IF NOT EXISTS idx_files_team ON files(team_id)`,
		migrationFileVersions,
		migrationPermissions,
		migrationRefreshTokens,
		migrationOOSessions,
		migrationOOSessionsModeUpdate,
		migrationJoinRequests,
	}
	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
