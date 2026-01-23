package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"storage-api/internal/config"
	"storage-api/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config config.Config
	db     *pgxpool.Pool
	router *chi.Mux
}

func New(cfg config.Config) (*Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	s := &Server{
		config: cfg,
		db:     db,
		router: chi.NewRouter(),
	}

	s.routes()
	return s, nil
}

func (s *Server) routes() {
	// CORS middleware - allow frontend connections
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	healthHandler := handlers.NewHealthHandler(s.db)
	userHandler := handlers.NewUserHandler(s.db)
	logsHandler := handlers.NewLogsHandler()
	mediaHandler := handlers.NewMediaHandler(s.db)
	householdsHandler := handlers.NewHouseholdsHandler(s.db)

	s.router.Get("/health", healthHandler.Health)
	s.router.Get("/health/db", healthHandler.HealthDB)
	s.router.Get("/v1/me", userHandler.GetMe)
	s.router.Get("/v1/logs/stream", logsHandler.StreamLogs)

	// Households endpoint (for dev mode / household selection)
	s.router.Get("/v1/households", householdsHandler.List)

	// Media endpoints
	s.router.Post("/v1/media/upload", mediaHandler.Upload)
	s.router.Get("/v1/media", mediaHandler.List)
	s.router.Get("/v1/media/{id}", mediaHandler.Get)
	s.router.Get("/v1/media/{id}/download", mediaHandler.Download)
	s.router.Get("/v1/media/{id}/original", mediaHandler.Original)
	s.router.Delete("/v1/media/{id}", mediaHandler.Delete)
}

func (s *Server) Start() error {
	log.Printf("API listening on %s", s.config.Addr)
	return http.ListenAndServe(s.config.Addr, s.router)
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

