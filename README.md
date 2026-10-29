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

Environment overrides: `PORT`, `TFL_BASE_URL`, `TFL_APP_KEY`, and
`DASHBOARD_REFRESH_INTERVAL_SECONDS`. Set `dashboard.refresh_interval_seconds`
in the config file to change how often the dashboard refreshes (default: 30 seconds).

## Monitoring arrivals
Configure each monitored station as `<line>/<stationId>` under `tfl.monitored_stations`.
Each entry triggers a separate TfL request to `/line/{line}/arrivals/{stationId}`.
List a station more than once to monitor it on multiple lines; each result is keyed by its line/station pair.
The `/api/v1/status` response keeps `arrivals` as a time-sorted array for compatibility and also
provides `arrivalsByDirection`, keyed by the same line/station pair and then by each arrival's
TfL `direction` value (e.g. `northbound`, `southbound`, `clockwise`). Arrivals in each direction remain
sorted by time to arrival; arrivals without a `direction` are grouped under `Unknown direction`.
The dashboard shows each direction under a line-and-direction heading (e.g. "Jubilee - Inbound").

```yaml
tfl:
  monitored_stations:
    - northern/940GZZLUOXC
    - victoria/940GZZLUOXC
```

## Docker Compose
```sh
cp config.example.yaml config.yaml
cp .env.example .env
docker compose up --build
```
The container runs as non-root with a read-only root filesystem.
