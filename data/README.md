# Data

Local datasets and generated files live here.

- `raw/scrutins-publics/`: raw JSON files from Assemblee nationale public votes.
- `raw/acteur/`: raw JSON files from Assemblee nationale deputies.
- `raw/organe/`: raw JSON files from Assemblee nationale bodies.
- `raw/amendements/`: extracted amendment JSON files used to enrich amendment PDF links.
- `raw/dossiers/`: extracted dossier JSON files used to enrich linked text and PDF links.
- `processed/`: generated/preprocessed files consumed by the website. The main runtime artifact is `processed/gouv-viz.sqlite`, generated with `make preprocess`.
- `fixtures/`: small representative samples safe to commit for tests and examples.

Large raw and processed datasets are ignored by Git. Keep only small fixtures committed.

The SQLite database is rebuilt from local raw JSON files in:

- `raw/scrutins-publics/`
- `raw/acteur/`
- `raw/organe/`
- `raw/amendements/json/` when present, for amendment PDF link enrichment
- `raw/dossiers/json/` when present, for dossier, linked text, and text PDF enrichment

Populate the raw data folders with:

```sh
./scripts/download-data.sh
```

The script downloads the official zip archives, refreshes `raw/scrutins-publics/`, `raw/acteur/`, `raw/organe/`, `raw/amendements/`, and `raw/dossiers/`, then extracts the JSON files into the expected layout.

Use `go run ./cmd/preprocess -raw data/raw -out data/processed/gouv-viz.sqlite` to override the default paths.

## Sources

- https://data.assemblee-nationale.fr/static/openData/repository/17/loi/scrutins/Scrutins.json.zip
- https://data.assemblee-nationale.fr/static/openData/repository/17/amo/deputes_senateurs_ministres_legislature/AMO20_dep_sen_min_tous_mandats_et_organes.json.zip
- https://data.assemblee-nationale.fr/static/openData/repository/17/loi/amendements_div_legis/Amendements.json.zip
- https://data.assemblee-nationale.fr/static/openData/repository/17/loi/dossiers_legislatifs/Dossiers_Legislatifs.json.zip
