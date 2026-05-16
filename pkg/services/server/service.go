package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/pkg/errors"
)

var (
	_ lifecycle.Servable     = (*Service)(nil)
	_ lifecycle.Shutdownable = (*Service)(nil)
)

type Service struct {
	server *http.Server
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewService(
	config Config,
	logger *slog.Logger,
) *Service {
	var res Service

	res.logger = logger

	res.mux = http.NewServeMux()

	res.server = &http.Server{
		Handler:           res.mux,
		WriteTimeout:      time.Minute,
		ReadTimeout:       time.Minute,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: time.Minute,
		Addr:              ":" + strconv.FormatInt(config.Port, 10),
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	return &res
}

func (s *Service) Serve(ctx context.Context) error {
	s.logger.Info("server started",
		slog.String("address", s.server.Addr),
	)

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)

	return errors.WithStack(err)
}

func (s *Service) Mux() *http.ServeMux {
	return s.mux
}
