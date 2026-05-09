package store

import (
	"context"
	"fmt"
)

func (s *Store) ScrutinDetailPage(ctx context.Context, uid string) (ScrutinDetailPage, error) {
	var page ScrutinDetailPage
	var scrutin ScrutinDetailData
	if err := s.db.QueryRowContext(ctx, `
SELECT
  s.uid,
  s.numero,
  s.legislature,
  COALESCE(s.organe_uid, ''),
  COALESCE(o.libelle_abrege, o.libelle, s.organe_uid, ''),
  COALESCE(s.session_ref, ''),
  COALESCE(s.seance_ref, ''),
  s.date_scrutin,
  COALESCE(s.quantieme_jour_seance, 0),
  COALESCE(s.code_type_vote, ''),
  COALESCE(s.libelle_type_vote, ''),
  COALESCE(s.type_majorite, ''),
  COALESCE(s.sort_code, ''),
  COALESCE(s.sort_libelle, ''),
  COALESCE(s.titre, ''),
  COALESCE(s.linked_text_num, ''),
  COALESCE(s.linked_text_kind, ''),
  COALESCE(s.linked_text_url, ''),
  COALESCE(s.linked_text_pdf_url, ''),
  COALESCE(s.linked_dossier_ref, ''),
  COALESCE(s.linked_dossier_libelle, ''),
  COALESCE(s.linked_amendement_num, ''),
  COALESCE(s.linked_amendement_text_num, ''),
  COALESCE(s.linked_amendement_organe, ''),
  COALESCE(s.linked_amendement_url, ''),
  COALESCE(s.linked_reference_source, ''),
  COALESCE(s.demandeur_texte, ''),
  COALESCE(s.objet_libelle, ''),
  COALESCE(s.mode_publication_votes, ''),
  COALESCE(s.nombre_votants, 0),
  COALESCE(s.suffrages_exprimes, 0),
  COALESCE(s.suffrages_requis, 0),
  COALESCE(s.non_votants, 0),
  COALESCE(s.pour, 0),
  COALESCE(s.contre, 0),
  COALESCE(s.abstentions, 0),
  COALESCE(s.non_votants_volontaires, 0),
  COALESCE(s.source_file, '')
FROM scrutins s
LEFT JOIN organes o ON o.uid = s.organe_uid
WHERE s.uid = ?
`, uid).Scan(
		&scrutin.UID,
		&scrutin.Numero,
		&scrutin.Legislature,
		&scrutin.OrganeUID,
		&scrutin.Organe,
		&scrutin.SessionRef,
		&scrutin.SeanceRef,
		&scrutin.Date,
		&scrutin.QuantiemeJourSeance,
		&scrutin.CodeTypeVote,
		&scrutin.TypeVote,
		&scrutin.TypeMajorite,
		&scrutin.SortCode,
		&scrutin.SortLibelle,
		&scrutin.Titre,
		&scrutin.LinkedTextNum,
		&scrutin.LinkedTextKind,
		&scrutin.LinkedTextURL,
		&scrutin.LinkedTextPDFURL,
		&scrutin.LinkedDossierRef,
		&scrutin.LinkedDossierLibelle,
		&scrutin.LinkedAmendementNum,
		&scrutin.LinkedAmendementTextNum,
		&scrutin.LinkedAmendementOrgane,
		&scrutin.LinkedAmendementURL,
		&scrutin.LinkedReferenceSource,
		&scrutin.Demandeur,
		&scrutin.Objet,
		&scrutin.ModePublicationVotes,
		&scrutin.NombreVotants,
		&scrutin.SuffragesExprimes,
		&scrutin.SuffragesRequis,
		&scrutin.NonVotants,
		&scrutin.Pour,
		&scrutin.Contre,
		&scrutin.Abstentions,
		&scrutin.NonVotantsVolontaires,
		&scrutin.SourceFile,
	); err != nil {
		return ScrutinDetailPage{}, fmt.Errorf("query scrutin: %w", err)
	}
	page.Scrutin = scrutin

	rows, err := s.db.QueryContext(ctx, `
SELECT
  sgv.groupe_uid,
  COALESCE(o.libelle_abrege, o.libelle, sgv.groupe_uid, ''),
  COALESCE(sgv.nombre_membres_groupe, 0),
  COALESCE(sgv.position_majoritaire, ''),
  COALESCE(sgv.non_votants, 0),
  COALESCE(sgv.pour, 0),
  COALESCE(sgv.contre, 0),
  COALESCE(sgv.abstentions, 0),
  COALESCE(sgv.non_votants_volontaires, 0)
FROM scrutin_groupe_votes sgv
LEFT JOIN organes o ON o.uid = sgv.groupe_uid
WHERE sgv.scrutin_uid = ?
ORDER BY COALESCE(o.preseance, 9999), COALESCE(o.libelle, sgv.groupe_uid)
`, uid)
	if err != nil {
		return ScrutinDetailPage{}, fmt.Errorf("query scrutin group votes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var groupVote ScrutinGroupVote
		if err := rows.Scan(
			&groupVote.GroupeUID,
			&groupVote.Groupe,
			&groupVote.NombreMembres,
			&groupVote.PositionMajoritaire,
			&groupVote.NonVotants,
			&groupVote.Pour,
			&groupVote.Contre,
			&groupVote.Abstentions,
			&groupVote.NonVotantsVolontaires,
		); err != nil {
			return ScrutinDetailPage{}, fmt.Errorf("scan scrutin group vote: %w", err)
		}
		page.GroupVotes = append(page.GroupVotes, groupVote)
	}
	if err := rows.Err(); err != nil {
		return ScrutinDetailPage{}, fmt.Errorf("iterate scrutin group votes: %w", err)
	}

	return page, nil
}
