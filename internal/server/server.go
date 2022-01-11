package server

import (
	"net/http"

	"github.com/go-dummy/dummy/internal/config"
	"github.com/go-dummy/dummy/internal/logger"
	"github.com/go-dummy/dummy/internal/middleware"
)

// Server is struct for Server
type Server struct {
	Config   config.Server
	Server   *http.Server
	Logger   *logger.Logger
	Handlers Handlers
}

// NewServer returns a new instance of Server instance
func NewServer(config config.Server, l *logger.Logger, h Handlers) *Server {
	return &Server{
		Config:   config,
		Logger:   l,
		Handlers: h,
	}
}

// Run -.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.Handler)

	handler := middleware.Logging(mux, s.Logger)

	s.Server = &http.Server{
		Addr:    ":" + s.Config.Port,
		Handler: handler,
	}

	s.Logger.Info().Msgf("Running mock server on %s port", s.Config.Port)

	err := s.Server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
