<div align="center">

![gouv.viz header](docs/img/header.png)

# gouv.viz

### Unofficial visualization of French Assemblée nationale open-data votes

A static-first web application that builds a local SQLite database from official open-data exports, then serves searchable pages for public votes, deputies, and political groups.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](go.mod)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](Dockerfile)
[![templ](https://img.shields.io/badge/templ-0.3.1001-111827)](https://templ.guide/)
[![SQLite](https://img.shields.io/badge/SQLite-Read--only-003B57?logo=sqlite)](internal/store/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

</div>

## Features

- Home dashboard with dataset volume, covered period, and recent public votes.
- Searchable and paginated list of public votes.
- Vote filters by legislature, result, vote type, parliamentary body, date range, and close votes.
- Vote detail pages with ballot metadata, group positions, and individual deputy votes.
- Deputy pages with mandate information, vote summaries, and searchable voting history.
- Political group pages with aggregated voting statistics and member lists.
- htmx-enhanced filtering and pagination with normal links and forms as a fallback.
- Local preprocessing pipeline from Assemblée nationale JSON exports to SQLite.
- Read-only web runtime with startup validation for schema and required tables.

## Data Sources

The preprocessing command expects the Assemblée nationale open-data JSON exports under `data/raw/`.

The helper script downloads the 17th legislature datasets used by the project:

- `Scrutins.json.zip` for public votes.
- `AMO20_dep_sen_min_tous_mandats_et_organes.json.zip` for actors, mandates, and organs.
- `Amendements.json.zip` for optional amendment links.
- `Dossiers_Legislatifs.json.zip` for optional legislative dossier links.

Generated data is written to `data/processed/gouv-viz.sqlite`. Raw and processed datasets are ignored by Git.

## Stack

- Go 1.25
- Echo for HTTP routing and middleware
- templ for server-rendered components
- htmx for progressive interactions
- SQLite through `modernc.org/sqlite`
- `golang-migrate` for embedded schema migrations
- Plain CSS bundled from `web/assets/css/src/`

## Requirements

- Go 1.25 or newer
- `templ` v0.3.1001
- `curl` and `unzip` to use `scripts/download-data.sh`
- Air, optional, for `make dev`
- Docker, optional, for container builds

Install `templ`:

```sh
go install github.com/a-h/templ/cmd/templ@v0.3.1001
```

## Quick Start

```sh
cp .env.example .env
./scripts/download-data.sh
make preprocess
make run
```

The web application listens on `http://localhost:9456` by default.

## Configuration

Configuration is read from environment variables. A local `.env` file is loaded when present.

| Variable | Default | Description |
| --- | --- | --- |
| `ENV` | `dev` | Runtime environment. Set to `prod` to disable development-only routes. |
| `PORT` | `9456` | HTTP port. |
| `ASSETS_PATH` | `web/assets` | Static assets directory served at `/assets`. |
| `DATABASE_PATH` | `data/processed/gouv-viz.sqlite` | SQLite database used by the web server. |

## Development Workflow

Generate CSS and templ files:

```sh
make generate
```

Run the web server:

```sh
make run
```

Run with Air live reload:

```sh
make dev
```

Run tests:

```sh
make test
```

Run the full local verification suite:

```sh
make verify
```

`make verify` rebuilds CSS, regenerates templ code, formats Go files, tidies modules, runs tests, runs `go vet`, and builds all packages.

## Data Pipeline

Download the current supported raw datasets:

```sh
./scripts/download-data.sh
```

Build the SQLite database:

```sh
make preprocess
```

Use custom input or output paths:

```sh
go run ./cmd/preprocess -raw data/raw -out data/processed/gouv-viz.sqlite
```

The preprocessing command creates a temporary SQLite file, applies embedded migrations, imports actors, organs, mandates, public votes, group vote totals, and individual votes, then atomically replaces the output database after validation.

Optional amendment and dossier data is used when `data/raw/amendements/json/` and `data/raw/dossiers/json/` exist.

## Performance Checks

`make perf-check` runs representative store queries against the generated SQLite database and reports slow queries or suspicious query plans.

```sh
make perf-check
go run ./cmd/storeperf -db data/processed/gouv-viz.sqlite -runs 5 -slow 500ms
go run ./cmd/storeperf -db data/processed/gouv-viz.sqlite -strict
```

Use `-strict` in CI-like checks when slow queries or plan issues should fail the command.

## Docker

The Docker image contains the web binary and static assets. It does not contain the generated Assemblée nationale database.

Build the database first, then mount `data/processed` at runtime:

```sh
make preprocess
docker build -t gouv-viz .
docker run --rm -p 9456:9456 -v "$PWD/data/processed:/data:ro" gouv-viz
```

The container uses these defaults:

```sh
ENV=prod
PORT=9456
ASSETS_PATH=web/assets
DATABASE_PATH=/data/gouv-viz.sqlite
```

At startup, the web process opens SQLite in read-only immutable mode and validates the expected schema version and required tables. Missing or incompatible database mounts fail at startup.

## HTTP Routes

| Route | Description |
| --- | --- |
| `/` | Home dashboard and recent votes. |
| `/scrutins` | Searchable public vote list. |
| `/scrutins/:uid` | Public vote detail page. |
| `/deputes` | Searchable deputy list. |
| `/deputes/:uid` | Deputy detail page. |
| `/groupes` | Political group list. |
| `/groupes/:uid` | Political group detail page. |
| `/ping` | Health check endpoint. |

In development mode, `/ws` is also registered for hot reload support.

## Repository Layout

```text
cmd/web/                 Web server entrypoint
cmd/preprocess/          JSON-to-SQLite preprocessing command
cmd/storeperf/           Store query timing and query-plan checks
internal/config/         Environment loading and validation
internal/dbmigration/    Embedded SQLite migrations
internal/store/          Read-side SQLite queries and page DTOs
internal/testdb/         Test helpers for migrated SQLite databases
web/handlers/            HTTP handlers and request parsing
web/components/          templ components and generated Go files
web/assets/              Static assets served by Echo
web/assets/css/src/      Source CSS modules
scripts/build-css.sh     CSS bundle builder
scripts/download-data.sh Assemblée nationale dataset downloader
data/fixtures/           Small committed fixtures for tests
data/raw/                Local raw open-data exports, ignored by Git
data/processed/          Generated SQLite output, ignored by Git
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
