package main

import (
	"database/sql"
	"fmt"
)

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
