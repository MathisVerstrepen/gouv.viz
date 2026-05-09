package main

import (
	"regexp"
	"strings"
)

type scrutinLinkedReference struct {
	TextNum       string
	TextKind      string
	AmendementNum string
	Source        string
}

var (
	amendementReferenceRE = regexp.MustCompile(`(?i)\bamendements?\s*(?:n\s*[°ºo]\s*|num(?:e|é)ro\s*)?([[:alpha:]]{0,4}\s*\d+[[:alpha:]]?(?:\s*(?:rect(?:ifie|ifié)?|bis|ter|quater))?)`)
	textReferenceREs      = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\btexte\s*(?:n\s*[°ºo]\s*|num(?:e|é)ro\s*)?(\d+)`),
		regexp.MustCompile(`(?i)\b(?:projet|proposition)\s+de\s+loi[^()]*\(\s*n\s*[°ºo]\s*(\d+)`),
		regexp.MustCompile(`(?i)\b(?:projet|proposition)\s+de\s+loi[^,.]*?\s+n\s*[°ºo]\s*(\d+)`),
	}
	textKindREs = []struct {
		kind string
		re   *regexp.Regexp
	}{
		{kind: "proposition-loi", re: regexp.MustCompile(`(?i)\bproposition\s+de\s+loi\b`)},
		{kind: "projet-loi", re: regexp.MustCompile(`(?i)\bprojet\s+de\s+loi\b`)},
	}
)

func extractScrutinLinkedReference(title, objet string) scrutinLinkedReference {
	var ref scrutinLinkedReference
	var sources []string

	for _, candidate := range []struct {
		name string
		text string
	}{
		{name: "titre", text: title},
		{name: "objet", text: objet},
	} {
		if candidate.text == "" {
			continue
		}

		matched := false
		if ref.AmendementNum == "" {
			ref.AmendementNum = firstReferenceMatch(amendementReferenceRE, candidate.text)
			matched = ref.AmendementNum != ""
		}
		if ref.TextNum == "" {
			ref.TextNum = firstTextReference(candidate.text)
			matched = matched || ref.TextNum != ""
		}
		if ref.TextKind == "" {
			ref.TextKind = firstTextKind(candidate.text)
			matched = matched || ref.TextKind != ""
		}
		if matched {
			sources = appendSource(sources, candidate.name)
		}
	}

	ref.Source = strings.Join(sources, "+")
	return ref
}

func firstTextKind(text string) string {
	for _, candidate := range textKindREs {
		if candidate.re.MatchString(text) {
			return candidate.kind
		}
	}
	return ""
}

func firstTextReference(text string) string {
	for _, re := range textReferenceREs {
		if match := firstReferenceMatch(re, text); match != "" {
			return match
		}
	}
	return ""
}

func firstReferenceMatch(re *regexp.Regexp, text string) string {
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return cleanReferenceNumber(matches[1])
}

func cleanReferenceNumber(value string) string {
	value = strings.Trim(value, " \t\n\r.,;:)")
	return strings.Join(strings.Fields(value), " ")
}

func appendSource(sources []string, source string) []string {
	for _, existing := range sources {
		if existing == source {
			return sources
		}
	}
	return append(sources, source)
}
