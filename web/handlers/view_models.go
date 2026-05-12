package handlers

import (
	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func homeView(page store.HomePage) components.HomePage {
	return components.HomePage{
		TotalScrutins:    page.TotalScrutins,
		FirstScrutinDate: page.FirstScrutinDate,
		LastScrutinDate:  page.LastScrutinDate,
		Scrutins:         scrutinListItemViews(page.Scrutins),
	}
}

func scrutinsView(page store.ScrutinsPage) components.ScrutinsPage {
	return components.ScrutinsPage{
		Query: components.ScrutinsQuery{
			Search:      page.Query.Search,
			Sort:        page.Query.Sort,
			Page:        page.Query.Page,
			PerPage:     page.Query.PerPage,
			Legislature: page.Query.Legislature,
			Result:      page.Query.Result,
			VoteType:    page.Query.VoteType,
			Organe:      page.Query.Organe,
			DateFrom:    page.Query.DateFrom,
			DateTo:      page.Query.DateTo,
			CloseVotes:  page.Query.CloseVotes,
		},
		DefaultSort:   page.DefaultSort,
		SortOptions:   scrutinSortOptionViews(page.SortOptions),
		FilterOptions: scrutinFilterOptionsView(page.FilterOptions),
		Scrutins:      scrutinListItemViews(page.Scrutins),
		TotalResults:  page.TotalResults,
		TotalPages:    page.TotalPages,
		StartItem:     page.StartItem,
		EndItem:       page.EndItem,
	}
}

func scrutinDetailView(page store.ScrutinDetailPage) components.ScrutinDetailPage {
	return components.ScrutinDetailPage{
		Scrutin: components.ScrutinDetailData{
			UID:                     page.Scrutin.UID,
			Numero:                  page.Scrutin.Numero,
			Legislature:             page.Scrutin.Legislature,
			OrganeUID:               page.Scrutin.OrganeUID,
			Organe:                  page.Scrutin.Organe,
			SessionRef:              page.Scrutin.SessionRef,
			SeanceRef:               page.Scrutin.SeanceRef,
			Date:                    page.Scrutin.Date,
			QuantiemeJourSeance:     page.Scrutin.QuantiemeJourSeance,
			CodeTypeVote:            page.Scrutin.CodeTypeVote,
			TypeVote:                page.Scrutin.TypeVote,
			TypeMajorite:            page.Scrutin.TypeMajorite,
			SortCode:                page.Scrutin.SortCode,
			SortLibelle:             page.Scrutin.SortLibelle,
			Titre:                   page.Scrutin.Titre,
			LinkedTextNum:           page.Scrutin.LinkedTextNum,
			LinkedTextKind:          page.Scrutin.LinkedTextKind,
			LinkedTextURL:           page.Scrutin.LinkedTextURL,
			LinkedTextPDFURL:        page.Scrutin.LinkedTextPDFURL,
			LinkedDossierRef:        page.Scrutin.LinkedDossierRef,
			LinkedDossierLibelle:    page.Scrutin.LinkedDossierLibelle,
			LinkedAmendementNum:     page.Scrutin.LinkedAmendementNum,
			LinkedAmendementTextNum: page.Scrutin.LinkedAmendementTextNum,
			LinkedAmendementOrgane:  page.Scrutin.LinkedAmendementOrgane,
			LinkedAmendementURL:     page.Scrutin.LinkedAmendementURL,
			LinkedReferenceSource:   page.Scrutin.LinkedReferenceSource,
			Demandeur:               page.Scrutin.Demandeur,
			Objet:                   page.Scrutin.Objet,
			ModePublicationVotes:    page.Scrutin.ModePublicationVotes,
			NombreVotants:           page.Scrutin.NombreVotants,
			SuffragesExprimes:       page.Scrutin.SuffragesExprimes,
			SuffragesRequis:         page.Scrutin.SuffragesRequis,
			NonVotants:              page.Scrutin.NonVotants,
			Pour:                    page.Scrutin.Pour,
			Contre:                  page.Scrutin.Contre,
			Abstentions:             page.Scrutin.Abstentions,
			NonVotantsVolontaires:   page.Scrutin.NonVotantsVolontaires,
			SourceFile:              page.Scrutin.SourceFile,
		},
		GroupVotes:      scrutinGroupVoteViews(page.GroupVotes),
		IndividualVotes: scrutinIndividualVoteViews(page.IndividualVotes),
	}
}

func deputyDetailView(page store.DeputyDetailPage) components.DeputyDetailPage {
	return components.DeputyDetailPage{
		Deputy: components.DeputyDetailData{
			UID:            page.Deputy.UID,
			Civilite:       page.Deputy.Civilite,
			Prenom:         page.Deputy.Prenom,
			Nom:            page.Deputy.Nom,
			Alpha:          page.Deputy.Alpha,
			DisplayName:    page.Deputy.DisplayName,
			DateNaissance:  page.Deputy.DateNaissance,
			VilleNaissance: page.Deputy.VilleNaissance,
			DepNaissance:   page.Deputy.DepNaissance,
			PaysNaissance:  page.Deputy.PaysNaissance,
			DateDeces:      page.Deputy.DateDeces,
			Profession:     page.Deputy.Profession,
			URIHATVP:       page.Deputy.URIHATVP,
			SourceFile:     page.Deputy.SourceFile,
		},
		Query: components.DeputyDetailQuery{
			VotesPage:     page.Query.VotesPage,
			VotesPerPage:  page.Query.VotesPerPage,
			VotesSearch:   page.Query.VotesSearch,
			VotesSort:     page.Query.VotesSort,
			VotesPosition: page.Query.VotesPosition,
		},
		Mandats:           deputyMandatViews(page.Mandats),
		Stats:             deputyVoteStatViews(page.Stats),
		Votes:             deputyVoteViews(page.Votes),
		VotesTotalResults: page.VotesTotalResults,
		VotesTotalPages:   page.VotesTotalPages,
		VotesStartItem:    page.VotesStartItem,
		VotesEndItem:      page.VotesEndItem,
	}
}

func scrutinListItemViews(items []store.ScrutinListItem) []components.ScrutinListItem {
	views := make([]components.ScrutinListItem, 0, len(items))
	for _, item := range items {
		views = append(views, components.ScrutinListItem{
			UID:           item.UID,
			Numero:        item.Numero,
			Date:          item.Date,
			Titre:         item.Titre,
			SortCode:      item.SortCode,
			TypeVote:      item.TypeVote,
			Organe:        item.Organe,
			Pour:          item.Pour,
			Contre:        item.Contre,
			Abstentions:   item.Abstentions,
			NombreVotants: item.NombreVotants,
		})
	}
	return views
}

func scrutinSortOptionViews(options []store.ScrutinSortOption) []components.ScrutinSortOption {
	views := make([]components.ScrutinSortOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.ScrutinSortOption{
			Value: option.Value,
			Label: option.Label,
		})
	}
	return views
}

func scrutinFilterOptionsView(options store.ScrutinFilterOptions) components.ScrutinFilterOptions {
	return components.ScrutinFilterOptions{
		Legislatures: scrutinFilterOptionViews(options.Legislatures),
		Results:      scrutinFilterOptionViews(options.Results),
		VoteTypes:    scrutinFilterOptionViews(options.VoteTypes),
		Organes:      scrutinFilterOptionViews(options.Organes),
	}
}

func scrutinFilterOptionViews(options []store.ScrutinFilterOption) []components.ScrutinFilterOption {
	views := make([]components.ScrutinFilterOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.ScrutinFilterOption{
			Value: option.Value,
			Label: option.Label,
		})
	}
	return views
}

func scrutinGroupVoteViews(groupVotes []store.ScrutinGroupVote) []components.ScrutinGroupVote {
	views := make([]components.ScrutinGroupVote, 0, len(groupVotes))
	for _, groupVote := range groupVotes {
		views = append(views, components.ScrutinGroupVote{
			GroupeUID:             groupVote.GroupeUID,
			Groupe:                groupVote.Groupe,
			NombreMembres:         groupVote.NombreMembres,
			PositionMajoritaire:   groupVote.PositionMajoritaire,
			NonVotants:            groupVote.NonVotants,
			Pour:                  groupVote.Pour,
			Contre:                groupVote.Contre,
			Abstentions:           groupVote.Abstentions,
			NonVotantsVolontaires: groupVote.NonVotantsVolontaires,
		})
	}
	return views
}

func scrutinIndividualVoteViews(votes []store.ScrutinIndividualVote) []components.ScrutinIndividualVote {
	views := make([]components.ScrutinIndividualVote, 0, len(votes))
	for _, vote := range votes {
		views = append(views, components.ScrutinIndividualVote{
			GroupeUID:     vote.GroupeUID,
			Groupe:        vote.Groupe,
			ActeurUID:     vote.ActeurUID,
			Depute:        vote.Depute,
			Alpha:         vote.Alpha,
			MandatUID:     vote.MandatUID,
			Position:      vote.Position,
			ParDelegation: vote.ParDelegation,
			NumPlace:      vote.NumPlace,
		})
	}
	return views
}

func deputyMandatViews(mandats []store.DeputyMandat) []components.DeputyMandat {
	views := make([]components.DeputyMandat, 0, len(mandats))
	for _, mandat := range mandats {
		views = append(views, components.DeputyMandat{
			UID:             mandat.UID,
			Legislature:     mandat.Legislature,
			TypeOrgane:      mandat.TypeOrgane,
			DateDebut:       mandat.DateDebut,
			DateFin:         mandat.DateFin,
			DatePublication: mandat.DatePublication,
			Preseance:       mandat.Preseance,
			NominPrincipale: mandat.NominPrincipale,
			CodeQualite:     mandat.CodeQualite,
			LibQualite:      mandat.LibQualite,
			LibQualiteSex:   mandat.LibQualiteSex,
			Organes:         deputyMandatOrganeViews(mandat.Organes),
		})
	}
	return views
}

func deputyMandatOrganeViews(organes []store.DeputyMandatOrgane) []components.DeputyMandatOrgane {
	views := make([]components.DeputyMandatOrgane, 0, len(organes))
	for _, organe := range organes {
		views = append(views, components.DeputyMandatOrgane{
			UID:     organe.UID,
			Libelle: organe.Libelle,
		})
	}
	return views
}

func deputyVoteStatViews(stats []store.DeputyVoteStat) []components.DeputyVoteStat {
	views := make([]components.DeputyVoteStat, 0, len(stats))
	for _, stat := range stats {
		views = append(views, components.DeputyVoteStat{
			Legislature: stat.Legislature,
			TotalVotes:  stat.TotalVotes,
			Pour:        stat.Pour,
			Contre:      stat.Contre,
			Abstentions: stat.Abstentions,
			NonVotants:  stat.NonVotants,
		})
	}
	return views
}

func deputyVoteViews(votes []store.DeputyVote) []components.DeputyVote {
	views := make([]components.DeputyVote, 0, len(votes))
	for _, vote := range votes {
		views = append(views, components.DeputyVote{
			ScrutinUID:    vote.ScrutinUID,
			Numero:        vote.Numero,
			Legislature:   vote.Legislature,
			Date:          vote.Date,
			Titre:         vote.Titre,
			SortCode:      vote.SortCode,
			SortLibelle:   vote.SortLibelle,
			TypeVote:      vote.TypeVote,
			Organe:        vote.Organe,
			GroupeUID:     vote.GroupeUID,
			Groupe:        vote.Groupe,
			MandatUID:     vote.MandatUID,
			Position:      vote.Position,
			ParDelegation: vote.ParDelegation,
			NumPlace:      vote.NumPlace,
		})
	}
	return views
}
