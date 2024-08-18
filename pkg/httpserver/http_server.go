package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/danielboakye/github-repo-stats/pkg/repository"
	"github.com/danielboakye/github-repo-stats/pkg/services/githubrepo"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Server represents an HTTP server
type Server struct {
	addr       string
	router     *chi.Mux
	logger     *slog.Logger
	repository repository.Repository
	githubSvc  *githubrepo.Service
}

// NewServer creates and returns a new Server instance
func NewServer(addr string, r repository.Repository, githubSvc *githubrepo.Service, logger *slog.Logger) *Server {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	s := &Server{
		addr:       addr,
		router:     router,
		logger:     logger,
		repository: r,
		githubSvc:  githubSvc,
	}

	s.RegisterRoutes()

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting-server", slog.String("url", "http://"+s.addr))
	return http.ListenAndServe(s.addr, s.router)
}
