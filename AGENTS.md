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
