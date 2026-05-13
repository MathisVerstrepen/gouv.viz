# gouv.viz

Static-first website scaffold for visualizing statistics from French Assemblee nationale public votes.

## Stack

- Go HTTP server
- templ for server-rendered components
- htmx for progressive interactions
- Static assets served from `web/assets/`

## Structure

- `cmd/web/`: website server entrypoint
- `cmd/preprocess/`: future preprocessing command for raw public-vote JSON
- `web/handlers/`: HTTP handlers and render helpers
- `web/components/`: templ components
- `web/assets/`: CSS, JavaScript, images, and vendored static files
- `internal/assemblee/`: future data loading, parsing, and normalization code
- `internal/charts/`: future graph preparation code
- `data/raw/scrutins-publics/`: local raw Scrutins publics JSON dataset, ignored by Git
- `data/raw/amendements/`: optional extracted amendment open-data JSON files, ignored by Git
- `data/raw/dossiers/`: optional extracted dossier open-data JSON files, ignored by Git
- `data/processed/`: generated/preprocessed outputs, ignored by Git
- `data/fixtures/`: small sample data safe to commit

## Development

1. Copy `.env.example` to `.env` if you want local overrides.
2. Install `templ`: `go install github.com/a-h/templ/cmd/templ@v0.3.1001`.
3. Run `make generate` before building, or `make dev` with Air installed.
4. Put large raw Scrutins publics JSON files in `data/raw/scrutins-publics/`.
5. Run `./scripts/download-data.sh` to download and extract the open-data datasets before `make preprocess`.
6. Run `make perf-check` after `make preprocess` to time representative store queries against the generated SQLite database.

## Performance Checks

`make perf-check` runs `cmd/storeperf` against `data/processed/gouv-viz.sqlite`. It samples real scrutin, deputy, and political group IDs from the generated database, exercises representative store pages/details, and reports slow queries plus suspicious SQLite query plans such as full scans or automatic indexes.

Useful options:

```sh
go run ./cmd/storeperf -db data/processed/gouv-viz.sqlite -runs 5 -slow 500ms
go run ./cmd/storeperf -db data/processed/gouv-viz.sqlite -strict
```

## Docker

The container does not embed generated Assemblee nationale data. Build the SQLite database first, then mount it at `/data/gouv-viz.sqlite`:

```sh
make preprocess
docker build -t gouv-viz .
docker run --rm -p 9456:9456 -v "$PWD/data/processed:/data:ro" gouv-viz
```

Startup validates the database path, schema tables, and `dataset_meta.schema_version`, so a missing or incompatible mount fails fast.
