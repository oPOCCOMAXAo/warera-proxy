build:
	go build -o bin/proxy cmd/proxy/main.go

lint:
	golangci-lint-v2 run

build-docker:
	docker build -t poccomaxa/proxy:latest .

push-docker:
	docker push poccomaxa/proxy:latest
