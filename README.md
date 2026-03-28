# jstats

SPL-style `stats` command for JSON data. Reads a JSON array or JSONL stream from stdin and computes aggregations grouped by one or more fields.

## Install

```bash
go install github.com/nlink-jp/jstats@latest
```

Or download a binary from [Releases](https://github.com/nlink-jp/jstats/releases).

## Usage

```
jstats [flags] <expression>

Flags:
  -format string   output format: json (default), text, md, csv
  -version         print version and exit
```

### Expression syntax

```
func1, func2(field) [as alias], ... [by field1, field2, ...]
```

## Functions

| Function | Description |
|---|---|
| `count` | COUNT(*) — total rows in group |
| `count(field)` | count of non-null values |
| `sum(field)` | sum |
| `min(field)` | minimum |
| `max(field)` | maximum |
| `avg(field)` | average |
| `median(field)` | median (= p50) |
| `stdev(field)` | sample standard deviation |
| `var(field)` | sample variance |
| `range(field)` | max − min |
| `p<N>(field)` | Nth percentile (0–100), e.g. `p95(latency)` |
| `dc(field)` | distinct count |
| `first(field)` | first value (input order) |
| `last(field)` | last value (input order) |
| `mode(field)` | most frequent value |
| `values(field)` | distinct values as array |
| `list(field)` | all values as array |

## Examples

```bash
# Count by status code
cat access.json | jstats "count by status"

# Latency percentiles by service
cat metrics.json | jstats "count, avg(latency), p95(latency), p99(latency), stdev(latency) by service"

# Distinct values and mode
cat events.json | jstats "dc(user_id), values(action), mode(action) by host" -format text

# No grouping — whole dataset
cat data.json | jstats "count, min(score), max(score), avg(score), median(score)"

# Aliases
cat sales.json | jstats "sum(amount) as total, avg(amount) as avg_sale by region" -format md

# JSONL input works too
cat events.jsonl | jstats "count by type"
```

## Output formats

| Flag | Output |
|---|---|
| `json` (default) | JSON array |
| `text` | ASCII table |
| `md` | Markdown table |
| `csv` | CSV |
