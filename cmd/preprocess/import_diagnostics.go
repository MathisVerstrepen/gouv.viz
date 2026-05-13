package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

const maxDiagnosticExamples = 25

type importDiagnostics struct {
	counts   map[string]int
	examples []importDiagnostic
}

type importDiagnostic struct {
	Kind    string
	Source  string
	Message string
}

func (d *importDiagnostics) Warn(kind, source, format string, args ...any) {
	if d == nil {
		return
	}
	if d.counts == nil {
		d.counts = make(map[string]int)
	}
	d.counts[kind]++
	if len(d.examples) < maxDiagnosticExamples {
		d.examples = append(d.examples, importDiagnostic{
			Kind:    kind,
			Source:  source,
			Message: fmt.Sprintf(format, args...),
		})
	}
}

func (d importDiagnostics) Total() int {
	total := 0
	for _, count := range d.counts {
		total += count
	}
	return total
}

func (d importDiagnostics) Count(kind string) int {
	return d.counts[kind]
}

func (d importDiagnostics) Report() string {
	if d.Total() == 0 {
		return ""
	}

	kinds := make([]string, 0, len(d.counts))
	for kind := range d.counts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)

	var report strings.Builder
	fmt.Fprintf(&report, "import anomalies: total=%d", d.Total())
	for _, kind := range kinds {
		fmt.Fprintf(&report, " %s=%d", kind, d.counts[kind])
	}
	for _, example := range d.examples {
		fmt.Fprintf(&report, "\n- [%s] %s: %s", example.Kind, example.Source, example.Message)
	}
	if len(d.examples) < d.Total() {
		fmt.Fprintf(&report, "\n- ... %d more anomalies omitted", d.Total()-len(d.examples))
	}
	return report.String()
}

func collectDatabaseDiagnostics(tx *sql.Tx, diagnostics *importDiagnostics) error {
	checks := []struct {
		kind  string
		query string
	}{
		{
			kind: "unresolved_organe",
			query: `
SELECT COALESCE(source_file, ''), uid, organe_parent_uid
FROM organes
WHERE organe_parent_uid IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM organes parent WHERE parent.uid = organes.organe_parent_uid)
`,
		},
		{
			kind: "unresolved_organe",
			query: `
SELECT COALESCE(source_file, ''), uid, organe_uid
FROM scrutins
WHERE organe_uid IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM organes o WHERE o.uid = scrutins.organe_uid)
`,
		},
		{
			kind: "unresolved_organe",
			query: `
SELECT COALESCE(s.source_file, ''), sgv.scrutin_uid, sgv.groupe_uid
FROM scrutin_groupe_votes sgv
JOIN scrutins s ON s.uid = sgv.scrutin_uid
WHERE NOT EXISTS (SELECT 1 FROM organes o WHERE o.uid = sgv.groupe_uid)
`,
		},
		{
			kind: "unresolved_organe",
			query: `
SELECT COALESCE(s.source_file, ''), v.scrutin_uid || '/' || v.acteur_uid, v.groupe_uid
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
WHERE v.groupe_uid IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM organes o WHERE o.uid = v.groupe_uid)
`,
		},
		{
			kind: "unresolved_organe",
			query: `
SELECT COALESCE(a.source_file, ''), mo.mandat_uid, mo.organe_uid
FROM mandat_organes mo
JOIN mandats m ON m.uid = mo.mandat_uid
JOIN acteurs a ON a.uid = m.acteur_uid
WHERE NOT EXISTS (SELECT 1 FROM organes o WHERE o.uid = mo.organe_uid)
`,
		},
		{
			kind: "unresolved_acteur",
			query: `
SELECT COALESCE(s.source_file, ''), v.scrutin_uid || '/' || v.acteur_uid, v.acteur_uid
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
WHERE NOT EXISTS (SELECT 1 FROM acteurs a WHERE a.uid = v.acteur_uid)
`,
		},
		{
			kind: "unresolved_mandat",
			query: `
SELECT COALESCE(s.source_file, ''), v.scrutin_uid || '/' || v.acteur_uid, v.mandat_uid
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
WHERE v.mandat_uid IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM mandats m WHERE m.uid = v.mandat_uid)
`,
		},
		{
			kind: "unknown_vote_position",
			query: `
SELECT COALESCE(s.source_file, ''), v.scrutin_uid || '/' || v.acteur_uid, v.position
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
WHERE LOWER(REPLACE(REPLACE(TRIM(v.position), '-', '_'), ' ', '_')) NOT IN ('pour', 'contre', 'abstention', 'non_votant', 'non_votant_volontaire')
`,
		},
	}

	for _, check := range checks {
		if err := collectDatabaseDiagnostic(tx, diagnostics, check.kind, check.query); err != nil {
			return err
		}
	}
	return nil
}

func collectDatabaseDiagnostic(tx *sql.Tx, diagnostics *importDiagnostics, kind, query string) error {
	rows, err := tx.Query(query)
	if err != nil {
		return fmt.Errorf("query import diagnostics %s: %w", kind, err)
	}
	defer rows.Close()

	for rows.Next() {
		var source, subject, value string
		if err := rows.Scan(&source, &subject, &value); err != nil {
			return fmt.Errorf("scan import diagnostics %s: %w", kind, err)
		}
		diagnostics.Warn(kind, source, "%s references %q", subject, value)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate import diagnostics %s: %w", kind, err)
	}
	return nil
}

func withDiagnostics(err error, diagnostics importDiagnostics) error {
	if err == nil || diagnostics.Total() == 0 {
		return err
	}
	return fmt.Errorf("%w\n%s", err, diagnostics.Report())
}
