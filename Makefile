VERSION ?= $(shell git describe --tags --always --dirty)

build:
	go build -o bin/proxy cmd/proxy/main.go

lint:
	golangci-lint-v2 run

test:
	go test -race ./...

bench:
	go run ./cmd/bench \
	-addr http://localhost:8084 \
	-method get \
	-duration 120s \
	-concurrency 100

build-docker:
	docker build -t poccomaxa/warera-proxy:latest -t poccomaxa/warera-proxy:$(VERSION) .

push-docker:
	docker push poccomaxa/warera-proxy:latest
	docker push poccomaxa/warera-proxy:$(VERSION)
