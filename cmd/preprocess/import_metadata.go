package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func insertMetadata(tx *sql.Tx, result stats) error {
	stmt, err := tx.Prepare(`INSERT INTO dataset_meta (key, value) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare metadata insert: %w", err)
	}
	defer stmt.Close()

	values := map[string]string{
		"schema_version":        schemaVersion,
		"generated_at":          time.Now().UTC().Format(time.RFC3339),
		"organes_count":         strconv.Itoa(result.Organes),
		"acteurs_count":         strconv.Itoa(result.Acteurs),
		"acteur_adresses_count": strconv.Itoa(result.Adresses),
		"mandats_count":         strconv.Itoa(result.Mandats),
		"mandat_organes_count":  strconv.Itoa(result.MandatOrganes),
		"scrutins_count":        strconv.Itoa(result.Scrutins),
		"groupe_votes_count":    strconv.Itoa(result.ScrutinGroupeVotes),
		"votes_count":           strconv.Itoa(result.Votes),
	}

	for key, value := range values {
		if _, err := stmt.Exec(key, value); err != nil {
			return fmt.Errorf("insert metadata %s: %w", key, err)
		}
	}

	if err := insertDerivedStats(tx); err != nil {
		return err
	}
	return insertScrutinSearch(tx)
}

func insertDerivedStats(tx *sql.Tx) error {
	if _, err := tx.Exec(`
INSERT INTO acteur_vote_stats (acteur_uid, legislature, total_votes, pour, contre, abstentions, non_votants)
SELECT
  v.acteur_uid,
  s.legislature,
  COUNT(*) AS total_votes,
  SUM(CASE WHEN v.position = 'pour' THEN 1 ELSE 0 END) AS pour,
  SUM(CASE WHEN v.position = 'contre' THEN 1 ELSE 0 END) AS contre,
  SUM(CASE WHEN v.position = 'abstention' THEN 1 ELSE 0 END) AS abstentions,
  SUM(CASE WHEN v.position IN ('non_votant', 'non_votant_volontaire') THEN 1 ELSE 0 END) AS non_votants
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
GROUP BY v.acteur_uid, s.legislature
`); err != nil {
		return fmt.Errorf("insert actor vote stats: %w", err)
	}

	if _, err := tx.Exec(`
INSERT INTO groupe_vote_stats (groupe_uid, legislature, total_scrutins, pour, contre, abstentions, non_votants)
SELECT
  v.groupe_uid,
  s.legislature,
  COUNT(DISTINCT v.scrutin_uid) AS total_scrutins,
  SUM(CASE WHEN v.position = 'pour' THEN 1 ELSE 0 END) AS pour,
  SUM(CASE WHEN v.position = 'contre' THEN 1 ELSE 0 END) AS contre,
  SUM(CASE WHEN v.position = 'abstention' THEN 1 ELSE 0 END) AS abstentions,
  SUM(CASE WHEN v.position IN ('non_votant', 'non_votant_volontaire') THEN 1 ELSE 0 END) AS non_votants
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
WHERE v.groupe_uid IS NOT NULL
GROUP BY v.groupe_uid, s.legislature
`); err != nil {
		return fmt.Errorf("insert group vote stats: %w", err)
	}

	if _, err := tx.Exec(`
INSERT INTO acteur_latest_group (acteur_uid, legislature, groupe_uid)
SELECT acteur_uid, legislature, groupe_uid
FROM (
  SELECT
    m.acteur_uid,
    m.legislature,
    o.uid AS groupe_uid,
    ROW_NUMBER() OVER (
      PARTITION BY m.acteur_uid
      ORDER BY COALESCE(m.date_debut, '') DESC, COALESCE(m.preseance, 9999), m.uid, COALESCE(o.preseance, 9999)
    ) AS rn
  FROM mandats m
  JOIN mandat_organes mo ON mo.mandat_uid = m.uid
  JOIN organes o ON o.uid = mo.organe_uid
  WHERE UPPER(COALESCE(o.code_type, '')) = 'GP'
)
WHERE rn = 1
`); err != nil {
		return fmt.Errorf("insert actor latest groups: %w", err)
	}

	if _, err := tx.Exec(`
INSERT INTO groupe_member_stats (groupe_uid, deputies_count)
SELECT mo.organe_uid, COUNT(DISTINCT m.acteur_uid) AS deputies_count
FROM mandat_organes mo
JOIN mandats m ON m.uid = mo.mandat_uid
JOIN organes o ON o.uid = mo.organe_uid
WHERE UPPER(COALESCE(o.code_type, '')) = 'GP'
GROUP BY mo.organe_uid
`); err != nil {
		return fmt.Errorf("insert group member stats: %w", err)
	}

	return nil
}

func insertScrutinSearch(tx *sql.Tx) error {
	if _, err := tx.Exec(`
INSERT INTO scrutin_search (uid, document)
SELECT
  s.uid,
  printf(
    '%s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s',
    s.numero,
    COALESCE(s.titre, ''),
    COALESCE(s.objet_libelle, ''),
    COALESCE(s.linked_text_num, ''),
    COALESCE(s.linked_text_kind, ''),
    COALESCE(s.linked_text_url, ''),
    COALESCE(s.linked_text_pdf_url, ''),
    COALESCE(s.linked_dossier_ref, ''),
    COALESCE(s.linked_dossier_libelle, ''),
    COALESCE(s.linked_amendement_num, ''),
    COALESCE(s.linked_amendement_text_num, ''),
    COALESCE(s.linked_amendement_organe, ''),
    COALESCE(s.demandeur_texte, ''),
    COALESCE(s.sort_code, ''),
    COALESCE(s.sort_libelle, ''),
    COALESCE(s.libelle_type_vote, ''),
    COALESCE(o.libelle_abrege, o.libelle, s.organe_uid, '')
  )
FROM scrutins s
LEFT JOIN organes o ON o.uid = s.organe_uid
`); err != nil {
		return fmt.Errorf("insert scrutin search index: %w", err)
	}

	return nil
}
