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

func deputiesView(page store.DeputiesPage) components.DeputiesPage {
	return components.DeputiesPage{
		Query: components.DeputiesQuery{
			Search:      page.Query.Search,
			Sort:        page.Query.Sort,
			Page:        page.Query.Page,
			PerPage:     page.Query.PerPage,
			Legislature: page.Query.Legislature,
			Group:       page.Query.Group,
		},
		DefaultSort:   page.DefaultSort,
		SortOptions:   deputySortOptionViews(page.SortOptions),
		FilterOptions: deputyFilterOptionsView(page.FilterOptions),
		Deputies:      deputyListItemViews(page.Deputies),
		TotalResults:  page.TotalResults,
		TotalPages:    page.TotalPages,
		StartItem:     page.StartItem,
		EndItem:       page.EndItem,
	}
}

func politicalGroupsView(page store.PoliticalGroupsPage) components.PoliticalGroupsPage {
	return components.PoliticalGroupsPage{
		Query: components.PoliticalGroupsQuery{
			Search:      page.Query.Search,
			Sort:        page.Query.Sort,
			Page:        page.Query.Page,
			PerPage:     page.Query.PerPage,
			Legislature: page.Query.Legislature,
		},
		DefaultSort:   page.DefaultSort,
		SortOptions:   politicalGroupSortOptionViews(page.SortOptions),
		FilterOptions: politicalGroupFilterOptionsView(page.FilterOptions),
		Groups:        politicalGroupListItemViews(page.Groups),
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

func politicalGroupDetailView(page store.PoliticalGroupDetailPage) components.PoliticalGroupDetailPage {
	return components.PoliticalGroupDetailPage{
		Group: components.PoliticalGroupData{
			UID:               page.Group.UID,
			CodeType:          page.Group.CodeType,
			Libelle:           page.Group.Libelle,
			LibelleAbrege:     page.Group.LibelleAbrege,
			LibelleAbrev:      page.Group.LibelleAbrev,
			LibelleEdition:    page.Group.LibelleEdition,
			Legislature:       page.Group.Legislature,
			Chambre:           page.Group.Chambre,
			Regime:            page.Group.Regime,
			PositionPolitique: page.Group.PositionPolitique,
			CouleurAssociee:   page.Group.CouleurAssociee,
			Preseance:         page.Group.Preseance,
			DateDebut:         page.Group.DateDebut,
			DateFin:           page.Group.DateFin,
			SourceFile:        page.Group.SourceFile,
		},
		Query: components.PoliticalGroupDetailQuery{
			VotesPage:     page.Query.VotesPage,
			VotesPerPage:  page.Query.VotesPerPage,
			VotesSearch:   page.Query.VotesSearch,
			VotesSort:     page.Query.VotesSort,
			VotesPosition: page.Query.VotesPosition,
		},
		Stats:             politicalGroupVoteStatViews(page.Stats),
		Deputies:          politicalGroupDeputyViews(page.Deputies),
		Votes:             politicalGroupVoteViews(page.Votes),
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

func deputyListItemViews(items []store.DeputyListItem) []components.DeputyListItem {
	views := make([]components.DeputyListItem, 0, len(items))
	for _, item := range items {
		views = append(views, components.DeputyListItem{
			UID:           item.UID,
			DisplayName:   item.DisplayName,
			Alpha:         item.Alpha,
			Profession:    item.Profession,
			DateNaissance: item.DateNaissance,
			Legislature:   item.Legislature,
			GroupUID:      item.GroupUID,
			Group:         item.Group,
			GroupAbrege:   item.GroupAbrege,
			GroupAbrev:    item.GroupAbrev,
			TotalVotes:    item.TotalVotes,
			Pour:          item.Pour,
			Contre:        item.Contre,
			Abstentions:   item.Abstentions,
			NonVotants:    item.NonVotants,
		})
	}
	return views
}

func deputySortOptionViews(options []store.DeputySortOption) []components.DeputySortOption {
	views := make([]components.DeputySortOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.DeputySortOption{Value: option.Value, Label: option.Label})
	}
	return views
}

func deputyFilterOptionsView(options store.DeputyFilterOptions) components.DeputyFilterOptions {
	return components.DeputyFilterOptions{
		Legislatures: deputyFilterOptionViews(options.Legislatures),
		Groups:       deputyFilterOptionViews(options.Groups),
	}
}

func deputyFilterOptionViews(options []store.DeputyFilterOption) []components.DeputyFilterOption {
	views := make([]components.DeputyFilterOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.DeputyFilterOption{Value: option.Value, Label: option.Label})
	}
	return views
}

func politicalGroupListItemViews(items []store.PoliticalGroupListItem) []components.PoliticalGroupListItem {
	views := make([]components.PoliticalGroupListItem, 0, len(items))
	for _, item := range items {
		views = append(views, components.PoliticalGroupListItem{
			UID:           item.UID,
			Libelle:       item.Libelle,
			LibelleAbrege: item.LibelleAbrege,
			LibelleAbrev:  item.LibelleAbrev,
			Legislature:   item.Legislature,
			Position:      item.Position,
			Preseance:     item.Preseance,
			DateDebut:     item.DateDebut,
			DateFin:       item.DateFin,
			DeputiesCount: item.DeputiesCount,
			TotalScrutins: item.TotalScrutins,
			Pour:          item.Pour,
			Contre:        item.Contre,
			Abstentions:   item.Abstentions,
			NonVotants:    item.NonVotants,
		})
	}
	return views
}

func politicalGroupSortOptionViews(options []store.PoliticalGroupSortOption) []components.PoliticalGroupSortOption {
	views := make([]components.PoliticalGroupSortOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.PoliticalGroupSortOption{Value: option.Value, Label: option.Label})
	}
	return views
}

func politicalGroupFilterOptionsView(options store.PoliticalGroupFilterOptions) components.PoliticalGroupFilterOptions {
	return components.PoliticalGroupFilterOptions{
		Legislatures: politicalGroupFilterOptionViews(options.Legislatures),
	}
}

func politicalGroupFilterOptionViews(options []store.PoliticalGroupFilterOption) []components.PoliticalGroupFilterOption {
	views := make([]components.PoliticalGroupFilterOption, 0, len(options))
	for _, option := range options {
		views = append(views, components.PoliticalGroupFilterOption{Value: option.Value, Label: option.Label})
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
			UID:           organe.UID,
			CodeType:      organe.CodeType,
			Libelle:       organe.Libelle,
			LibelleAbrege: organe.LibelleAbrege,
			LibelleAbrev:  organe.LibelleAbrev,
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

func politicalGroupVoteStatViews(stats []store.PoliticalGroupVoteStat) []components.PoliticalGroupVoteStat {
	views := make([]components.PoliticalGroupVoteStat, 0, len(stats))
	for _, stat := range stats {
		views = append(views, components.PoliticalGroupVoteStat{
			Legislature:   stat.Legislature,
			TotalScrutins: stat.TotalScrutins,
			Pour:          stat.Pour,
			Contre:        stat.Contre,
			Abstentions:   stat.Abstentions,
			NonVotants:    stat.NonVotants,
		})
	}
	return views
}

func politicalGroupDeputyViews(deputies []store.PoliticalGroupDeputy) []components.PoliticalGroupDeputy {
	views := make([]components.PoliticalGroupDeputy, 0, len(deputies))
	for _, deputy := range deputies {
		views = append(views, components.PoliticalGroupDeputy{
			UID:         deputy.UID,
			DisplayName: deputy.DisplayName,
			Alpha:       deputy.Alpha,
			Legislature: deputy.Legislature,
			MandatUID:   deputy.MandatUID,
			DateDebut:   deputy.DateDebut,
			DateFin:     deputy.DateFin,
			Qualite:     deputy.Qualite,
		})
	}
	return views
}

func politicalGroupVoteViews(votes []store.PoliticalGroupVote) []components.PoliticalGroupVote {
	views := make([]components.PoliticalGroupVote, 0, len(votes))
	for _, vote := range votes {
		views = append(views, components.PoliticalGroupVote{
			ScrutinUID:            vote.ScrutinUID,
			Numero:                vote.Numero,
			Legislature:           vote.Legislature,
			Date:                  vote.Date,
			Titre:                 vote.Titre,
			SortCode:              vote.SortCode,
			SortLibelle:           vote.SortLibelle,
			TypeVote:              vote.TypeVote,
			PositionMajoritaire:   vote.PositionMajoritaire,
			NombreMembres:         vote.NombreMembres,
			Pour:                  vote.Pour,
			Contre:                vote.Contre,
			Abstentions:           vote.Abstentions,
			NonVotants:            vote.NonVotants,
			NonVotantsVolontaires: vote.NonVotantsVolontaires,
		})
	}
	return views
}
