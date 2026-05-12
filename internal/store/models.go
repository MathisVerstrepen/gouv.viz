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
	Search      string
	Sort        string
	Page        int
	PerPage     int
	Legislature int
	Result      string
	VoteType    string
	Organe      string
	DateFrom    string
	DateTo      string
	CloseVotes  bool
}

type ScrutinSortOption struct {
	Value string
	Label string
}

type ScrutinFilterOption struct {
	Value string
	Label string
}

type ScrutinFilterOptions struct {
	Legislatures []ScrutinFilterOption
	Results      []ScrutinFilterOption
	VoteTypes    []ScrutinFilterOption
	Organes      []ScrutinFilterOption
}

type ScrutinsPage struct {
	Query         ScrutinsQuery
	DefaultSort   string
	SortOptions   []ScrutinSortOption
	FilterOptions ScrutinFilterOptions
	Scrutins      []ScrutinListItem
	TotalResults  int
	TotalPages    int
	StartItem     int
	EndItem       int
}

type ScrutinDetailPage struct {
	Scrutin         ScrutinDetailData
	GroupVotes      []ScrutinGroupVote
	IndividualVotes []ScrutinIndividualVote
}

type ScrutinDetailData struct {
	UID                     string
	Numero                  int
	Legislature             int
	OrganeUID               string
	Organe                  string
	SessionRef              string
	SeanceRef               string
	Date                    string
	QuantiemeJourSeance     int
	CodeTypeVote            string
	TypeVote                string
	TypeMajorite            string
	SortCode                string
	SortLibelle             string
	Titre                   string
	LinkedTextNum           string
	LinkedTextKind          string
	LinkedTextURL           string
	LinkedTextPDFURL        string
	LinkedDossierRef        string
	LinkedDossierLibelle    string
	LinkedAmendementNum     string
	LinkedAmendementTextNum string
	LinkedAmendementOrgane  string
	LinkedAmendementURL     string
	LinkedReferenceSource   string
	Demandeur               string
	Objet                   string
	ModePublicationVotes    string
	NombreVotants           int
	SuffragesExprimes       int
	SuffragesRequis         int
	NonVotants              int
	Pour                    int
	Contre                  int
	Abstentions             int
	NonVotantsVolontaires   int
	SourceFile              string
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

type ScrutinIndividualVote struct {
	GroupeUID     string
	Groupe        string
	ActeurUID     string
	Depute        string
	Alpha         string
	MandatUID     string
	Position      string
	ParDelegation bool
	NumPlace      string
}
