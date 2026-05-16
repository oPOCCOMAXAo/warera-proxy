# warera-proxy

Go proxy for the [Warera API](https://api2.warera.io). Forwards tRPC HTTP requests with built-in rate limiting and batch processing.

It is local drop-in replacement for actual warera api server. Setup local proxy and forget about rate-limiting and batching.

Batch size and number of requests per minute are configurable, so if actual warera limits changed, you can just update config.

## Docker

Compose example: [here](./docker-compose.example.yml)

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

### Processing pipeline

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
