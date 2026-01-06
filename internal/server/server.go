package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"storage-api/internal/config"
	"storage-api/internal/handlers"
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
	healthHandler := handlers.NewHealthHandler(s.db)
	userHandler := handlers.NewUserHandler(s.db)

	s.router.Get("/health", healthHandler.Health)
	s.router.Get("/health/db", healthHandler.HealthDB)
	s.router.Get("/v1/me", userHandler.GetMe)
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

