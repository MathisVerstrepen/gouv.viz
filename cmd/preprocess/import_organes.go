package main

import (
	"database/sql"
	"fmt"
)

func importOrganes(tx *sql.Tx, dir string, result *stats) error {
	stmt, err := tx.Prepare(`
INSERT INTO organes (
  uid, code_type, libelle, libelle_abrege, libelle_abrev, libelle_edition,
  legislature, chambre, regime, organe_parent_uid, position_politique,
  couleur_associee, preseance, date_debut, date_agrement, date_fin,
  secretaire_01_uid, secretaire_02_uid, source_file, source_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO UPDATE SET
  code_type = excluded.code_type,
  libelle = excluded.libelle,
  libelle_abrege = excluded.libelle_abrege,
  libelle_abrev = excluded.libelle_abrev,
  libelle_edition = excluded.libelle_edition,
  legislature = excluded.legislature,
  chambre = excluded.chambre,
  regime = excluded.regime,
  organe_parent_uid = excluded.organe_parent_uid,
  position_politique = excluded.position_politique,
  couleur_associee = excluded.couleur_associee,
  preseance = excluded.preseance,
  date_debut = excluded.date_debut,
  date_agrement = excluded.date_agrement,
  date_fin = excluded.date_fin,
  secretaire_01_uid = excluded.secretaire_01_uid,
  secretaire_02_uid = excluded.secretaire_02_uid,
  source_file = excluded.source_file,
  source_hash = excluded.source_hash
`)
	if err != nil {
		return fmt.Errorf("prepare organe insert: %w", err)
	}
	defer stmt.Close()

	return walkRawJSON(dir, func(file rawFile) error {
		organe := objectAt(file.Root, "organe")
		uid := stringAt(organe, "uid")
		if uid == "" {
			return fmt.Errorf("%s: missing organe.uid", file.SourcePath)
		}

		_, err := stmt.Exec(
			uid,
			nullString(stringAt(organe, "codeType")),
			nullString(stringAt(organe, "libelle")),
			nullString(stringAt(organe, "libelleAbrege")),
			nullString(stringAt(organe, "libelleAbrev")),
			nullString(stringAt(organe, "libelleEdition")),
			nullInt(intAt(organe, "legislature")),
			nullString(stringAt(organe, "chambre")),
			nullString(stringAt(organe, "regime")),
			nullString(stringAt(organe, "organeParent")),
			nullString(stringAt(organe, "positionPolitique")),
			nullString(stringAt(organe, "couleurAssociee")),
			nullInt(intAt(organe, "preseance")),
			nullString(stringAt(organe, "viMoDe", "dateDebut")),
			nullString(stringAt(organe, "viMoDe", "dateAgrement")),
			nullString(stringAt(organe, "viMoDe", "dateFin")),
			nullString(stringAt(organe, "secretariat", "secretaire01")),
			nullString(stringAt(organe, "secretariat", "secretaire02")),
			file.SourcePath,
			file.SourceHash,
		)
		if err != nil {
			return fmt.Errorf("%s: insert organe %s: %w", file.SourcePath, uid, err)
		}

		result.Organes++
		return nil
	})
}

func insertSyntheticOrganes(tx *sql.Tx, result *stats) error {
	// Scrutin public files use PO0 for non-inscrits, but this pseudo-group is not
	// present in the raw organe dataset.
	res, err := tx.Exec(`
INSERT INTO organes (
  uid, code_type, libelle, libelle_abrege, libelle_abrev, libelle_edition,
  source_file, source_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO NOTHING
`, "PO0", "GP", "Non inscrits", "NI", "NI", "Non inscrits", "generated", "generated")
	if err != nil {
		return fmt.Errorf("insert synthetic organes: %w", err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("count synthetic organes: %w", err)
	}
	result.Organes += int(inserted)
	return nil
}
