package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

func render(rows []Row, headers []string, format string) (string, error) {
	switch format {
	case "json", "":
		return renderJSON(rows, headers)
	case "text":
		return renderText(rows, headers)
	case "md", "markdown":
		return renderMarkdown(rows, headers)
	case "csv":
		return renderCSV(rows, headers)
	default:
		return "", fmt.Errorf("unknown format %q: choose json, text, md, csv", format)
	}
}

func renderJSON(rows []Row, headers []string) (string, error) {
	// Output as ordered array; each element preserves header order.
	type orderedRow struct {
		keys []string
		vals map[string]interface{}
	}

	// Build output preserving column order.
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		m := make(map[string]interface{}, len(headers))
		for _, h := range headers {
			m[h] = r[h]
		}
		out[i] = m
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func renderText(rows []Row, headers []string) (string, error) {
	table := toStringTable(rows, headers)
	if len(table) == 0 {
		return "", nil
	}

	// Compute column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range table {
		for i, cell := range row {
			if w := utf8.RuneCountInString(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var sb strings.Builder
	sep := buildSep(widths)

	sb.WriteString(sep)
	sb.WriteString(buildRow(headers, widths))
	sb.WriteString(sep)
	for _, row := range table {
		sb.WriteString(buildRow(row, widths))
	}
	sb.WriteString(sep)
	return sb.String(), nil
}

func buildSep(widths []int) string {
	var sb strings.Builder
	sb.WriteByte('+')
	for _, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		sb.WriteByte('+')
	}
	sb.WriteByte('\n')
	return sb.String()
}

func buildRow(cells []string, widths []int) string {
	var sb strings.Builder
	sb.WriteByte('|')
	for i, cell := range cells {
		w := utf8.RuneCountInString(cell)
		sb.WriteByte(' ')
		sb.WriteString(cell)
		sb.WriteString(strings.Repeat(" ", widths[i]-w+1))
		sb.WriteByte('|')
	}
	sb.WriteByte('\n')
	return sb.String()
}

func renderMarkdown(rows []Row, headers []string) (string, error) {
	table := toStringTable(rows, headers)
	var sb strings.Builder

	// Header row
	sb.WriteByte('|')
	for _, h := range headers {
		sb.WriteString(" " + h + " |")
	}
	sb.WriteByte('\n')

	// Separator
	sb.WriteByte('|')
	for range headers {
		sb.WriteString(" --- |")
	}
	sb.WriteByte('\n')

	// Data rows
	for _, row := range table {
		sb.WriteByte('|')
		for _, cell := range row {
			sb.WriteString(" " + cell + " |")
		}
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}

func renderCSV(rows []Row, headers []string) (string, error) {
	table := toStringTable(rows, headers)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return "", err
	}
	for _, row := range table {
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}

// toStringTable converts result rows to a 2D string table in header order.
func toStringTable(rows []Row, headers []string) [][]string {
	table := make([][]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(headers))
		for j, h := range headers {
			cells[j] = cellString(row[h])
		}
		table[i] = cells
	}
	return table
}

func cellString(v interface{}) string {
	if v == nil {
		return ""
	}
	// Arrays (values/list) → compact JSON
	switch val := v.(type) {
	case []interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	}
	return strVal(v)
}
