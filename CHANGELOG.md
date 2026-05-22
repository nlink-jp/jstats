# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [1.1.1] - 2026-05-23

### Changed

- **Darwin releases are now Developer ID signed and Apple-notarized.**
  `jstats-v1.1.1-darwin-{amd64,arm64}.zip` carry full Apple Developer
  ID Application signatures and notarization tickets from Apple. End
  users on macOS no longer need to bypass Gatekeeper with right-click
  → Open or `xattr -d com.apple.quarantine` on first launch; local
  users who place `jstats` under Dropbox-synced (or any other
  FileProvider-managed) paths are no longer killed by macOS's
  ad-hoc + provenance distrust policy. Pipeline:
  `scripts/codesign-darwin.sh` + `scripts/notarize-darwin.sh`,
  driven by `make package`. Adopts the org-wide convention in
  `nlink-jp/.github` CONVENTIONS.md §Code Signing.

No behaviour change to the binary itself — feature-wise this is
identical to v1.1.0.

## [1.1.0] - 2026-04-17

### Fixed

- Numeric functions (sum, min, max, avg, median, stdev, var, range, p<N>) now handle string-encoded numbers (e.g. `"10"` instead of `10`)
- `values()` and `list()` return empty array `[]` instead of `null` when no values found

### Added

- Warning to stderr when a specified field is not found in any input record (helps catch typos)

## [1.0.0] - 2026-03-28

### Added
- Initial release
- SPL-style `stats` expression parser with `by` clause
- Functions: `count`, `sum`, `min`, `max`, `avg`, `median`, `stdev`, `var`, `range`, `p<N>`, `dc`, `first`, `last`, `mode`, `values`, `list`
- Output formats: `json` (default), `text`, `md`, `csv`
- JSON array and JSONL input support
- Alias support via `as` keyword (e.g. `avg(latency) as avg_ms`)

[Unreleased]: https://github.com/nlink-jp/jstats/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/nlink-jp/jstats/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/nlink-jp/jstats/releases/tag/v1.0.0
