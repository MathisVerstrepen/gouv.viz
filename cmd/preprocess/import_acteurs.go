package main

import (
	"database/sql"
	"fmt"
)

func importActeurs(tx *sql.Tx, dir string, result *stats) error {
	acteurStmt, err := tx.Prepare(`
INSERT INTO acteurs (
  uid, civilite, prenom, nom, alpha, date_naissance, ville_naissance,
  dep_naissance, pays_naissance, date_deces, profession, uri_hatvp,
  source_file, source_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO UPDATE SET
  civilite = excluded.civilite,
  prenom = excluded.prenom,
  nom = excluded.nom,
  alpha = excluded.alpha,
  date_naissance = excluded.date_naissance,
  ville_naissance = excluded.ville_naissance,
  dep_naissance = excluded.dep_naissance,
  pays_naissance = excluded.pays_naissance,
  date_deces = excluded.date_deces,
  profession = excluded.profession,
  uri_hatvp = excluded.uri_hatvp,
  source_file = excluded.source_file,
  source_hash = excluded.source_hash
`)
	if err != nil {
		return fmt.Errorf("prepare acteur insert: %w", err)
	}
	defer acteurStmt.Close()

	adresseStmt, err := tx.Prepare(`
INSERT INTO acteur_adresses (uid, acteur_uid, type_code, type_libelle, poids, adresse_rattachement, valeur)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO UPDATE SET
  acteur_uid = excluded.acteur_uid,
  type_code = excluded.type_code,
  type_libelle = excluded.type_libelle,
  poids = excluded.poids,
  adresse_rattachement = excluded.adresse_rattachement,
  valeur = excluded.valeur
`)
	if err != nil {
		return fmt.Errorf("prepare acteur address insert: %w", err)
	}
	defer adresseStmt.Close()

	mandatStmt, err := tx.Prepare(`
INSERT INTO mandats (
  uid, acteur_uid, legislature, type_organe, date_debut, date_fin,
  date_publication, preseance, nomin_principale, code_qualite,
  lib_qualite, lib_qualite_sex
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO UPDATE SET
  acteur_uid = excluded.acteur_uid,
  legislature = excluded.legislature,
  type_organe = excluded.type_organe,
  date_debut = excluded.date_debut,
  date_fin = excluded.date_fin,
  date_publication = excluded.date_publication,
  preseance = excluded.preseance,
  nomin_principale = excluded.nomin_principale,
  code_qualite = excluded.code_qualite,
  lib_qualite = excluded.lib_qualite,
  lib_qualite_sex = excluded.lib_qualite_sex
`)
	if err != nil {
		return fmt.Errorf("prepare mandat insert: %w", err)
	}
	defer mandatStmt.Close()

	mandatOrganeStmt, err := tx.Prepare(`
INSERT OR IGNORE INTO mandat_organes (mandat_uid, organe_uid)
VALUES (?, ?)
`)
	if err != nil {
		return fmt.Errorf("prepare mandat organe insert: %w", err)
	}
	defer mandatOrganeStmt.Close()

	return walkRawJSON(dir, func(file rawFile) error {
		acteur := objectAt(file.Root, "acteur")
		uid := stringAt(acteur, "uid")
		if uid == "" {
			return fmt.Errorf("%s: missing acteur.uid", file.SourcePath)
		}

		_, err := acteurStmt.Exec(
			uid,
			nullString(stringAt(acteur, "etatCivil", "ident", "civ")),
			nullString(stringAt(acteur, "etatCivil", "ident", "prenom")),
			nullString(stringAt(acteur, "etatCivil", "ident", "nom")),
			nullString(stringAt(acteur, "etatCivil", "ident", "alpha")),
			nullString(stringAt(acteur, "etatCivil", "infoNaissance", "dateNais")),
			nullString(stringAt(acteur, "etatCivil", "infoNaissance", "villeNais")),
			nullString(stringAt(acteur, "etatCivil", "infoNaissance", "depNais")),
			nullString(stringAt(acteur, "etatCivil", "infoNaissance", "paysNais")),
			nullString(stringAt(acteur, "etatCivil", "dateDeces")),
			nullString(stringAt(acteur, "profession", "libelleCourant")),
			nullString(stringAt(acteur, "uri_hatvp")),
			file.SourcePath,
			file.SourceHash,
		)
		if err != nil {
			return fmt.Errorf("%s: insert acteur %s: %w", file.SourcePath, uid, err)
		}
		result.Acteurs++

		if err := insertActeurAdresses(adresseStmt, acteur, uid, result); err != nil {
			return fmt.Errorf("%s: %w", file.SourcePath, err)
		}
		if err := insertMandats(mandatStmt, mandatOrganeStmt, acteur, uid, result); err != nil {
			return fmt.Errorf("%s: %w", file.SourcePath, err)
		}

		return nil
	})
}

func insertActeurAdresses(stmt *sql.Stmt, acteur map[string]any, acteurUID string, result *stats) error {
	for index, item := range items(valueAt(acteur, "adresses", "adresse")) {
		adresse := asObject(item)
		uid := stringAt(adresse, "uid")
		if uid == "" {
			uid = fmt.Sprintf("%s:adresse:%d", acteurUID, index+1)
		}

		_, err := stmt.Exec(
			uid,
			acteurUID,
			nullString(stringAt(adresse, "type")),
			nullString(stringAt(adresse, "typeLibelle")),
			nullInt(intAt(adresse, "poids")),
			nullString(stringAt(adresse, "adresseDeRattachement")),
			nullString(stringAt(adresse, "valElec")),
		)
		if err != nil {
			return fmt.Errorf("insert adresse %s: %w", uid, err)
		}
		result.Adresses++
	}
	return nil
}

func insertMandats(mandatStmt, mandatOrganeStmt *sql.Stmt, acteur map[string]any, acteurUID string, result *stats) error {
	for _, item := range items(valueAt(acteur, "mandats", "mandat")) {
		mandat := asObject(item)
		uid := stringAt(mandat, "uid")
		if uid == "" {
			continue
		}
		mandatActeurUID := stringAt(mandat, "acteurRef")
		if mandatActeurUID == "" {
			mandatActeurUID = acteurUID
		}

		_, err := mandatStmt.Exec(
			uid,
			mandatActeurUID,
			nullInt(intAt(mandat, "legislature")),
			nullString(stringAt(mandat, "typeOrgane")),
			nullString(stringAt(mandat, "dateDebut")),
			nullString(stringAt(mandat, "dateFin")),
			nullString(stringAt(mandat, "datePublication")),
			nullInt(intAt(mandat, "preseance")),
			nullInt(intAt(mandat, "nominPrincipale")),
			nullString(stringAt(mandat, "infosQualite", "codeQualite")),
			nullString(stringAt(mandat, "infosQualite", "libQualite")),
			nullString(stringAt(mandat, "infosQualite", "libQualiteSex")),
		)
		if err != nil {
			return fmt.Errorf("insert mandat %s: %w", uid, err)
		}
		result.Mandats++

		for _, organeUID := range stringsFromValue(valueAt(mandat, "organes", "organeRef")) {
			if organeUID == "" {
				continue
			}
			if _, err := mandatOrganeStmt.Exec(uid, organeUID); err != nil {
				return fmt.Errorf("insert mandat organe %s/%s: %w", uid, organeUID, err)
			}
			result.MandatOrganes++
		}
	}
	return nil
}
