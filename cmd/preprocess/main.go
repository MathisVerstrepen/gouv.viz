package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schemaVersion = "1"

type stats struct {
	Organes            int
	Acteurs            int
	Adresses           int
	Mandats            int
	MandatOrganes      int
	Scrutins           int
	ScrutinGroupeVotes int
	Votes              int
}

type rawFile struct {
	Path       string
	SourcePath string
	SourceHash string
	Root       map[string]any
}

func main() {
	rawDir := flag.String("raw", filepath.Join("data", "raw"), "raw Assemblee nationale data directory")
	outPath := flag.String("out", filepath.Join("data", "processed", "gouv-viz.sqlite"), "output SQLite database path")
	flag.Parse()

	result, err := buildDatabase(*rawDir, *outPath)
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
}

func buildDatabase(rawDir, outPath string) (stats, error) {
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

	if err := importOrganes(tx, filepath.Join(rawDir, "organe"), &result); err != nil {
		return result, err
	}
	if err := insertSyntheticOrganes(tx, &result); err != nil {
		return result, err
	}
	if err := importActeurs(tx, filepath.Join(rawDir, "acteur"), &result); err != nil {
		return result, err
	}
	if err := importScrutins(tx, filepath.Join(rawDir, "scrutins-publics"), &result); err != nil {
		return result, err
	}
	if err := insertMetadata(tx, result); err != nil {
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit import transaction: %w", err)
	}
	committed = true

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
