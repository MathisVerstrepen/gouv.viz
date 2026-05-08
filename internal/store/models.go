package store

type HomePage struct {
	TotalScrutins    int
	FirstScrutinDate string
	LastScrutinDate  string
	Scrutins         []ScrutinListItem
}

type ScrutinListItem struct {
	UID           string
	Numero        int
	Date          string
	Titre         string
	SortCode      string
	TypeVote      string
	Organe        string
	Pour          int
	Contre        int
	Abstentions   int
	NombreVotants int
}

type ScrutinsQuery struct {
	Search  string
	Sort    string
	Page    int
	PerPage int
}

type ScrutinSortOption struct {
	Value string
	Label string
}

type ScrutinsPage struct {
	Query        ScrutinsQuery
	SortOptions  []ScrutinSortOption
	Scrutins     []ScrutinListItem
	TotalResults int
	TotalPages   int
	StartItem    int
	EndItem      int
}

type ScrutinDetailPage struct {
	Scrutin    ScrutinDetailData
	GroupVotes []ScrutinGroupVote
}

type ScrutinDetailData struct {
	UID                   string
	Numero                int
	Legislature           int
	OrganeUID             string
	Organe                string
	SessionRef            string
	SeanceRef             string
	Date                  string
	QuantiemeJourSeance   int
	CodeTypeVote          string
	TypeVote              string
	TypeMajorite          string
	SortCode              string
	SortLibelle           string
	Titre                 string
	Demandeur             string
	Objet                 string
	ModePublicationVotes  string
	NombreVotants         int
	SuffragesExprimes     int
	SuffragesRequis       int
	NonVotants            int
	Pour                  int
	Contre                int
	Abstentions           int
	NonVotantsVolontaires int
	SourceFile            string
}

type ScrutinGroupVote struct {
	GroupeUID             string
	Groupe                string
	NombreMembres         int
	PositionMajoritaire   string
	NonVotants            int
	Pour                  int
	Contre                int
	Abstentions           int
	NonVotantsVolontaires int
}
