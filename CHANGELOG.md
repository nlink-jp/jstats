# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Fixed

- **`make verify-release` now fails closed.** Its last block chained unzip, the
  packaged binary's `--version` and `spctl` with `&&` and ended the whole chain
  in `|| true`, so a zip that did not unpack or a binary that did not run exited
  0 and the upload proceeded. Each step is now judged on its own, the packaged
  binary's `--version` must contain the tag being released, and only the
  informational `spctl` line may be ignored. Matches the org template
  (CONVENTIONS.md §Code Signing → Verifying a release).
- **The Linux archives no longer carry macOS file metadata.** macOS `tar` wrote
  each bundled file's extended attributes (`com.apple.provenance`, and a Dropbox
  attribute where the tree is synced) into the `.tar.gz` twice: as AppleDouble
  `._` members, which GNU tar extracts as stray `._<name>` files beside the real
  ones, and as `LIBARCHIVE.xattr.*` / `SCHILY.xattr.*` pax headers, which it
  reports as unknown keywords. `make package` now archives with
  `COPYFILE_DISABLE=1 tar --no-xattrs`; each setting stops one of the two.
  Archives already published still carry them; the files themselves are
  unaffected.

### Internal

- `make verify-release` also judges each Linux archive: no AppleDouble or other
  macOS metadata members — listed with `--options 'tar:!mac-ext'`, because a
  plain macOS listing folds `._` members away — no extended attributes as pax
  headers, and exactly the canonical binary, `README.md` and `LICENSE`, compared
  in the C locale.

## [1.2.0] - 2026-07-12

### Removed

- **darwin/amd64 (Intel) pre-built binary.** macOS releases now ship
  **arm64 only**, per the org-wide policy (darwin is Apple-Silicon only; no
  universal binaries). Intel Mac users can build from source.

### Changed

- **Linux release archives are now `.tar.gz`** (darwin/windows remain `.zip`),
  per `nlink-jp/.github` CONVENTIONS.md §Release Archive Standard.
- **darwin code-signature identifier** is now the canonical `jstats`
  (was `jstats-darwin-arm64`), set via `codesign -i` so it stays stable
  after the archived binary is renamed to its canonical name.

No change to the binary's behaviour — a packaging / build-config release.

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
