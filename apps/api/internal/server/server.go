package server

import (
	"log"
	"net/http"

	"storage-api/internal/config"
	"storage-api/internal/db"
	"storage-api/internal/handlers"
	"storage-api/internal/middleware"
	"storage-api/internal/repository"
	"storage-api/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
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
	// Initialize Clerk for JWT verification
	if cfg.ClerkSecretKey != "" {
		clerk.SetKey(cfg.ClerkSecretKey)
	} else {
		log.Println("WARNING: CLERK_SECRET_KEY not set, authentication will fail")
	}

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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Household-ID"},
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
	mediaHandler := handlers.NewMediaHandler(mediaSvc, userSvc)
	householdsHandler := handlers.NewHouseholdsHandler(householdSvc)

	// Public routes (no auth required)
	s.router.Get("/health", healthHandler.Health)
	s.router.Get("/health/db", healthHandler.HealthDB)

	// Protected routes (require Clerk JWT)
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.ClerkAuth())

		// User routes
		r.Get("/me", userHandler.GetMe)

		// Logs routes
		r.Get("/logs/stream", logsHandler.StreamLogs)

		// Households routes
		r.Get("/households", householdsHandler.List)

		// Media routes
		r.Post("/media/upload", mediaHandler.Upload)
		r.Get("/media", mediaHandler.List)
		r.Get("/media/{id}", mediaHandler.Get)
		r.Get("/media/{id}/download", mediaHandler.Download)
		r.Get("/media/{id}/thumbnail", mediaHandler.Thumbnail)
		r.Get("/media/{id}/original", mediaHandler.Original)
		r.Delete("/media/{id}", mediaHandler.Delete)
	})
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
