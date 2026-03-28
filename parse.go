package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// AggFunc represents one aggregation function in the expression.
type AggFunc struct {
	Func  string // "count", "sum", "avg", "p", ...
	Field string // field name; empty for bare count
	Alias string // from "as <alias>"
	Perc  int    // percentile N for p<N>()
}

// OutputName returns the column name for this aggregate in the result.
func (a AggFunc) OutputName() string {
	if a.Alias != "" {
		return a.Alias
	}
	switch a.Func {
	case "count":
		if a.Field == "" {
			return "count"
		}
		return fmt.Sprintf("count(%s)", a.Field)
	case "p":
		return fmt.Sprintf("p%d(%s)", a.Perc, a.Field)
	default:
		return fmt.Sprintf("%s(%s)", a.Func, a.Field)
	}
}

// StatsQuery is the parsed result of a stats expression.
type StatsQuery struct {
	Funcs    []AggFunc
	ByFields []string
}

// knownFuncs lists all supported function names (excluding p<N> which is detected by pattern).
var knownFuncs = map[string]bool{
	"count": true, "sum": true, "min": true, "max": true, "avg": true,
	"median": true, "stdev": true, "var": true, "range": true,
	"dc": true, "first": true, "last": true, "mode": true,
	"values": true, "list": true,
}

// parseExpr parses a stats expression like:
//
//	"count, avg(latency) as avg_ms, p95(latency) by host, service"
func parseExpr(expr string) (StatsQuery, error) {
	// Split on "by" keyword (case-insensitive, word boundary).
	statsPart, byPart, err := splitByKeyword(expr)
	if err != nil {
		return StatsQuery{}, err
	}

	funcs, err := parseStatsList(statsPart)
	if err != nil {
		return StatsQuery{}, err
	}

	var byFields []string
	if byPart != "" {
		byFields = splitFields(byPart)
	}

	return StatsQuery{Funcs: funcs, ByFields: byFields}, nil
}

// splitByKeyword splits the expression into stats part and by part.
func splitByKeyword(expr string) (string, string, error) {
	// Find " by " with word boundaries using a simple scan.
	lower := strings.ToLower(expr)
	// Look for standalone "by" — preceded and followed by non-identifier chars.
	for i := 0; i < len(lower); i++ {
		if lower[i] != 'b' {
			continue
		}
		if i+2 > len(lower) {
			continue
		}
		if lower[i:i+2] != "by" {
			continue
		}
		// Check preceding char
		if i > 0 && isIdentChar(rune(lower[i-1])) {
			continue
		}
		// Check following char
		if i+2 < len(lower) && isIdentChar(rune(lower[i+2])) {
			continue
		}
		return strings.TrimSpace(expr[:i]), strings.TrimSpace(expr[i+2:]), nil
	}
	return strings.TrimSpace(expr), "", nil
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// parseStatsList parses a comma-separated list of aggregation function expressions.
func parseStatsList(s string) ([]AggFunc, error) {
	// Split on commas, respecting parentheses.
	parts := splitCommaRespectingParens(s)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty stats list")
	}

	var funcs []AggFunc
	for _, part := range parts {
		fn, err := parseSingleFunc(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}
	return funcs, nil
}

func splitCommaRespectingParens(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, c := range s {
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// parseSingleFunc parses one aggregation expression like "avg(latency) as avg_ms".
func parseSingleFunc(s string) (AggFunc, error) {
	var alias string

	// Check for "as <alias>" suffix (case-insensitive).
	lower := strings.ToLower(s)
	if idx := strings.Index(lower, " as "); idx >= 0 {
		alias = strings.TrimSpace(s[idx+4:])
		s = strings.TrimSpace(s[:idx])
	}

	// "count" with no parens
	if strings.ToLower(s) == "count" {
		return AggFunc{Func: "count", Alias: alias}, nil
	}

	// func(field)
	lparen := strings.Index(s, "(")
	rparen := strings.LastIndex(s, ")")
	if lparen < 0 || rparen < 0 || rparen < lparen {
		return AggFunc{}, fmt.Errorf("invalid function expression: %q", s)
	}

	fnName := strings.ToLower(strings.TrimSpace(s[:lparen]))
	field := strings.TrimSpace(s[lparen+1 : rparen])

	if field == "" {
		return AggFunc{}, fmt.Errorf("missing field in %q", s)
	}

	// p<N> pattern
	if strings.HasPrefix(fnName, "p") && len(fnName) > 1 {
		nStr := fnName[1:]
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 0 || n > 100 {
			return AggFunc{}, fmt.Errorf("invalid percentile %q: N must be 0–100", fnName)
		}
		return AggFunc{Func: "p", Field: field, Alias: alias, Perc: n}, nil
	}

	if !knownFuncs[fnName] {
		return AggFunc{}, fmt.Errorf("unknown function %q", fnName)
	}

	return AggFunc{Func: fnName, Field: field, Alias: alias}, nil
}

// splitFields splits a comma-separated list of field names.
func splitFields(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
