package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opoccomaxao/warera-proxy/pkg/api/trpc"
	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func newTRPCTestServer(
	t *testing.T,
	upstreamHandler http.HandlerFunc,
) *httptest.Server {
	t.Helper()

	upstream := httptest.NewServer(upstreamHandler)
	t.Cleanup(upstream.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	client := warera.NewClient(warera.Config{
		Token:             "test-token",
		Host:              upstream.URL,
		BatchSize:         50,
		RequestsPerMinute: 200,
	}, slog.New(slog.DiscardHandler))

	go func() {
		_ = client.Serve(ctx)
	}()

	svc := trpc.NewService(client)
	mux := http.NewServeMux()

	err := svc.Register(mux)
	if err != nil {
		t.Fatalf("failed to register service: %v", err)
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
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

func TestTRPCGet_SingleMethod(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		server.URL+`/trpc/user.getUserLite?input={"userId":"123"}`,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result json.RawMessage

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestTRPCGet_BatchMethod(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		server.URL+`/trpc/user.getUserLite,user.getSettings?batch=1&input={"0":{"userId":"123"},"1":{"userId":"123"}}`,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestTRPCGet_NoInput(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		server.URL+"/trpc/user.getUserLite",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty input, got %d", resp.StatusCode)
	}
}

func TestTRPCGet_InvalidInput(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		server.URL+`/trpc/user.getUserLite?input={bad}`,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error in response")
	}

	code, _ := errObj["code"].(float64)
	if code != warera.ErrorCodeInvalidJSON {
		t.Errorf("expected error code %d, got %v", warera.ErrorCodeInvalidJSON, code)
	}
}

func TestTRPCPost_SingleMethod(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL+"/trpc/user.getUserLite",
		bytes.NewReader([]byte(`{"userId":"123"}`)),
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestTRPCPost_BatchMethod(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL+"/trpc/user.getUserLite,user.getSettings?batch=1",
		bytes.NewReader([]byte(`{"0":{"userId":"123"},"1":{"userId":"456"}}`)),
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestTRPCPost_InvalidJSON(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL+"/trpc/user.getUserLite",
		bytes.NewReader([]byte(`{invalid}`)),
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error in response")
	}

	code, _ := errObj["code"].(float64)
	if code != warera.ErrorCodeInvalidJSON {
		t.Errorf("expected error code %d, got %v", warera.ErrorCodeInvalidJSON, code)
	}
}

func TestTRPCPost_EmptyBody(t *testing.T) {
	server := newTRPCTestServer(t, mockUpstreamSuccess)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL+"/trpc/user.getUserLite",
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty body, got %d", resp.StatusCode)
	}
}

func TestTRPCPost_UpstreamError(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}

	server := newTRPCTestServer(t, handler)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL+"/trpc/user.getUserLite",
		bytes.NewReader([]byte(`{"userId":"123"}`)),
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error in response")
	}

	code, _ := errObj["code"].(float64)
	if code != warera.ErrorCodeProxy {
		t.Errorf("expected proxy error code %d, got %v", warera.ErrorCodeProxy, code)
	}
}
