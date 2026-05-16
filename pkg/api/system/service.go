package system

import (
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/api/internal"
)

var _ internal.Service = (*Service)(nil)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Register(mux *http.ServeMux) error {
	mux.HandleFunc("/api/health", s.Health)

	return nil
}
