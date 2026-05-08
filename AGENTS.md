# AGENTS.md

## Repo Shape

- This is a Go 1.25 static-first website for French Assemblee nationale "Scrutins publics" statistics.
- The web stack is Go + Echo + templ + htmx; templ components live in `web/components/` and render through Go handlers in `web/handlers/`.
- Read-side SQLite access for the web app lives in `internal/store/`; handlers should call store methods rather than embed SQL.
- templ is less common than standard Go templates; if unsure about syntax or codegen, check https://templ.guide/ before guessing.
- The website entrypoint is `cmd/web/main.go`; the preprocessing entrypoint is `cmd/preprocess/main.go`.
- Static files live under `web/assets/`; Echo serves them at `/assets` using `ASSETS_PATH`.
- Local raw JSON belongs in `data/raw/scrutins-publics/`; generated preprocessing output belongs in `data/processed/`; both are ignored except `.gitkeep`.
- Only small committed samples should go in `data/fixtures/`.

## App Architecture

- `cmd/web/main.go` owns process setup: environment initialization, Echo middleware, static assets, SQLite opening, store construction, and route registration.
- Open the web SQLite database once at startup with `database/sql`; pass the resulting `*sql.DB` into `internal/store.New` and close it when the process exits.
- Keep HTTP handlers in `web/handlers/` thin: parse request parameters, call the store, translate expected store errors into HTTP responses, and render templ components.
- Keep SQL, row scanning, pagination totals, search clauses, and whitelisted SQL sort definitions inside `internal/store/`, not in handlers or templ components.
- Use `QueryContext` and `QueryRowContext` in store methods with `ctx.Request().Context()` from handlers so request cancellation can reach SQLite calls.
- Keep dynamic SQL limited to safe fragments selected from internal allowlists, such as sort definitions; user values must still be passed as SQL parameters.
- Prefer adding focused store methods such as `HomePage`, `ScrutinsPage`, or `ScrutinDetailPage` over exposing generic query helpers to handlers.
- For now, store methods may return component page DTOs to keep the static-first rendering path simple; introduce separate domain/view models only when reuse or testing pressure justifies it.
- Expose data-layer sentinel errors from `internal/store` when handlers need HTTP-specific translation; do not import `database/sql` in handlers just to check `sql.ErrNoRows`.

## Commands

- Generate templ code after editing `.templ` files: `templ generate -path ./web/components`.
- Build the web binary: `make build` or `go build ./cmd/web`.
- Run the web app locally: `make run` or `go run ./cmd/web`.
- Run the preprocessing stub: `make preprocess` or `go run ./cmd/preprocess`.
- Tidy dependencies: `go mod tidy`.
- Focused verification for most changes: `templ generate -path ./web/components && gofmt -w ./cmd ./web && go mod tidy && go build ./...`.
- `make dev` requires Air and uses `.air.toml`; it regenerates templ and builds `./cmd/web` on changes.

## Design Guidelines

- Follow the French government design system fundamentals (DSFR) for visual decisions: neutral surfaces, Bleu France for primary actions, Rouge Marianne only for identity/accent use, and functional colors only for status meanings.
- Keep design tokens centralized in `web/assets/css/main.css`; add new colors, spacing, type sizes, borders, shadows, and breakpoints as CSS variables before using them in components.
- Prefer DSFR-style decision token names when practical, such as `--background-*`, `--text-*`, and `--border-*`, with project aliases like `--color-*` for local readability.
- Preserve the static-first approach: use semantic HTML in `.templ` files and avoid JavaScript for layout or visual behavior unless it is required for interaction.
- Design mobile-first and keep layouts responsive with the existing container, grid, spacing, and utility classes before adding new one-off classes.
- Maintain accessible defaults: visible focus states, sufficient contrast, text alternatives for visual status, skip-link support, and `forced-colors: active` handling for custom interactive or bordered components.
- Do not import the full DSFR package unless explicitly requested; this project currently mirrors DSFR fundamentals with local CSS rather than depending on DSFR assets.
- Use Marianne/Spectral font-family tokens as the preferred typography contract, with system fallbacks because local font files are not committed yet.
- Keep visual language sober, institutional, and data-oriented; avoid decorative patterns that reduce readability or make future data visualizations harder to scan.

## Gotchas

- Do not edit generated `*_templ.go` files directly; edit `.templ` files and regenerate.
- Keep generated `web/components/*_templ.go` in sync with `.templ` sources because the app imports the generated Go package.
- Do not commit large raw Assemblee nationale JSON or processed datasets; `.gitignore` is set up to keep those local.
