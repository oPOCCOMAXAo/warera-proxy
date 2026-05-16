# warera-proxy

Go proxy for the [Warera API](https://api2.warera.io). Forwards tRPC HTTP requests with built-in rate limiting and batch processing.

It is local drop-in replacement for actual warera api server. Setup local proxy and forget about rate-limiting and batching.

Batch size and number of requests per minute are configurable, so if actual warera limits changed, you can just update config.

## Docker

### Compose example

```yml
name: "example"

services:
  we_proxy:
    container_name: we_proxy
    image: poccomaxa/warera-proxy:latest
    restart: unless-stopped
    environment:
      - WARERA_TOKEN=<your-token>
      - WARERA_BATCH_SIZE=50
      - WARERA_REQUESTS_PER_MINUTE=200
    ports:
      - 127.0.0.1:<PORT>:8080
```

Replace `<your-token>` with your warera API token and `<PORT>` with your preferred port.

### Build

```sh
docker build -t poccomaxa/warera-proxy:latest .
```

### Run

```sh
docker run -e WARERA_TOKEN=<your-token> -p 127.0.0.1:8080:8080 poccomaxa/warera-proxy:latest
```

The container exposes port `8080` and includes a health check at `/api/health` (checks every 30s).

It's recommended to keep internal port `8080` and change only external port. Also it is recommended to bind container only to localhost so only your local scripts can access it.

## Environment Variables

| Variable                     | Type     | Default                  | Required | Description                           |
| ---------------------------- | -------- | ------------------------ | -------- | ------------------------------------- |
| `WARERA_TOKEN`               | string   | -                        | **yes**  | API authentication token              |
| `WARERA_HOST`                | string   | `https://api2.warera.io` | no       | Upstream API base URL                 |
| `WARERA_BATCH_SIZE`          | int      | `50`                     | no       | Max methods per batched request       |
| `WARERA_REQUESTS_PER_MINUTE` | float    | `200.0`                  | no       | Rate limit for upstream requests      |
| `SERVER_PORT`                | int      | `8080`                   | no       | HTTP server listen port               |
| `LOGGER_DEBUG`               | bool     | `false`                  | no       | Enable debug logging                  |
| `LIFECYCLE_SHUTDOWN_TIMEOUT` | duration | `15s`                    | no       | Graceful shutdown timeout per service |

## tRPC HTTP Flow

### Endpoints

```
GET  /trpc/{methods}?input=...&batch=1
POST /trpc/{methods}?batch=1
```

### GET request

- `{methods}` - one or more comma-separated tRPC procedure names
- `input` query parameter - JSON-encoded procedure input
- `batch` query parameter - set to `1` to enable batch mode

```
GET /trpc/user.getUserLite,user.getUserLite?input={"userId":"..."}&batch=1
```

### POST request

- `{methods}` - tRPC procedure name
- Request body - JSON-encoded procedure input
- `batch` query parameter - set to `1` to enable batch mode

```
POST /trpc/user.getUserLite
Content-Type: application/json

{"userId": "..."}
```

### Errors

All responses are `Content-Type: application/json`. On proxy errors, the response follows the tRPC error format:

```json
{
  "error": {
    "message": "proxy: <description>",
    "code": -1000000000,
    "data": {
      "code": "PROXY",
      "httpStatus": 500,
      "path": ""
    }
  }
}
```

## Benchmark

Can achieve theoretical maximum 166 requests per second.

```txt
Warming up http://localhost:8084 ...
Benchmarking http://localhost:8084 for 2m0s with 100 goroutines (get) ...
Completed 20051 requests in 2m0.451s (166 req/s)

=== Benchmark Report ===
Target:         http://localhost:8084
Method:         get
Duration:       2m0s
Concurrency:    100
Total requests: 20051

--- Latency ---
  Min:  55.067345ms
  P50:  600.021319ms
  P95:  636.919146ms
  P99:  829.232335ms
  Max:  878.895575ms
  Avg:  599.980101ms

--- Status Codes ---
  200: 20051

Throughput: 167 req/s
Success rate: 100.0%
```
