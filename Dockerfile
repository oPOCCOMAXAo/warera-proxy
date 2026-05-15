FROM golang:1.26-alpine3.23 AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/proxy ./cmd/proxy

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

RUN addgroup -S app && adduser -S -G app app

COPY --chown=app:app --from=builder /usr/local/bin/proxy /usr/local/bin/proxy

USER app
WORKDIR /home/app

EXPOSE 8080
ENV PORT=8080

# Healthcheck (uses busybox wget available in Alpine)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -q --spider http://127.0.0.1:8080/api/health || exit 1

ENTRYPOINT ["/usr/local/bin/proxy"]
