## Why

Multiplex mode is a session-picker: launch, pick a repo, and it clones-if-needed
then runs the `switch_command`. Today launching drops the user into passive
list-navigation at the provider level, so every session start costs an extra `/`
keypress (and a drill into the flat list) before they can type a repo name. A
filter-first launch turns `hup tui` into an `fzf`-style picker: launch, type,
Enter.

## What Changes

- Add a `--filter` launch flag to `hup tui`. When set, the TUI boots directly
  into the combined cross-provider flat repository list with the fuzzy filter
  already active and focused, so the first keystroke narrows the repo list.
- `--filter` implies the flat combined view (the fuzzy filter only means
  "search repositories" at the repos level; at the provider/owner levels it
  would narrow the wrong thing). Passing `--filter` alone is equivalent to
  passing `--flatlist --filter`.
- `--filter` composes with `--mode multiplex`: launch filter-first, type,
  Enter clones-if-needed and switches — the full picker flow with no extra
  keypresses.
- Esc semantics are unchanged and already correct as a two-stage unwind: the
  first Esc clears the query and stops filtering (landing in the flat list),
  the second Esc goes one level up (the flat view backs out to the provider
  list). This behavior is documented, not modified.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tui`: the launch requirement gains a filter-active launch option
  (`--filter`) that opens the combined flat list with the fuzzy filter active,
  composing with the existing flat-list and mode-selection launch behavior.

## Impact

- `cmd/hup/main.go`: new `--filter` bool flag on the `tui` flag set, threaded
  into `tui.Run`.
- `internal/tui/run.go`: `Run` gains a `filter bool` parameter; when set it
  starts the model flat (reusing `startFlat`) and activates the filter input
  (`filtering = true`, `filter.Focus()`).
- No config schema change, no new dependencies. Existing Esc/Enter key handling
  and the mode chrome (which already renders the `filter` element while
  filtering) are reused unchanged.
