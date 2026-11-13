# Tasks

## 1. List navigation

- [x] 1.1 Add `g`/`G` first/last and `Ctrl+F`/`Ctrl+B` page keys to the main
  navigation list, reusing the existing jump and page paths; verify with a
  focused test in `internal/tui/viewport_test.go` covering each key and the
  resulting cursor and window position.
- [x] 1.2 Add a filter regression test proving `g` and `G` are appended to the
  query and do not trigger navigation; verify with
  `go test ./internal/tui -run Filter`.

## 2. Detail navigation

- [x] 2.1 Handle `g`, `G`, `Home`, `End`, `Ctrl+F`, and `Ctrl+B` in
  `handleDetailKey` using the viewport jump and page methods; verify with a
  test in `internal/tui/detail_test.go` asserting top, bottom, and page
  offsets.

## 3. Hints and documentation

- [x] 3.1 Update the list and detail footers and the README key table with the
  new keys; verify the footer tests and that the README documents every new
  key.

## 4. Integration check

- [x] 4.1 Run `go test ./...` and `bash scripts/coverage.sh`; verify overall
  coverage is at least 70% and core package coverage at least 80%.

## Workflow follow-up

- Add an `### Added` bullet under `## [Unreleased]` in `CHANGELOG.md`.
- Archive the change and commit it as Pim Snel without self-promotion.
- Close bean `skull2-2p9w` with a summary of the change.
