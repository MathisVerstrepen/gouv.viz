#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
RAW_DIR="$ROOT_DIR/data/raw"

SCRUTINS_URL="https://data.assemblee-nationale.fr/static/openData/repository/17/loi/scrutins/Scrutins.json.zip"
AMO_URL="https://data.assemblee-nationale.fr/static/openData/repository/17/amo/deputes_senateurs_ministres_legislature/AMO20_dep_sen_min_tous_mandats_et_organes.json.zip"
AMENDEMENTS_URL="https://data.assemblee-nationale.fr/static/openData/repository/17/loi/amendements_div_legis/Amendements.json.zip"
DOSSIERS_URL="https://data.assemblee-nationale.fr/static/openData/repository/17/loi/dossiers_legislatifs/Dossiers_Legislatifs.json.zip"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

download() {
  local url="$1"
  local output="$2"

  printf 'Downloading %s\n' "$url"
  curl --fail --location --retry 3 --retry-delay 2 --connect-timeout 20 --output "$output" "$url"
}

reset_json_dir() {
  local dir="$1"

  mkdir -p "$dir"
  find "$dir" -type f -name '*.json' -delete
}

reset_data_dir() {
  local dir="$1"

  mkdir -p "$dir"
  find "$dir" -mindepth 1 ! -name '.gitkeep' -exec rm -rf {} +
}

extract_json_flat() {
  local archive="$1"
  local pattern="$2"
  local target_dir="$3"

  reset_json_dir "$target_dir"
  unzip -q -j "$archive" "$pattern" -d "$target_dir"
}

extract_json_tree() {
  local archive="$1"
  local target_dir="$2"

  reset_data_dir "$target_dir"
  unzip -q "$archive" -d "$target_dir"
}

require_command curl
require_command unzip
require_command find

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p \
  "$RAW_DIR/scrutins-publics" \
  "$RAW_DIR/acteur" \
  "$RAW_DIR/organe" \
  "$RAW_DIR/amendements" \
  "$RAW_DIR/dossiers"

download "$SCRUTINS_URL" "$TMP_DIR/Scrutins.json.zip"
extract_json_flat "$TMP_DIR/Scrutins.json.zip" 'json/*.json' "$RAW_DIR/scrutins-publics"

download "$AMO_URL" "$TMP_DIR/AMO20_dep_sen_min_tous_mandats_et_organes.json.zip"
extract_json_flat "$TMP_DIR/AMO20_dep_sen_min_tous_mandats_et_organes.json.zip" 'json/acteur/*.json' "$RAW_DIR/acteur"
extract_json_flat "$TMP_DIR/AMO20_dep_sen_min_tous_mandats_et_organes.json.zip" 'json/organe/*.json' "$RAW_DIR/organe"

download "$AMENDEMENTS_URL" "$TMP_DIR/Amendements.json.zip"
extract_json_tree "$TMP_DIR/Amendements.json.zip" "$RAW_DIR/amendements"

download "$DOSSIERS_URL" "$TMP_DIR/Dossiers_Legislatifs.json.zip"
extract_json_tree "$TMP_DIR/Dossiers_Legislatifs.json.zip" "$RAW_DIR/dossiers"

printf 'Raw data populated in %s\n' "$RAW_DIR"
printf 'Next step: make preprocess\n'
