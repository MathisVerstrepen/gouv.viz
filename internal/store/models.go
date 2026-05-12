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

type DeputiesQuery struct {
	Search      string
	Sort        string
	Page        int
	PerPage     int
	Legislature int
	Group       string
}

type DeputyListItem struct {
	UID           string
	DisplayName   string
	Alpha         string
	Profession    string
	DateNaissance string
	Legislature   int
	GroupUID      string
	Group         string
	GroupAbrege   string
	GroupAbrev    string
	TotalVotes    int
	Pour          int
	Contre        int
	Abstentions   int
	NonVotants    int
}

type DeputySortOption struct {
	Value string
	Label string
}

type DeputyFilterOption struct {
	Value string
	Label string
}

type DeputyFilterOptions struct {
	Legislatures []DeputyFilterOption
	Groups       []DeputyFilterOption
}

type DeputiesPage struct {
	Query         DeputiesQuery
	DefaultSort   string
	SortOptions   []DeputySortOption
	FilterOptions DeputyFilterOptions
	Deputies      []DeputyListItem
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

type DeputyDetailPage struct {
	Deputy            DeputyDetailData
	Query             DeputyDetailQuery
	Mandats           []DeputyMandat
	Stats             []DeputyVoteStat
	Votes             []DeputyVote
	VotesTotalResults int
	VotesTotalPages   int
	VotesStartItem    int
	VotesEndItem      int
}

type DeputyDetailQuery struct {
	VotesPage     int
	VotesPerPage  int
	VotesSearch   string
	VotesSort     string
	VotesPosition string
}

type DeputyDetailData struct {
	UID            string
	Civilite       string
	Prenom         string
	Nom            string
	Alpha          string
	DisplayName    string
	DateNaissance  string
	VilleNaissance string
	DepNaissance   string
	PaysNaissance  string
	DateDeces      string
	Profession     string
	URIHATVP       string
	SourceFile     string
}

type DeputyMandat struct {
	UID             string
	Legislature     int
	TypeOrgane      string
	DateDebut       string
	DateFin         string
	DatePublication string
	Preseance       int
	NominPrincipale bool
	CodeQualite     string
	LibQualite      string
	LibQualiteSex   string
	Organes         []DeputyMandatOrgane
}

type DeputyMandatOrgane struct {
	UID           string
	CodeType      string
	Libelle       string
	LibelleAbrege string
	LibelleAbrev  string
}

type DeputyVoteStat struct {
	Legislature int
	TotalVotes  int
	Pour        int
	Contre      int
	Abstentions int
	NonVotants  int
}

type DeputyVote struct {
	ScrutinUID    string
	Numero        int
	Legislature   int
	Date          string
	Titre         string
	SortCode      string
	SortLibelle   string
	TypeVote      string
	Organe        string
	GroupeUID     string
	Groupe        string
	MandatUID     string
	Position      string
	ParDelegation bool
	NumPlace      string
}
