---
# skull2-u63k
title: expose active collection in multiplex switch_command
status: todo
type: feature
created_at: 2026-09-02T15:58:43Z
updated_at: 2026-09-02T15:58:43Z
parent: skull2-ok4c
---

The multiplex `switch_command` template cannot see which collection is active,
so "one tmux session per collection, one window per repo" is unexpressible —
the payoff of `--collection` is missing.

Wanted config:

```yaml
modes:
  multiplex:
    switch_command: >-
      tmux-helper open-or-new-session-with-name="{{.Collection}}"
      open-or-new-window-with-name="{{.Repo}}" layout-template="3-pane-coding"
```

`{{.Collection}}` today fails at template-execute time (the field does not
exist on `switchData`), which aborts the switch with an error.

## Decisions

- Add `Collection string` to `switchData` only — NOT to `clonepath.Data`. A repo
  can be in several collections; if the clone path varied by collection the same
  repo would clone twice and `sync` would lose the canonical copy.
- Populated from `m.selColl`, so it works both for `--collection` and for
  navigating into a collection in-app.
- Empty outside a collection, with no invented fallback. A different name per
  scope is a template conditional:
  `{{if .Collection}}{{.Collection}}{{else}}{{.Short}}->{{.OwnerLower}}{{end}}`.

Out of scope (own bean if wanted): a different `switch_command` per collection,
and collection-as-session (open a session with a window per member, cloning
everything missing first).

## Checklist

- [x] `Collection` field on `switchData`, populated from `m.selColl`
- [x] tests: inside a collection, outside (empty + conditional), space in the name
- [x] README: session-per-collection example + conditional fallback

## OpenSpec change

Captured as `2026-09-02-collection-in-switch-command`
(`openspec/changes/2026-09-02-collection-in-switch-command/`) — MODIFIED the tui
`switch_command` rendering requirement to include the active collection in the
template context. Validated. Depends on `2026-09-02-add-repo-collections`
(introduces `m.selColl`) — ship after it:
`/ship 2026-09-02-collection-in-switch-command`.
