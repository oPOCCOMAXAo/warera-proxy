package trpc

import (
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func (s *Service) TRPCGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	res := s.warera.ProxyRequest(r.Context(), warera.ProxyRequest{
		Methods: r.PathValue("methods"),
		Batch:   query.Get("batch"),
		Input:   []byte(query.Get("input")),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.HTTPStatus)
	_, _ = w.Write(res.RawBody)
}
