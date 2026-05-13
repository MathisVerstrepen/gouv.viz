package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gouv.viz/internal/dbmigration"

	_ "modernc.org/sqlite"
)

const schemaVersion = dbmigration.LatestVersionString

type stats struct {
	Organes            int
	Acteurs            int
	Adresses           int
	Mandats            int
	MandatOrganes      int
	Scrutins           int
	ScrutinGroupeVotes int
	Votes              int
	Diagnostics        importDiagnostics
}

type rawFile struct {
	Path       string
	SourcePath string
	SourceHash string
	Root       map[string]any
}

func main() {
	rawDir := flag.String("raw", filepath.Join("data", "raw"), "raw Assemblée nationale data directory")
	outPath := flag.String("out", filepath.Join("data", "processed", "gouv-viz.sqlite"), "output SQLite database path")
	flag.Parse()

	options := buildOptions{}
	amendementsPath, hasAmendements := amendementsDirectoryPath(*rawDir)
	if hasAmendements {
		resolver, err := newDirectoryAmendementResolver(amendementsPath)
		if err != nil {
			log.Fatal(err)
		}
		options.AmendementResolver = resolver
	}
	dossiersPath, hasDossiers := dossiersDirectoryPath(*rawDir)
	if hasDossiers {
		resolver, err := newDirectoryDossierResolver(dossiersPath)
		if err != nil {
			log.Fatal(err)
		}
		options.DossierResolver = resolver
	}

	result, err := buildDatabaseWithOptions(*rawDir, *outPath, options)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("generated %s", *outPath)
	log.Printf("organes=%d acteurs=%d adresses=%d mandats=%d mandat_organes=%d scrutins=%d scrutin_groupe_votes=%d votes=%d",
		result.Organes,
		result.Acteurs,
		result.Adresses,
		result.Mandats,
		result.MandatOrganes,
		result.Scrutins,
		result.ScrutinGroupeVotes,
		result.Votes,
	)
	if report := result.Diagnostics.Report(); report != "" {
		log.Print(report)
	}
}

func amendementsDirectoryPath(rawDir string) (string, bool) {
	path := filepath.Join(rawDir, "amendements")
	jsonPath := filepath.Join(path, "json")
	if info, err := os.Stat(jsonPath); err == nil && info.IsDir() {
		return path, true
	}
	return "", false
}

func dossiersDirectoryPath(rawDir string) (string, bool) {
	path := filepath.Join(rawDir, "dossiers")
	jsonPath := filepath.Join(path, "json")
	if info, err := os.Stat(jsonPath); err == nil && info.IsDir() {
		return path, true
	}
	return "", false
}

func buildDatabase(rawDir, outPath string) (stats, error) {
	return buildDatabaseWithOptions(rawDir, outPath, buildOptions{})
}

type buildOptions struct {
	AmendementResolver amendementLinkResolver
	DossierResolver    dossierLinkResolver
}

func buildDatabaseWithOptions(rawDir, outPath string, options buildOptions) (stats, error) {
	var result stats

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return result, fmt.Errorf("create output directory: %w", err)
	}

	tmpPath := outPath + ".tmp"
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return result, fmt.Errorf("remove previous temporary database: %w", err)
	}

	db, err := sql.Open("sqlite", tmpPath)
	if err != nil {
		return result, fmt.Errorf("open sqlite database: %w", err)
	}
	defer db.Close()

	if err := configureDatabase(db); err != nil {
		return result, err
	}
	if err := createSchema(db); err != nil {
		return result, err
	}

	tx, err := db.Begin()
	if err != nil {
		return result, fmt.Errorf("begin import transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := importOrganes(tx, filepath.Join(rawDir, "organe"), &result, &result.Diagnostics); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if err := insertSyntheticOrganes(tx, &result); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if err := importActeurs(tx, filepath.Join(rawDir, "acteur"), &result, &result.Diagnostics); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if err := importScrutins(tx, filepath.Join(rawDir, "scrutins-publics"), &result, options.AmendementResolver, options.DossierResolver, &result.Diagnostics); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if err := collectDatabaseDiagnostics(tx, &result.Diagnostics); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if err := insertMetadata(tx, result); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit import transaction: %w", err)
	}
	committed = true

	if err := validateDatabase(db); err != nil {
		return result, withDiagnostics(err, result.Diagnostics)
	}
	if _, err := db.Exec("PRAGMA optimize"); err != nil {
		return result, fmt.Errorf("optimize sqlite database: %w", err)
	}
	if err := db.Close(); err != nil {
		return result, fmt.Errorf("close sqlite database: %w", err)
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		return result, fmt.Errorf("move completed sqlite database: %w", err)
	}

	return result, nil
}
