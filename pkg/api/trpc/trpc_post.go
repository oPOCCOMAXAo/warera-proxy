package trpc

import (
	"io"
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func (s *Service) TRPCPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(warera.ProxyError("failed to read body"))

		return
	}

	res := s.warera.ProxyRequest(r.Context(), warera.ProxyRequest{
		Methods: r.PathValue("methods"),
		Batch:   query.Get("batch"),
		Input:   body,
	})

	w.WriteHeader(res.HTTPStatus)
	_, _ = w.Write(res.RawBody)
}
