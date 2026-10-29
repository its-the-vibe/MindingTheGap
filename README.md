[![CI](https://github.com/its-the-vibe/MindingTheGap/actions/workflows/ci.yaml/badge.svg)](https://github.com/its-the-vibe/MindingTheGap/actions/workflows/ci.yaml)

# MindingTheGap
Tube Status Dashboard: line statuses and live station arrivals from the TfL Unified API.

## Prerequisites
- Go 1.27 (for local builds)
- Docker and Docker Compose (optional)
- A TfL API app key (optional, raises rate limits)

## Local setup
```sh
cp config.example.yaml config.yaml
cp .env.example .env   # set TFL_APP_KEY
make test
make build             # outputs bin/server
make run
```
Open http://localhost:8080. JSON is served at `/api/v1/status`.

Environment overrides: `PORT`, `TFL_BASE_URL`, `TFL_APP_KEY`.

## Docker Compose
```sh
cp config.example.yaml config.yaml
cp .env.example .env
docker compose up --build
```
The container runs as non-root with a read-only root filesystem.
