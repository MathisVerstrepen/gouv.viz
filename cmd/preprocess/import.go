package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
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

func importScrutins(tx *sql.Tx, dir string, result *stats, amendementResolver amendementLinkResolver, dossierResolver dossierLinkResolver) error {
	scrutinStmt, err := tx.Prepare(`
INSERT INTO scrutins (
  uid, numero, legislature, organe_uid, session_ref, seance_ref,
  date_scrutin, quantieme_jour_seance, code_type_vote, libelle_type_vote,
  type_majorite, sort_code, sort_libelle, titre, linked_text_num,
  linked_text_kind, linked_text_url, linked_text_pdf_url, linked_dossier_ref, linked_dossier_libelle,
  linked_amendement_num, linked_amendement_text_num, linked_amendement_organe,
  linked_amendement_url, linked_reference_source, demandeur_texte, objet_libelle,
  mode_publication_votes, nombre_votants, suffrages_exprimes,
  suffrages_requis, non_votants, pour, contre, abstentions,
  non_votants_volontaires, source_file, source_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(uid) DO UPDATE SET
  numero = excluded.numero,
  legislature = excluded.legislature,
  organe_uid = excluded.organe_uid,
  session_ref = excluded.session_ref,
  seance_ref = excluded.seance_ref,
  date_scrutin = excluded.date_scrutin,
  quantieme_jour_seance = excluded.quantieme_jour_seance,
  code_type_vote = excluded.code_type_vote,
  libelle_type_vote = excluded.libelle_type_vote,
  type_majorite = excluded.type_majorite,
  sort_code = excluded.sort_code,
  sort_libelle = excluded.sort_libelle,
  titre = excluded.titre,
  linked_text_num = excluded.linked_text_num,
  linked_text_kind = excluded.linked_text_kind,
  linked_text_url = excluded.linked_text_url,
  linked_text_pdf_url = excluded.linked_text_pdf_url,
  linked_dossier_ref = excluded.linked_dossier_ref,
  linked_dossier_libelle = excluded.linked_dossier_libelle,
  linked_amendement_num = excluded.linked_amendement_num,
  linked_amendement_text_num = excluded.linked_amendement_text_num,
  linked_amendement_organe = excluded.linked_amendement_organe,
  linked_amendement_url = excluded.linked_amendement_url,
  linked_reference_source = excluded.linked_reference_source,
  demandeur_texte = excluded.demandeur_texte,
  objet_libelle = excluded.objet_libelle,
  mode_publication_votes = excluded.mode_publication_votes,
  nombre_votants = excluded.nombre_votants,
  suffrages_exprimes = excluded.suffrages_exprimes,
  suffrages_requis = excluded.suffrages_requis,
  non_votants = excluded.non_votants,
  pour = excluded.pour,
  contre = excluded.contre,
  abstentions = excluded.abstentions,
  non_votants_volontaires = excluded.non_votants_volontaires,
  source_file = excluded.source_file,
  source_hash = excluded.source_hash
`)
	if err != nil {
		return fmt.Errorf("prepare scrutin insert: %w", err)
	}
	defer scrutinStmt.Close()

	groupeVoteStmt, err := tx.Prepare(`
INSERT INTO scrutin_groupe_votes (
  scrutin_uid, groupe_uid, nombre_membres_groupe, position_majoritaire,
  non_votants, pour, contre, abstentions, non_votants_volontaires
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(scrutin_uid, groupe_uid) DO UPDATE SET
  nombre_membres_groupe = excluded.nombre_membres_groupe,
  position_majoritaire = excluded.position_majoritaire,
  non_votants = excluded.non_votants,
  pour = excluded.pour,
  contre = excluded.contre,
  abstentions = excluded.abstentions,
  non_votants_volontaires = excluded.non_votants_volontaires
`)
	if err != nil {
		return fmt.Errorf("prepare scrutin group vote insert: %w", err)
	}
	defer groupeVoteStmt.Close()

	voteStmt, err := tx.Prepare(`
INSERT INTO votes (scrutin_uid, acteur_uid, mandat_uid, groupe_uid, position, par_delegation, num_place)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(scrutin_uid, acteur_uid) DO UPDATE SET
  mandat_uid = excluded.mandat_uid,
  groupe_uid = excluded.groupe_uid,
  position = excluded.position,
  par_delegation = excluded.par_delegation,
  num_place = excluded.num_place
`)
	if err != nil {
		return fmt.Errorf("prepare vote insert: %w", err)
	}
	defer voteStmt.Close()

	return walkRawJSON(dir, func(file rawFile) error {
		scrutin := objectAt(file.Root, "scrutin")
		uid := stringAt(scrutin, "uid")
		if uid == "" {
			return fmt.Errorf("%s: missing scrutin.uid", file.SourcePath)
		}

		title := stringAt(scrutin, "titre")
		object := stringAt(scrutin, "objet", "libelle")
		dossier := objectAt(scrutin, "objet", "dossierLegislatif")
		dossierRef := stringAt(dossier, "dossierRef")
		linkedReference := extractScrutinLinkedReference(title, object)
		var linkedDossier officialDossierReference
		if dossierResolver != nil {
			hadLinkedTextNum := linkedReference.TextNum != ""
			linkedDossier = dossierResolver.Resolve(dossierRef, title, object, stringAt(scrutin, "sort", "code"))
			if dossierRef == "" {
				dossierRef = linkedDossier.DossierRef
			}
			if !hadLinkedTextNum && linkedDossier.Text.Num != "" {
				linkedReference.TextNum = linkedDossier.Text.Num
				linkedReference.TextKind = linkedDossier.Text.Kind
			}
		}
		var linkedAmendement officialAmendementReference
		if amendementResolver != nil && linkedReference.AmendementNum != "" && dossierRef != "" {
			legislature, _ := intAt(scrutin, "legislature")
			linkedAmendement, err = amendementResolver.Resolve(legislature, dossierRef, linkedReference.AmendementNum, stringAt(scrutin, "organeRef"), stringAt(scrutin, "seanceRef"))
			if err != nil {
				return fmt.Errorf("%s: resolve amendement link for scrutin %s: %w", file.SourcePath, uid, err)
			}
		}

		_, err := scrutinStmt.Exec(
			uid,
			nullInt(intAt(scrutin, "numero")),
			nullInt(intAt(scrutin, "legislature")),
			nullString(stringAt(scrutin, "organeRef")),
			nullString(stringAt(scrutin, "sessionRef")),
			nullString(stringAt(scrutin, "seanceRef")),
			nullString(stringAt(scrutin, "dateScrutin")),
			nullInt(intAt(scrutin, "quantiemeJourSeance")),
			nullString(stringAt(scrutin, "typeVote", "codeTypeVote")),
			nullString(stringAt(scrutin, "typeVote", "libelleTypeVote")),
			nullString(stringAt(scrutin, "typeVote", "typeMajorite")),
			nullString(stringAt(scrutin, "sort", "code")),
			nullString(stringAt(scrutin, "sort", "libelle")),
			nullString(title),
			nullString(linkedReference.TextNum),
			nullString(linkedReference.TextKind),
			nullString(linkedDossier.Text.URL),
			nullString(linkedDossier.Text.PDFURL),
			nullString(dossierRef),
			nullString(firstNonEmpty(stringAt(dossier, "libelle"), linkedDossier.DossierLibelle)),
			nullString(linkedReference.AmendementNum),
			nullString(linkedAmendement.TextNum),
			nullString(linkedAmendement.Organe),
			nullString(linkedAmendement.URL),
			nullString(linkedReference.Source),
			nullString(stringAt(scrutin, "demandeur", "texte")),
			nullString(object),
			nullString(stringAt(scrutin, "modePublicationDesVotes")),
			nullInt(intAt(scrutin, "syntheseVote", "nombreVotants")),
			nullInt(intAt(scrutin, "syntheseVote", "suffragesExprimes")),
			nullInt(intAt(scrutin, "syntheseVote", "nbrSuffragesRequis")),
			nullInt(intAt(scrutin, "syntheseVote", "decompte", "nonVotants")),
			nullInt(intAt(scrutin, "syntheseVote", "decompte", "pour")),
			nullInt(intAt(scrutin, "syntheseVote", "decompte", "contre")),
			nullInt(intAt(scrutin, "syntheseVote", "decompte", "abstentions")),
			nullInt(intAt(scrutin, "syntheseVote", "decompte", "nonVotantsVolontaires")),
			file.SourcePath,
			file.SourceHash,
		)
		if err != nil {
			return fmt.Errorf("%s: insert scrutin %s: %w", file.SourcePath, uid, err)
		}
		result.Scrutins++

		if err := insertScrutinGroupes(groupeVoteStmt, voteStmt, scrutin, uid, result); err != nil {
			return fmt.Errorf("%s: %w", file.SourcePath, err)
		}

		return nil
	})
}

func insertScrutinGroupes(groupeVoteStmt, voteStmt *sql.Stmt, scrutin map[string]any, scrutinUID string, result *stats) error {
	for _, item := range items(valueAt(scrutin, "ventilationVotes", "organe", "groupes", "groupe")) {
		groupe := asObject(item)
		groupeUID := stringAt(groupe, "organeRef")
		if groupeUID == "" {
			continue
		}

		vote := objectAt(groupe, "vote")
		decompte := objectAt(vote, "decompteVoix")
		_, err := groupeVoteStmt.Exec(
			scrutinUID,
			groupeUID,
			nullInt(intAt(groupe, "nombreMembresGroupe")),
			nullString(stringAt(vote, "positionMajoritaire")),
			nullInt(intAt(decompte, "nonVotants")),
			nullInt(intAt(decompte, "pour")),
			nullInt(intAt(decompte, "contre")),
			nullInt(intAt(decompte, "abstentions")),
			nullInt(intAt(decompte, "nonVotantsVolontaires")),
		)
		if err != nil {
			return fmt.Errorf("insert scrutin group vote %s/%s: %w", scrutinUID, groupeUID, err)
		}
		result.ScrutinGroupeVotes++

		nominatif := objectAt(vote, "decompteNominatif")
		buckets := map[string]string{
			"pours":                 "pour",
			"contres":               "contre",
			"abstentions":           "abstention",
			"nonVotants":            "non_votant",
			"nonVotantsVolontaires": "non_votant_volontaire",
		}
		for rawBucket, position := range buckets {
			if err := insertVotes(voteStmt, valueAt(nominatif, rawBucket), scrutinUID, groupeUID, position, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertVotes(stmt *sql.Stmt, bucket any, scrutinUID, groupeUID, position string, result *stats) error {
	votants := items(valueAt(asObject(bucket), "votant"))
	if len(votants) == 0 {
		votants = items(bucket)
	}

	for _, item := range votants {
		votant := asObject(item)
		acteurUID := stringAt(votant, "acteurRef")
		if acteurUID == "" {
			continue
		}

		_, err := stmt.Exec(
			scrutinUID,
			acteurUID,
			nullString(stringAt(votant, "mandatRef")),
			groupeUID,
			position,
			nullInt(boolIntAt(votant, "parDelegation")),
			nullString(stringAt(votant, "numPlace")),
		)
		if err != nil {
			return fmt.Errorf("insert vote %s/%s: %w", scrutinUID, acteurUID, err)
		}
		result.Votes++
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

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
