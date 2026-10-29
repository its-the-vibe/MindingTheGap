.PHONY: build run test docker-build

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

run:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/server -config config.yaml

test:
	go test -v -race ./...

docker-build:
	docker build -t minding-the-gap .
