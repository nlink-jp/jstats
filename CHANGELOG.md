# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [1.0.0] - 2026-03-28

### Added
- Initial release
- SPL-style `stats` expression parser with `by` clause
- Functions: `count`, `sum`, `min`, `max`, `avg`, `median`, `stdev`, `var`, `range`, `p<N>`, `dc`, `first`, `last`, `mode`, `values`, `list`
- Output formats: `json` (default), `text`, `md`, `csv`
- JSON array and JSONL input support
- Alias support via `as` keyword (e.g. `avg(latency) as avg_ms`)

[Unreleased]: https://github.com/nlink-jp/jstats/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/nlink-jp/jstats/releases/tag/v1.0.0
