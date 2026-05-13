package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"gouv.viz/internal/store"
)

type samples struct {
	ScrutinUID             string
	ScrutinSearch          string
	ScrutinLegislature     int
	ScrutinResult          string
	ScrutinVoteType        string
	ScrutinOrgane          string
	DeputyUID              string
	DeputySearch           string
	DeputyVotePosition     string
	DeputyVoteSearch       string
	PoliticalGroupUID      string
	PoliticalGroupSearch   string
	PoliticalGroupPosition string
	PoliticalGroupVoteTerm string
	RowCounts              map[string]int
}

type timedCase struct {
	Name string
	Run  func(context.Context) error
}

type timedResult struct {
	Name string
	Runs int
	Min  time.Duration
	Avg  time.Duration
	Max  time.Duration
	Err  error
}

type planCheck struct {
	Name               string
	SQL                string
	Args               []any
	ForbiddenFullScans []string
	RequireAny         []string
}

type planResult struct {
	Name    string
	Details []string
	Issues  []string
}

var tokenRE = regexp.MustCompile(`[\p{L}\p{N}]+`)

func main() {
	dbPath := flag.String("db", filepath.Join("data", "processed", "gouv-viz.sqlite"), "generated SQLite database path")
	runs := flag.Int("runs", 3, "timed runs per representative query")
	slow := flag.Duration("slow", 250*time.Millisecond, "warn when average query time exceeds this duration")
	strict := flag.Bool("strict", false, "exit non-zero when slow queries or plan issues are reported")
	flag.Parse()

	if *runs < 1 {
		log.Fatal("-runs must be at least 1")
	}

	if _, err := os.Stat(*dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("database %s does not exist; run make preprocess first or pass -db", *dbPath)
		}
		log.Fatalf("stat database: %v", err)
	}

	db, err := sql.Open("sqlite", readOnlyDSN(*dbPath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	st := store.New(db)
	if err := st.Validate(ctx); err != nil {
		log.Fatalf("validate database: %v", err)
	}

	s, err := loadSamples(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("Rows: acteurs=%d organes=%d scrutins=%d scrutin_groupe_votes=%d votes=%d\n",
		s.RowCounts["acteurs"], s.RowCounts["organes"], s.RowCounts["scrutins"], s.RowCounts["scrutin_groupe_votes"], s.RowCounts["votes"])
	fmt.Printf("Timing: runs=%d slow_threshold=%s\n\n", *runs, *slow)

	results := runTimedCases(ctx, *runs, representativeCases(st, s))
	hasSlow := printTimedResults(results, *slow)

	fmt.Println()
	planResults := runPlanChecks(ctx, db, representativePlanChecks(s))
	hasPlanIssue := printPlanResults(planResults)

	if *strict && (hasSlow || hasPlanIssue) {
		os.Exit(1)
	}
}

func readOnlyDSN(databasePath string) string {
	query := url.Values{}
	query.Set("mode", "ro")
	query.Set("immutable", "1")
	query.Add("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "query_only(1)")

	if !filepath.IsAbs(databasePath) {
		return "file:" + (&url.URL{Path: databasePath}).EscapedPath() + "?" + query.Encode()
	}

	return (&url.URL{Scheme: "file", Path: databasePath, RawQuery: query.Encode()}).String()
}

func loadSamples(ctx context.Context, db *sql.DB) (samples, error) {
	s := samples{RowCounts: make(map[string]int)}
	for _, table := range []string{"acteurs", "organes", "scrutins", "scrutin_groupe_votes", "votes"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			return samples{}, fmt.Errorf("count %s: %w", table, err)
		}
		s.RowCounts[table] = count
	}

	var scrutinTitle string
	if err := db.QueryRowContext(ctx, `
SELECT uid, COALESCE(titre, ''), legislature, COALESCE(sort_code, ''), COALESCE(code_type_vote, ''), COALESCE(organe_uid, '')
FROM scrutins
ORDER BY COALESCE(nombre_votants, 0) DESC, date_scrutin DESC, numero DESC
LIMIT 1
`).Scan(&s.ScrutinUID, &scrutinTitle, &s.ScrutinLegislature, &s.ScrutinResult, &s.ScrutinVoteType, &s.ScrutinOrgane); err != nil {
		return samples{}, fmt.Errorf("sample scrutin: %w", err)
	}
	s.ScrutinSearch = searchToken(scrutinTitle)
	s.DeputyVoteSearch = s.ScrutinSearch
	s.PoliticalGroupVoteTerm = s.ScrutinSearch

	if err := db.QueryRowContext(ctx, `
SELECT acteur_uid
FROM acteur_vote_stats
GROUP BY acteur_uid
ORDER BY SUM(total_votes) DESC, acteur_uid
LIMIT 1
`).Scan(&s.DeputyUID); err != nil {
		return samples{}, fmt.Errorf("sample deputy: %w", err)
	}

	var deputyLabel string
	if err := db.QueryRowContext(ctx, `
SELECT COALESCE(NULLIF(alpha, ''), nom, prenom, uid, '')
FROM acteurs
WHERE uid = ?
`, s.DeputyUID).Scan(&deputyLabel); err != nil {
		return samples{}, fmt.Errorf("sample deputy label: %w", err)
	}
	s.DeputySearch = searchToken(deputyLabel)

	if err := db.QueryRowContext(ctx, `
SELECT position
FROM votes
WHERE acteur_uid = ?
GROUP BY position
ORDER BY COUNT(*) DESC, position
LIMIT 1
`, s.DeputyUID).Scan(&s.DeputyVotePosition); err != nil {
		return samples{}, fmt.Errorf("sample deputy vote position: %w", err)
	}

	if err := db.QueryRowContext(ctx, `
SELECT groupe_uid
FROM groupe_vote_stats
GROUP BY groupe_uid
ORDER BY SUM(total_scrutins) DESC, groupe_uid
LIMIT 1
`).Scan(&s.PoliticalGroupUID); err != nil {
		return samples{}, fmt.Errorf("sample political group: %w", err)
	}

	var groupLabel string
	if err := db.QueryRowContext(ctx, `
SELECT COALESCE(NULLIF(libelle_abrege, ''), libelle, uid, '')
FROM organes
WHERE uid = ?
`, s.PoliticalGroupUID).Scan(&groupLabel); err != nil {
		return samples{}, fmt.Errorf("sample political group label: %w", err)
	}
	s.PoliticalGroupSearch = searchToken(groupLabel)

	if err := db.QueryRowContext(ctx, `
SELECT COALESCE(position_majoritaire, '')
FROM scrutin_groupe_votes
WHERE groupe_uid = ? AND COALESCE(position_majoritaire, '') <> ''
GROUP BY position_majoritaire
ORDER BY COUNT(*) DESC, position_majoritaire
LIMIT 1
`, s.PoliticalGroupUID).Scan(&s.PoliticalGroupPosition); err != nil {
		return samples{}, fmt.Errorf("sample political group vote position: %w", err)
	}

	return s, nil
}

func searchToken(value string) string {
	for _, token := range tokenRE.FindAllString(value, -1) {
		if len([]rune(token)) >= 4 {
			return token
		}
	}
	return strings.TrimSpace(value)
}

func representativeCases(st *store.Store, s samples) []timedCase {
	cases := []timedCase{
		{
			Name: "home_page_static_cache",
			Run: func(ctx context.Context) error {
				_, err := st.HomePage(ctx)
				return err
			},
		},
		{
			Name: "scrutins_recent_page",
			Run: func(ctx context.Context) error {
				_, err := st.ScrutinsPage(ctx, store.ScrutinsQuery{Sort: "date_desc", Page: 1})
				return err
			},
		},
		{
			Name: "scrutins_filters_pagination",
			Run: func(ctx context.Context) error {
				_, err := st.ScrutinsPage(ctx, store.ScrutinsQuery{
					Sort:        "votants_desc",
					Page:        3,
					Legislature: s.ScrutinLegislature,
					Result:      s.ScrutinResult,
					VoteType:    s.ScrutinVoteType,
					Organe:      s.ScrutinOrgane,
				})
				return err
			},
		},
		{
			Name: "deputies_votes_sort",
			Run: func(ctx context.Context) error {
				_, err := st.DeputiesPage(ctx, store.DeputiesQuery{Sort: "votes_desc", Page: 1})
				return err
			},
		},
		{
			Name: "deputies_group_search",
			Run: func(ctx context.Context) error {
				_, err := st.DeputiesPage(ctx, store.DeputiesQuery{Search: s.DeputySearch, Group: s.PoliticalGroupUID, Sort: "groupe_asc", Page: 1})
				return err
			},
		},
		{
			Name: "political_groups_votes_sort",
			Run: func(ctx context.Context) error {
				_, err := st.PoliticalGroupsPage(ctx, store.PoliticalGroupsQuery{Sort: "votes_desc", Page: 1})
				return err
			},
		},
		{
			Name: "scrutin_detail_votes",
			Run: func(ctx context.Context) error {
				_, err := st.ScrutinDetailPage(ctx, s.ScrutinUID)
				return err
			},
		},
		{
			Name: "deputy_detail_vote_search",
			Run: func(ctx context.Context) error {
				_, err := st.DeputyDetailPage(ctx, s.DeputyUID, store.DeputyDetailQuery{
					VotesPage:     2,
					VotesSearch:   s.DeputyVoteSearch,
					VotesPosition: s.DeputyVotePosition,
					VotesSort:     "date_desc",
				})
				return err
			},
		},
		{
			Name: "political_group_detail_vote_search",
			Run: func(ctx context.Context) error {
				_, err := st.PoliticalGroupDetailPage(ctx, s.PoliticalGroupUID, store.PoliticalGroupDetailQuery{
					VotesPage:     2,
					VotesSearch:   s.PoliticalGroupVoteTerm,
					VotesPosition: s.PoliticalGroupPosition,
					VotesSort:     "date_desc",
				})
				return err
			},
		},
	}

	if s.ScrutinSearch != "" {
		cases = append(cases, timedCase{
			Name: "scrutins_fts_search",
			Run: func(ctx context.Context) error {
				_, err := st.ScrutinsPage(ctx, store.ScrutinsQuery{Search: s.ScrutinSearch, Sort: "closest", Page: 1})
				return err
			},
		})
	}

	return cases
}

func runTimedCases(ctx context.Context, runs int, cases []timedCase) []timedResult {
	results := make([]timedResult, 0, len(cases))
	for _, c := range cases {
		result := timedResult{Name: c.Name, Runs: runs}
		var total time.Duration
		for i := 0; i < runs; i++ {
			start := time.Now()
			err := c.Run(ctx)
			duration := time.Since(start)
			if err != nil {
				result.Err = err
				break
			}
			if i == 0 || duration < result.Min {
				result.Min = duration
			}
			if duration > result.Max {
				result.Max = duration
			}
			total += duration
		}
		if result.Err == nil {
			result.Avg = total / time.Duration(runs)
		}
		results = append(results, result)
	}
	return results
}

func printTimedResults(results []timedResult, slow time.Duration) bool {
	hasSlow := false
	fmt.Println("Representative store queries:")
	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("  ERROR %-36s %v\n", result.Name, result.Err)
			hasSlow = true
			continue
		}
		status := "OK"
		if result.Avg > slow {
			status = "SLOW"
			hasSlow = true
		}
		fmt.Printf("  %-4s %-36s avg=%-9s min=%-9s max=%-9s\n", status, result.Name, roundDuration(result.Avg), roundDuration(result.Min), roundDuration(result.Max))
	}
	return hasSlow
}

func representativePlanChecks(s samples) []planCheck {
	checks := []planCheck{
		{
			Name: "deputy_votes_by_actor",
			SQL: `
SELECT s.uid
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
LEFT JOIN organes o ON o.uid = s.organe_uid
LEFT JOIN organes g ON g.uid = v.groupe_uid
WHERE v.acteur_uid = ? AND v.position = ?
ORDER BY s.date_scrutin DESC, s.numero DESC, s.uid
LIMIT 50 OFFSET 50
`,
			Args:               []any{s.DeputyUID, s.DeputyVotePosition},
			ForbiddenFullScans: []string{"v", "votes"},
		},
		{
			Name: "scrutin_individual_votes",
			SQL: `
SELECT v.acteur_uid
FROM votes v
LEFT JOIN acteurs a ON a.uid = v.acteur_uid
LEFT JOIN organes g ON g.uid = v.groupe_uid
WHERE v.scrutin_uid = ?
ORDER BY COALESCE(g.preseance, 9999), COALESCE(a.alpha, a.nom, a.prenom, v.acteur_uid), v.acteur_uid
`,
			Args:               []any{s.ScrutinUID},
			ForbiddenFullScans: []string{"v", "votes"},
		},
		{
			Name: "political_group_votes_by_group",
			SQL: `
SELECT COUNT(*)
FROM scrutin_groupe_votes sgv
JOIN scrutins s ON s.uid = sgv.scrutin_uid
WHERE sgv.groupe_uid = ?
`,
			Args:               []any{s.PoliticalGroupUID},
			ForbiddenFullScans: []string{"sgv", "scrutin_groupe_votes", "s", "scrutins"},
		},
		{
			Name: "political_group_deputies",
			SQL: `
WITH group_mandats AS (
  SELECT m.uid, m.acteur_uid
  FROM mandats m
  JOIN mandat_organes mo ON mo.mandat_uid = m.uid
  WHERE mo.organe_uid = ?
)
SELECT a.uid
FROM group_mandats gm
JOIN acteurs a ON a.uid = gm.acteur_uid
ORDER BY COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid), a.uid
`,
			Args:               []any{s.PoliticalGroupUID},
			ForbiddenFullScans: []string{"mo", "mandat_organes"},
		},
	}

	if s.ScrutinSearch != "" {
		checks = append(checks, planCheck{
			Name: "scrutin_search_fts",
			SQL: `
SELECT uid
FROM scrutin_search
WHERE scrutin_search MATCH ?
`,
			Args:       []any{`"` + strings.ReplaceAll(s.ScrutinSearch, `"`, `""`) + `"*`},
			RequireAny: []string{"VIRTUAL TABLE INDEX"},
		})
	}

	sort.SliceStable(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	return checks
}

func runPlanChecks(ctx context.Context, db *sql.DB, checks []planCheck) []planResult {
	results := make([]planResult, 0, len(checks))
	for _, check := range checks {
		result := planResult{Name: check.Name}
		rows, err := db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+check.SQL, check.Args...)
		if err != nil {
			result.Issues = append(result.Issues, err.Error())
			results = append(results, result)
			continue
		}

		for rows.Next() {
			var id, parent, notUsed int
			var detail string
			if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
				result.Issues = append(result.Issues, err.Error())
				break
			}
			result.Details = append(result.Details, detail)
		}
		if err := rows.Err(); err != nil {
			result.Issues = append(result.Issues, err.Error())
		}
		rows.Close()

		result.Issues = append(result.Issues, planIssues(check, result.Details)...)
		results = append(results, result)
	}
	return results
}

func planIssues(check planCheck, details []string) []string {
	var issues []string
	for _, detail := range details {
		upperDetail := strings.ToUpper(detail)
		if strings.Contains(upperDetail, "USING AUTOMATIC") {
			issues = append(issues, "uses an automatic index; consider adding a permanent schema index")
		}
		for _, table := range check.ForbiddenFullScans {
			if isFullScan(detail, table) {
				issues = append(issues, "full scan in plan: "+detail)
			}
		}
	}
	if len(check.RequireAny) > 0 {
		required := false
		for _, detail := range details {
			for _, requiredText := range check.RequireAny {
				if strings.Contains(strings.ToUpper(detail), strings.ToUpper(requiredText)) {
					required = true
				}
			}
		}
		if !required {
			issues = append(issues, "missing expected plan fragment: "+strings.Join(check.RequireAny, " or "))
		}
	}
	return issues
}

func isFullScan(detail, table string) bool {
	upperDetail := strings.ToUpper(detail)
	upperTable := strings.ToUpper(table)
	if !strings.Contains(upperDetail, "SCAN "+upperTable) {
		return false
	}
	return !strings.Contains(upperDetail, " USING ")
}

func printPlanResults(results []planResult) bool {
	hasIssue := false
	fmt.Println("Query plan checks:")
	for _, result := range results {
		status := "OK"
		if len(result.Issues) > 0 {
			status = "WARN"
			hasIssue = true
		}
		fmt.Printf("  %-4s %s\n", status, result.Name)
		for _, issue := range result.Issues {
			fmt.Printf("       issue: %s\n", issue)
		}
		for _, detail := range result.Details {
			fmt.Printf("       plan: %s\n", detail)
		}
	}
	return hasIssue
}

func roundDuration(duration time.Duration) time.Duration {
	if duration >= time.Second {
		return duration.Round(10 * time.Millisecond)
	}
	if duration >= time.Millisecond {
		return duration.Round(100 * time.Microsecond)
	}
	return duration.Round(time.Microsecond)
}
