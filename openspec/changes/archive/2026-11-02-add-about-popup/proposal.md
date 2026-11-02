# Proposal

## Why

The TUI gives no way to see which version is running, where the project lives,
or under which license it is distributed. Users must quit to `hip version`, and
release builds can differ from the embedded fallback version, so the in-app
value has to come from the same resolved version the CLI prints.

## What Changes

- Add a centered about popup, opened and closed with `?`; `esc` also closes it.
  While open it swallows all other keys.
- It is available at every navigation level and TUI mode, except while the
  filter is active or another modal owns the screen. While filtering, `?` stays
  a literal query character.
- The popup shows:
  - version: the resolved binary version, not just the embedded `VERSION` value
  - build: the Go runtime version and `GOOS/GOARCH`
  - license: MIT and the copyright line
  - github: `github.com/mipmip/hip`
  - credits: `Pim Snel, et al`
- Pass the resolved version from `cmd/hip` into the TUI so release builds built
  with `-X main.version` show the release tag.
- Add `?: about` to the built-in footer hints, including the detail view and
  multiplex mode.
- Document the key in the README key table.

No commit hash, build date, full license text, or clickable URL is added.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tui`: add a requirement for an in-app about popup and its key behavior.

## Impact

- `internal/tui`: model state, key handling, view, version injection, footer
  hints, tests.
- `cmd/hip`: pass `resolveVersion()` into `tui.Run`.
- `README.md`: key table entry.
- No new dependencies, configuration keys, CLI flags, provider APIs, or build
  pipeline changes.
