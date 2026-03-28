# jstats

SPL-style `stats` command for JSON data. Reads a JSON array or JSONL stream from stdin and computes aggregations grouped by one or more fields.

Japanese documentation: [README.ja.md](README.ja.md)

## Install

Download the latest binary for your platform from the [releases page](https://github.com/nlink-jp/jstats/releases).

```sh
unzip jstats-<version>-<os>-<arch>.zip
mv jstats /usr/local/bin/
```

Or build from source (requires Go 1.24+):

```sh
go install github.com/nlink-jp/jstats@latest
```

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
cat events.json | jstats -format text "dc(user_id), values(action), mode(action) by host"

# No grouping — whole dataset
cat data.json | jstats "count, min(score), max(score), avg(score), median(score)"

# Aliases
cat sales.json | jstats -format md "sum(amount) as total, avg(amount) as avg_sale by region"

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

## Build

```sh
git clone https://github.com/nlink-jp/jstats.git
cd jstats
make build        # current platform → dist/jstats
make build-all    # cross-compile for all platforms → dist/
make package      # build + create .zip archives
make test         # run tests
make clean        # remove dist/
```

## How it works

```
stdin (JSON array or JSONL)
        │
        ▼
  parseInput()          Parse bytes into []map[string]interface{}
        │               Accepts both [{"k":"v"}] and {"k":"v"}\n{"k":"v"}
        ▼
  parseExpr()           Tokenize and parse the stats expression
        │               e.g. "count, avg(latency) by host"
        │               → StatsQuery{Funcs: [...], ByFields: [...]}
        ▼
  computeStats()        Group rows by ByFields, then apply each AggFunc
        │               Groups are ordered by first appearance (stable output)
        │
        ├── count / sum / min / max / avg / range / dc
        │     Direct iteration over numeric or string values
        │
        ├── stdev / var
        │     Two-pass: mean → sum of squared deviations / (n-1)
        │
        ├── median / p<N>
        │     Sort copy of field values → linear interpolation
        │
        ├── mode
        │     Frequency map; first-seen order breaks ties
        │
        └── values / list / first / last
              Slice accumulation preserving input order
        │
        ▼
    render()             Format result rows as json / text / md / csv
        │
        ▼
     stdout
```

Null and missing field values are silently skipped in all numeric functions. `count(field)` counts only non-null values; bare `count` always counts all rows in the group.
