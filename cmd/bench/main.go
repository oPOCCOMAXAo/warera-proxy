package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultConcurrency = 10
	resultBufferCap    = 1024
	percentMultiplier  = 100
	warmupRequests     = 5
	warmupRetryDelay   = 500 * time.Millisecond
)

var errUnknownMethod = errors.New("unknown method")

type result struct {
	statusCode int
	latency    time.Duration
	err        error
}

type config struct {
	addr        string
	duration    time.Duration
	concurrency int
	method      string
	input       string
	methods     string
}

func main() {
	log.SetFlags(0)

	cfg := parseFlags()

	results := runBenchmark(cfg)
	printReport(results, cfg)
}

func parseFlags() config {
	var cfg config

	var duration string

	flag.StringVar(&cfg.addr, "addr", "http://localhost:8080", "target server address")
	flag.StringVar(&duration, "duration", "30s", "benchmark duration")
	flag.IntVar(
		&cfg.concurrency,
		"concurrency",
		defaultConcurrency,
		"number of concurrent goroutines",
	)
	flag.StringVar(&cfg.method, "method", "post", "HTTP method: get, post, get-batch, post-batch")
	flag.StringVar(
		&cfg.input,
		"input",
		`{"userId":"6907535cdc9c77c3d9a6aff3"}`,
		"JSON input payload",
	)
	flag.StringVar(
		&cfg.methods,
		"methods",
		"user.getUserLite",
		"tRPC methods (comma-separated for batch)",
	)
	flag.Parse()

	d, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatalf("invalid duration %q: %v", duration, err)
	}

	cfg.duration = d

	return cfg
}

func runBenchmark(cfg config) []result {
	log.Printf("Warming up %s ...", cfg.addr)

	warmup(cfg)

	log.Printf("Benchmarking %s for %s with %d goroutines (%s) ...",
		cfg.addr, cfg.duration, cfg.concurrency, cfg.method)

	var (
		wg         sync.WaitGroup
		resultsMux sync.Mutex
		allRes     []result
		stop       atomic.Bool
		start      = time.Now()
		totalReqs  atomic.Int64
	)

	for range cfg.concurrency {
		wg.Go(func() {
			localResults := make([]result, 0, resultBufferCap)

			for !stop.Load() {
				reqStart := time.Now()
				res := sendRequest(cfg)
				res.latency = time.Since(reqStart)

				totalReqs.Add(1)

				localResults = append(localResults, res)
			}

			resultsMux.Lock()

			allRes = append(allRes, localResults...)
			resultsMux.Unlock()
		})
	}

	time.AfterFunc(cfg.duration, func() { stop.Store(true) })
	wg.Wait()

	elapsed := time.Since(start)

	log.Printf("Completed %d requests in %s (%.0f req/s)",
		totalReqs.Load(), elapsed.Round(time.Millisecond),
		float64(totalReqs.Load())/elapsed.Seconds())

	return allRes
}

func warmup(cfg config) {
	for range warmupRequests {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			cfg.addr+"/api/health",
			nil,
		)
		if err != nil {
			log.Fatalf("warmup request creation failed: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("warmup request failed, is the server running? %v", err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			time.Sleep(warmupRetryDelay)
		}
	}
}

func sendRequest(cfg config) result {
	switch cfg.method {
	case "get":
		return sendGet(cfg, false)
	case "get-batch":
		return sendGet(cfg, true)
	case "post":
		return sendPost(cfg, false)
	case "post-batch":
		return sendPost(cfg, true)
	default:
		return result{err: errUnknownMethod}
	}
}

func sendGet(cfg config, batch bool) result {
	u, err := url.Parse(cfg.addr + "/trpc/" + cfg.methods)
	if err != nil {
		return result{err: err}
	}

	q := u.Query()

	if batch {
		q.Set("batch", "1")
	}

	q.Set("input", cfg.input)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return result{err: err}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result{err: err}
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return result{statusCode: resp.StatusCode}
}

func sendPost(cfg config, batch bool) result {
	u := cfg.addr + "/trpc/" + cfg.methods

	if batch {
		u += "?batch=1"
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		u,
		bytes.NewReader([]byte(cfg.input)),
	)
	if err != nil {
		return result{err: err}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result{err: err}
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return result{statusCode: resp.StatusCode}
}

func printReport(results []result, cfg config) {
	var (
		errCount    int
		latencies   []time.Duration
		statusCodes = make(map[int]int)
	)

	for _, r := range results {
		if r.err != nil {
			errCount++

			continue
		}

		latencies = append(latencies, r.latency)
		statusCodes[r.statusCode]++
	}

	log.Println()
	log.Println("=== Benchmark Report ===")
	log.Println("Target:         " + cfg.addr)
	log.Println("Method:         " + cfg.method)
	log.Println("Duration:       " + cfg.duration.String())
	log.Println("Concurrency:    " + strconv.Itoa(cfg.concurrency))
	log.Println("Total requests: " + strconv.Itoa(len(results)))
	log.Println()

	if len(latencies) == 0 {
		log.Println("No successful requests to report.")

		return
	}

	printLatencyReport(latencies)
	printStatusCodes(statusCodes)

	if errCount > 0 {
		log.Println()
		log.Println("--- Errors: " + strconv.Itoa(errCount) + " ---")
	}

	log.Println()

	total := len(latencies) + errCount
	throughput := float64(total) / cfg.duration.Seconds()
	successRate := float64(len(latencies)) / float64(total) * percentMultiplier

	log.Println("Throughput: " + strconv.FormatFloat(throughput, 'f', 0, 64) + " req/s")
	log.Println("Success rate: " + strconv.FormatFloat(successRate, 'f', 1, 64) + "%")
}

func printLatencyReport(latencies []time.Duration) {
	slices.Sort(latencies)

	log.Println("--- Latency ---")
	log.Println("  Min:  " + latencies[0].String())
	log.Println("  P50:  " + latencies[len(latencies)*50/percentMultiplier].String())
	log.Println("  P95:  " + latencies[len(latencies)*95/percentMultiplier].String())
	log.Println("  P99:  " + latencies[len(latencies)*99/percentMultiplier].String())
	log.Println("  Max:  " + latencies[len(latencies)-1].String())
	log.Println("  Avg:  " + avgDuration(latencies).String())
	log.Println()
}

func printStatusCodes(statusCodes map[int]int) {
	log.Println("--- Status Codes ---")

	codes := make([]int, 0, len(statusCodes))
	for code := range statusCodes {
		codes = append(codes, code)
	}

	slices.Sort(codes)

	for _, code := range codes {
		log.Println("  " + strconv.Itoa(code) + ": " + strconv.Itoa(statusCodes[code]))
	}
}

func avgDuration(durations []time.Duration) time.Duration {
	var total time.Duration

	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}
