## Why

`add-repo-collections` gives the multiplex mode a collection-scoped jump list
(`hup tui --collection terraform-modules --mode multiplex --filter`), but the
mode's `switch_command` cannot see which collection the user is in. The template
context is the clone-path fields plus `Target` (`switchData` in
`internal/tui/multiplex.go`), so the obvious tmux layout for a workspace — one
session per collection, one window per repository — is unexpressible:

```yaml
switch_command: >-
  tmux-helper open-or-new-session-with-name="{{.Collection}}"
  open-or-new-window-with-name="{{.Repo}}"
```

`{{.Collection}}` currently fails at execute time ("can't evaluate field
Collection in type tui.switchData"), which aborts the switch with an error. That
is the correct failure, but it means the collection feature stops exactly where
it becomes useful: a curated group of repositories that cannot name its own
workspace is only half a workspace.

## What Changes

- Add a `Collection` field to the `switch_command` template context, holding the
  name of the collection whose repository list is active — set both when the TUI
  was launched with `--collection` and when the user navigated into a collection
  in-app.
- Outside a collection (the combined flat view, or an owner drill-down) the field
  renders as the empty string. No fallback value is invented: users who want a
  different session name per scope express it with a template conditional, e.g.
  `{{if .Collection}}{{.Collection}}{{else}}{{.Short}}->{{.OwnerLower}}{{end}}`.
- Rendering stays shell-safe: the collection name is substituted before the
  shell-words split, so a name with spaces or metacharacters remains one literal
  argument.

## Capabilities

### Modified Capabilities
- `tui`: the switch_command rendering requirement gains the active collection
  in the template context, with defined behavior (empty) outside a collection.

## Impact

- `internal/tui/multiplex.go`: `switchData` gains `Collection string`;
  `renderSwitchCommand` populates it from `m.selColl`.
- Tests in `internal/tui`: rendering inside and outside a collection, and a
  collection name carrying a space staying one argument.
- No config schema change (`switch_command` is already a free-form template), no
  new dependencies, no change to the clone-path template.
- **Sequencing**: depends on `2026-09-02-add-repo-collections` (which introduces
  `m.selColl`); ship after it.
