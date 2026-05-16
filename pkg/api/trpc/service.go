package trpc

import (
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/api/internal"
	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

var _ internal.Service = (*Service)(nil)

type Service struct {
	warera *warera.Client
}

func NewService(
	warera *warera.Client,
) *Service {
	return &Service{
		warera: warera,
	}
}

func (s *Service) Register(mux *http.ServeMux) error {
	mux.HandleFunc("GET /trpc/{methods}", s.TRPCGet)
	mux.HandleFunc("POST /trpc/{methods}", s.TRPCPost)

	return nil
}
