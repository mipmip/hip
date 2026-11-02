# Tasks

## 1. Version plumbing

- [x] 1.1 Add a `version` field to `Model`; change `tui.New` and `tui.Run` to take the resolved version, pass `resolveVersion()` from `cmd/hip`, and update the smoke test call site. Verify `go test ./internal/tui ./cmd/hip` passes.

## 2. About modal

- [x] 2.1 Add `about` state and `aboutView`, rendering the resolved version, `runtime.Version()` plus `GOOS/GOARCH`, MIT with the copyright line, the GitHub URL, and `Pim Snel, et al`; show it from `View` after the existing modals. Verify `TestAboutPopupRenders` covers every field and the title.
- [x] 2.2 Route `?` globally in `handleKey`, close the popup on `?` or `esc`, swallow other keys while open, and keep `?` literal while filtering. Verify tests cover opening and closing from the list and detail views, key swallowing, and the filter case.
- [x] 2.3 Add `?: about` to the default, detail, and `switch_hint` footers, and add the key to the README key table. Verify the footer tests assert the hint and the README row is present.

## 3. Integration

- [x] 3.1 Run `bash scripts/coverage.sh` and verify the full suite passes with overall coverage at least 70% and core logic at least 80%.

## Workflow follow-up

- Add an `### Added` bullet for the about popup under `## [Unreleased]` in `CHANGELOG.md`.
- Archive the change after review.
- Commit as Pim Snel without self-promotion.
