## Context

The TUI already has all the pieces this change needs; it only lacks a way to
combine them at launch. `Run(cfg, mode, flatList)` (`internal/tui/run.go`)
constructs the model, optionally applies `--mode`, and optionally calls
`startFlat()` (`internal/tui/model.go:610`) to open the combined cross-provider
flat repo list. The fuzzy filter is toggled at runtime by the `/` key handler,
which does exactly two things (`internal/tui/update.go:361-363`):

```go
m.filtering = true
m.filter.Focus()
```

The mode chrome renders the `filter` element whenever `m.filtering ||
m.filter.Value() != ""` (`internal/tui/mode.go:57`), so an at-launch filter draws
itself with no chrome change. Esc handling is already a two-stage unwind: while
filtering, Esc clears the query and stops filtering (`update.go:282-287`); in the
flat list, Esc → `goBack` backs `flatAll` out to the provider level
(`update.go:506-513`).

## Goals / Non-Goals

**Goals:**
- Boot `hup tui` filter-first with a single `--filter` flag, reusing existing
  flat-list and filter machinery.
- Make the flow `launch → type → Enter` a complete picker, especially under
  `--mode multiplex`.

**Non-Goals:**
- No config-schema toggle and no per-mode "start filtering" setting — the launch
  flag was the chosen home (a mode-intrinsic behavior was considered and
  rejected as too implicit).
- No change to Esc/Enter key semantics — the desired two-stage Esc is already the
  implemented behavior.
- No filtering at the provider or owner levels.

## Decisions

**Launch flag, not mode-intrinsic or config toggle.** `--filter` is ephemeral and
explicit; the user opts in per launch. Alternatives: (a) make multiplex always
filter-first — rejected as too implicit and unavailable to management mode; (b) a
`ModeConfig.start_filtering` field — rejected as more surface area than the PoC
needs.

**`--filter` implies flat.** The fuzzy filter is level-aware and only searches
repositories at the repos level; at the provider/owner levels it narrows the
wrong list, and every drill-in `clearFilter()`s. So `--filter` turns on the flat
combined view regardless of `--flatlist`, rather than filtering a provider list
that no one wants to search. Passing both flags is harmless (idempotent).

**Thread a `filter bool` through `Run`.** Add the parameter to `Run` and a
`--filter` bool to the `tui` flag set in `cmd/hup/main.go`. When set, `Run` calls
`startFlat()` (so `--filter` alone still lands flat) and then activates the filter
input with the same two lines the `/` handler uses. Keeping the activation in
`Run` (not a new model constructor path) mirrors how `flatList` is already
handled and keeps `New` unchanged.

**Reuse, don't re-document, Esc.** The spec's Esc scenario pins the existing
behavior as the contract so a future refactor can't silently break the picker's
unwind; no code implements anything new for it.

## Risks / Trade-offs

- [A future refactor changes Esc-in-filter to not clear the query] → The new spec
  scenario ("first Esc stops filtering, second Esc goes one level up") locks the
  two-stage unwind as a tested contract.
- [`filter` element missing from a custom mode's chrome makes the active input
  invisible] → Out of scope here (the built-in management and default multiplex
  chrome both include `filter`); the input still captures keys, and Esc exits.
  Noted as a known limitation rather than adding validation in this change.
- [Adding a parameter to `Run` touches its one caller] → Trivial; `cmd/hup` is
  the sole caller and is updated in the same change.
