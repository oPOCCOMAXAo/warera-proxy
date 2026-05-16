build:
	go build -o bin/proxy cmd/proxy/main.go

lint:
	golangci-lint-v2 run

build-docker:
	docker build -t poccomaxa/warera-proxy:latest .

push-docker:
	docker push poccomaxa/warera-proxy:latest
