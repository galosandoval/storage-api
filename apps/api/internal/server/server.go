package server

import (
	"log"
	"net/http"

	"storage-api/internal/config"
	"storage-api/internal/db"
	"storage-api/internal/handlers"
	"storage-api/internal/repository"
	"storage-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

type Server struct {
	config config.Config
	db     *gorm.DB
	router *chi.Mux
}

func New(cfg config.Config) (*Server, error) {
	// Initialize GORM database
	gormDB, err := db.New(cfg.DSN)
	if err != nil {
		return nil, err
	}

	s := &Server{
		config: cfg,
		db:     gormDB,
		router: chi.NewRouter(),
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) setupRoutes() {
	// CORS middleware - allow frontend connections
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Household-ID", "X-Dev-User"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize repositories
	mediaRepo := repository.NewMediaRepository(s.db)
	userRepo := repository.NewUserRepository(s.db)
	householdRepo := repository.NewHouseholdRepository(s.db)

	// Initialize services
	mediaSvc := service.NewMediaService(mediaRepo)
	userSvc := service.NewUserService(userRepo)
	householdSvc := service.NewHouseholdService(householdRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(s.db)
	userHandler := handlers.NewUserHandler(userSvc)
	logsHandler := handlers.NewLogsHandler()
	mediaHandler := handlers.NewMediaHandler(mediaSvc)
	householdsHandler := handlers.NewHouseholdsHandler(householdSvc)

	// Health routes
	s.router.Get("/health", healthHandler.Health)
	s.router.Get("/health/db", healthHandler.HealthDB)

	// User routes
	s.router.Get("/v1/me", userHandler.GetMe)

	// Logs routes
	s.router.Get("/v1/logs/stream", logsHandler.StreamLogs)

	// Households routes
	s.router.Get("/v1/households", householdsHandler.List)

	// Media routes
	s.router.Post("/v1/media/upload", mediaHandler.Upload)
	s.router.Get("/v1/media", mediaHandler.List)
	s.router.Get("/v1/media/{id}", mediaHandler.Get)
	s.router.Get("/v1/media/{id}/download", mediaHandler.Download)
	s.router.Get("/v1/media/{id}/thumbnail", mediaHandler.Thumbnail)
	s.router.Get("/v1/media/{id}/original", mediaHandler.Original)
	s.router.Delete("/v1/media/{id}", mediaHandler.Delete)
}

func (s *Server) Start() error {
	log.Printf("API listening on %s", s.config.Addr)
	return http.ListenAndServe(s.config.Addr, s.router)
}

func (s *Server) Close() {
	if s.db != nil {
		db.Close(s.db)
	}
}
