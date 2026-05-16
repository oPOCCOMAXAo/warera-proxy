package warera_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func newTestClient(
	t *testing.T,
	upstreamHandler http.HandlerFunc,
) *warera.Client {
	t.Helper()

	upstream := httptest.NewServer(upstreamHandler)
	t.Cleanup(upstream.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	client := warera.NewClient(warera.Config{
		Token:             "test-token",
		Host:              upstream.URL,
		BatchSize:         50,
		RequestsPerMinute: 100000,
	}, slog.New(slog.DiscardHandler))

	go func() {
		_ = client.Serve(ctx)
	}()

	return client
}

func mockUpstreamSuccess(
	w http.ResponseWriter,
	r *http.Request,
) {
	methods := strings.Split(
		strings.TrimPrefix(r.URL.Path, "/trpc/"),
		",",
	)

	w.Header().Set("Content-Type", "application/json")

	var b strings.Builder

	b.WriteByte('[')

	for i := range methods {
		if i > 0 {
			b.WriteByte(',')
		}

		b.WriteString(`{"result":{"data":"mock"}}`)
	}

	b.WriteByte(']')

	_, _ = w.Write([]byte(b.String()))
}

func TestProxyRequest_Single_Success(t *testing.T) {
	client := newTestClient(t, mockUpstreamSuccess)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "user.getUserLite",
		Input:   json.RawMessage(`{"userId":"123"}`),
	})

	if res.HTTPStatus != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.HTTPStatus)
	}
}

func TestProxyRequest_Single_InvalidJSON(t *testing.T) {
	client := newTestClient(t, mockUpstreamSuccess)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "user.getUserLite",
		Input:   []byte("{invalid}"),
	})

	if res.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", res.HTTPStatus)
	}

	var resp warera.Response

	err := json.Unmarshal(res.RawBody, &resp)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error in response")
	}

	if resp.Error.Code != warera.ErrorCodeInvalidJSON {
		t.Errorf("expected error code %d, got %d", warera.ErrorCodeInvalidJSON, resp.Error.Code)
	}
}

func TestProxyRequest_Batch_Success(t *testing.T) {
	client := newTestClient(t, mockUpstreamSuccess)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "user.getUserLite,user.getSettings",
		Batch:   "1",
		Input:   json.RawMessage(`{"0":{"userId":"123"},"1":{"userId":"123"}}`),
	})

	if res.HTTPStatus != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.HTTPStatus)
	}
}

func TestProxyRequest_Batch_InvalidJSON(t *testing.T) {
	client := newTestClient(t, mockUpstreamSuccess)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "user.getUserLite,user.getSettings",
		Batch:   "1",
		Input:   []byte("{invalid}"),
	})

	if res.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", res.HTTPStatus)
	}

	var results map[int]json.RawMessage

	err := json.Unmarshal(res.RawBody, &results)
	if err != nil {
		t.Fatalf("failed to parse batch response: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	for i, raw := range results {
		var resp warera.Response

		err := json.Unmarshal(raw, &resp)
		if err != nil {
			t.Fatalf("failed to parse result %d: %v", i, err)
		}

		if resp.Error == nil {
			t.Errorf("expected error in result %d", i)
		}
	}
}

func TestProxyRequest_Batch_MixedStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		methods := strings.Split(
			strings.TrimPrefix(r.URL.Path, "/trpc/"),
			",",
		)

		w.Header().Set("Content-Type", "application/json")

		var b strings.Builder

		b.WriteByte('[')

		for i, method := range methods {
			if i > 0 {
				b.WriteByte(',')
			}

			if method == "method1" {
				b.WriteString(
					`{"error":{"message":"not found","code":-32001,` +
						`"data":{"code":"NOT_FOUND","httpStatus":404,` +
						`"path":"method1"}}}`,
				)
			} else {
				b.WriteString(`{"result":{"data":"ok"}}`)
			}
		}

		b.WriteByte(']')

		_, _ = w.Write([]byte(b.String()))
	}

	client := newTestClient(t, handler)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "method1,method2",
		Batch:   "1",
		Input:   json.RawMessage(`{"0":{},"1":{}}`),
	})

	if res.HTTPStatus != http.StatusMultiStatus {
		t.Errorf("expected status 207, got %d", res.HTTPStatus)
	}
}

func TestProxyRequest_UpstreamError(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("not json"))
	}

	client := newTestClient(t, handler)

	res := client.ProxyRequest(context.Background(), warera.ProxyRequest{
		Methods: "user.getUserLite",
		Input:   json.RawMessage(`{"userId":"123"}`),
	})

	if res.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", res.HTTPStatus)
	}

	var resp warera.Response

	err := json.Unmarshal(res.RawBody, &resp)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error in response")
	}

	if resp.Error.Code != warera.ErrorCodeProxy {
		t.Errorf("expected error code %d, got %d", warera.ErrorCodeProxy, resp.Error.Code)
	}
}
