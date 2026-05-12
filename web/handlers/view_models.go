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
		GroupVotes: scrutinGroupVoteViews(page.GroupVotes),
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
