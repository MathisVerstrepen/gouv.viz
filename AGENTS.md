# AGENTS.md

## Repo Shape

- This is a Go 1.25 static-first website for French Assemblee nationale "Scrutins publics" statistics.
- The web stack is Go + Echo + templ + htmx; templ components live in `web/components/` and render through Go handlers in `web/handlers/`.
- templ is less common than standard Go templates; if unsure about syntax or codegen, check https://templ.guide/ before guessing.
- The website entrypoint is `cmd/web/main.go`; the preprocessing entrypoint is `cmd/preprocess/main.go`.
- Static files live under `web/assets/`; Echo serves them at `/assets` using `ASSETS_PATH`.
- Local raw JSON belongs in `data/raw/scrutins-publics/`; generated preprocessing output belongs in `data/processed/`; both are ignored except `.gitkeep`.
- Only small committed samples should go in `data/fixtures/`.

## Commands

- Generate templ code after editing `.templ` files: `templ generate -path ./web/components`.
- Build the web binary: `make build` or `go build ./cmd/web`.
- Run the web app locally: `make run` or `go run ./cmd/web`.
- Run the preprocessing stub: `make preprocess` or `go run ./cmd/preprocess`.
- Tidy dependencies: `go mod tidy`.
- Focused verification for most changes: `templ generate -path ./web/components && gofmt -w ./cmd ./web && go mod tidy && go build ./...`.
- `make dev` requires Air and uses `.air.toml`; it regenerates templ and builds `./cmd/web` on changes.

## Gotchas

- Do not edit generated `*_templ.go` files directly; edit `.templ` files and regenerate.
- Keep generated `web/components/*_templ.go` in sync with `.templ` sources because the app imports the generated Go package.
- Do not commit large raw Assemblee nationale JSON or processed datasets; `.gitignore` is set up to keep those local.
