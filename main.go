package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

var version = "dev"

func main() {
	format := flag.String("format", "json", "output format: json, text, md, csv")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: jstats [flags] <expression>

Compute SPL-style stats aggregations over a JSON array or JSONL stream.

Expression syntax:
  func1, func2(field) [as alias], ... [by field1, field2, ...]

Supported functions:
  count               count(*)
  count(field)        count of non-null values
  sum(field)          sum
  min(field)          minimum
  max(field)          maximum
  avg(field)          average
  median(field)       median (= p50)
  stdev(field)        sample standard deviation
  var(field)          sample variance
  range(field)        max - min
  p<N>(field)         Nth percentile, e.g. p95(latency)
  dc(field)           distinct count
  first(field)        first value (input order)
  last(field)         last value (input order)
  mode(field)         most frequent value
  values(field)       distinct values as array
  list(field)         all values as array

Examples:
  cat data.json | jstats "count by status"
  cat data.json | jstats "count, avg(duration), p95(duration) by service"
  cat data.json | jstats "count, values(status) by host" -format text

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "error: no input — pipe JSON or JSONL to stdin")
		os.Exit(1)
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	rows, err := parseInput(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing input: %v\n", err)
		os.Exit(1)
	}

	query, err := parseExpr(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing expression: %v\n", err)
		os.Exit(1)
	}

	result, headers, err := computeStats(rows, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing stats: %v\n", err)
		os.Exit(1)
	}

	out, err := render(result, headers, *format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error rendering output: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}

// parseInput accepts a JSON array or JSONL stream.
func parseInput(data []byte) ([]Row, error) {
	data = trimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// Try JSON array first
	if data[0] == '[' {
		var rows []Row
		if err := json.Unmarshal(data, &rows); err != nil {
			return nil, fmt.Errorf("JSON array parse: %w", err)
		}
		return rows, nil
	}

	// Try JSONL
	var rows []Row
	dec := json.NewDecoder(bytesReader(data))
	for dec.More() {
		var row Row
		if err := dec.Decode(&row); err != nil {
			return nil, fmt.Errorf("JSONL parse: %w", err)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no records found in input")
	}
	return rows, nil
}
