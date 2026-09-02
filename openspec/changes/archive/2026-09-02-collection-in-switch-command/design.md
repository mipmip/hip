## Context

`renderSwitchCommand` (`internal/tui/multiplex.go`) builds its context as:

```go
type switchData struct {
	clonepath.Data
	Target string
}
```

`clonepath.Data` carries `BaseDir, Provider, Type, Short, Host, Owner,
OwnerLower, Repo, RepoLower`; `add-repo-collections` added `m.selColl`, the name
of the collection whose members are listed at the repos level (empty otherwise),
set by both the `--collection` launch option and in-app navigation into a
collection.

Go's `text/template` reports an unknown struct field as an execute error
regardless of the `missingkey` option (which only governs maps), so
`{{.Collection}}` fails loudly today and no existing configuration can depend on
it. Adding the field is therefore purely additive.

## Goals / Non-Goals

**Goals:**
- Let a `switch_command` name the active collection, so one tmux session per
  collection with one window per repository is expressible in configuration.
- Keep the shell-injection guarantee: the name is template-substituted before
  `splitArgs`, never handed to a shell.

**Non-Goals:**
- **No `Collection` in the clone-path template.** A repository may belong to
  several collections; if the clone path could vary by collection, the same
  repository would clone to two different directories and `sync` would lose track
  of which is canonical. The on-disk layout must stay a pure function of
  provider/owner/repo.
- **No per-collection `switch_command`.** One template per mode; a user who wants
  branching writes it in the template.
- **No collection-as-session action** (open a session with a window per member,
  cloning everything missing first). That is a separate, larger change with its
  own progress UI and re-open semantics.
- **No invented fallback** for "not in a collection".

## Decisions

**Empty outside a collection, with a documented conditional.** The alternatives
were a silent fallback (the owner, or a literal like `all`) and refusing to run
the command when the template references `Collection` outside a collection. Both
hide the rule from the user: the first produces a session name whose origin is
unclear, the second turns a working keybinding into an error depending on where
the cursor happens to be. An empty string plus `{{if .Collection}}…{{else}}…{{end}}`
puts the choice in the user's own config, where it is visible and versioned.

**Populate from `m.selColl`, not from the launch flag.** The launch option is one
of two ways into a collection; reading the model's current scope covers in-app
navigation too, and keeps the field meaning exactly "the collection you are
looking at".

**Field on `switchData`, not on `clonepath.Data`.** `clonepath.Data` is shared
with the clone-path engine and the syncer, where a collection has no meaning (see
the non-goal above). Keeping `Collection` on the switch-only struct means the
navigation context cannot leak into on-disk layout.

## Risks / Trade-offs

- [A user references `{{.Collection}}` and gets an empty session name outside a
  collection] → Documented in the spec as a scenario and in the README with the
  conditional form. The failure is visible (tmux gets an empty name) rather than
  silent, and the conditional is a one-line fix.
- [A collection name containing spaces or quotes breaks the rendered command] →
  The existing `splitArgs` contract already covers this; a test pins a
  space-carrying collection name as a single argument.
