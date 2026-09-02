## Context

Three pieces of existing machinery make this change small.

**Selection already lives in the backing store.** `toggleSelect`
(`internal/tui/update.go:797`) writes `Selected` into
`m.reposByProvider[provider]`, and `selectedItems()` (`internal/tui/model.go:739`)
walks every configured provider. A selection therefore already survives query
changes, facet changes, sort changes and level changes, and can already span
forges. The narrow → pick → re-narrow → pick loop needs no new state — only a
key that works while the input is focused.

**The filter is a two-state machine, and only one state has the conflict.**
`m.filtering == true` is filter-edit mode: keys are fed to the textinput
(`update.go:303-311`), so `space` is a query separator. `m.filtering == false &&
m.filter.Value() != ""` is a narrowed list under the normal keymap, where
`space` already toggles selection. The gap is that filter-edit mode has only two
exits: Enter (blur, keep query, **and drill**) and Esc (blur, **clear** query).
`Tab` and `shift+Tab` are unbound anywhere in the TUI.

**Facets are a proven narrow-only pattern.** `facets` (`internal/tui/filters.go`)
holds three independent narrowing filters, applied before the fuzzy query
(`model.go:546-553`) and rendered by the existing `facet_status` chrome element.
A fourth one is additive.

The synthetic `allReposLabel` row (`model.go:608`) is the precedent for "a row at
the provider level that opens a virtual flat repository list", and the clone
progress popup (`view.go:100`, `lipgloss.Place` centered over the dimmed list) is
the precedent for a modal overlay with its own key gate at the top of
`handleKey`.

## Goals / Non-Goals

**Goals:**
- Make hand-curating a cross-forge collection fast: filter, pick with `Tab`,
  refilter, pick, review the basket, commit.
- Persist collections in a plain, hand-editable file the user can later feed into
  Home Manager.
- Full CRUD from inside the TUI: create, add members, remove members, rename,
  delete.
- Let a collection scope the multiplex jump list.

**Non-Goals:**
- **Automatic / predicate-defined collections** (language, forge topics,
  archived, forked). Language and topics live in the lazily-fetched tier-2
  `provider.Details`, so they are unknown for repositories whose details were
  never fetched; a taxonomy built on them would be silently incomplete. Split to
  its own bean.
- **Home Manager layering.** v1 writes and reads exactly one file. A declared
  (read-only) layer plus merge/tombstone semantics and an `export` subcommand are
  deliberately deferred; the file format is chosen to make that additive.
- **Collection-as-tmux-session** (one window per member, clone-all-first). A
  collection scopes the existing multiplex flow; it does not gain a second
  command template.
- No `hup collections` CLI subcommand in this change.
- No change to `config.yaml`.

## Decisions

**`Tab` toggles in place of a mode exit.** Rejected alternatives: (a) a
"commit filter" key that blurs the input while keeping the query, then `space`
works — this keeps a mode transition in the middle of the picking loop and
costs a keystroke per query; (b) changing the AND separator away from
whitespace — breaks the fzf-style query language and the `search_strategy`
work; (c) `ctrl+space` — many terminals cannot send it. `Tab`/`shift+Tab`
mirrors `fzf --multi` exactly (toggle + move down / up), so it is muscle memory,
and it leaves `space` a pure filter character. The same binding is added in list
mode for consistency (`space` toggles in place, `Tab` toggles and advances).

**`Tab` never changes the query or resets the window.** In filter-edit mode the
cursor/offset reset is tied to the query text changing (`update.go:307-310`);
`Tab` is intercepted before `m.filter.Update`, so no tab character is inserted
and the window is preserved. It is inert at the provider/owner levels and in
multiplex mode, matching `toggleSelect`.

**The basket is a facet, not a view.** `facets.selected triState` reuses
`cycle`/`label`/status rendering wholesale, composes with the other facets and
the fuzzy query for free, and lets the user un-pick a mistake directly in the
basket list (`Tab`/`space` still toggle). Because `Selected` lives on `repoItem`
rather than `provider.Repo`, `facets.match` gains a second parameter
(`match(r provider.Repo, selected bool)`) — the one signature change.

**The selection count is honest about scope.** `selected:only` narrows the
*current* list, so in a per-owner list it shows only that owner's picks, while
the count in the position indicator is always the global selection total (e.g.
`41-60 of 213 · 4 selected`). The mismatch is the correct signal that the full
basket is visible in the combined flat view. Rejected: auto-switching to the flat
view when the facet is enabled (too magic). The count goes into `indicatorText`
rather than a new chrome element, so `config.KnownElements` — and with it the
config-example gate — is untouched.

**Keys.** `x` cycles the selected facet (`[x]` is the checkbox glyph; `f`/`a`/`v`
are taken by the other facets and `s`/`S` by sorting). `A` opens
add-to-collection (`a` is the archived facet). In a collection's repository list
`d` drops the highlighted or selected members from that collection; at the
collections level `d` deletes the collection (behind a confirmation modal) and
`R` renames it (`r` is refresh).

**Store shape: lists, not maps.** yaml.v3 marshals Go maps sorted by key, which
would churn the file and lose curated order. A list of collections, each with a
list of members, keeps the file diff-stable and insertion-ordered — which also
leaves the door open to using member order as tmux window order later. A rename
mutates one `name` field in place instead of moving a map key.

```yaml
version: 1
collections:
  - name: terraform-modules
    members:
      - provider: github-mipmip   # providers[].name, the unique config key
        owner: org1
        repo: terraform-aws
      - provider: codeberg-pim
        owner: pim
        repo: tf-modules-dns
```

**Members key on the provider `name`, not `short`.** `short` is a display prefix
(`gh`) and is not guaranteed unique across configured providers, while
`providers[].name` is the unique key already used for cache scoping. Owner and
repository names are stored as fetched but compared case-insensitively, since
forges differ in case handling.

**Dangling members are kept.** A member that resolves to no cached repository is
rendered dim and counted, not dropped. Silent dropping would let a curated list
erode whenever an owner's cache is missing or a fetch has not happened yet, which
is exactly the failure a curated list is supposed to protect against.

**Adds are idempotent, and commit clears the selection.** Adding a repository
already in the collection is a no-op (reported in the status line as
"n already members"), mirroring bulk clone, which clears `Selected` on success
(`update.go:65-70`).

**Rename refuses collisions.** A rename (or a create) onto an existing name is
rejected with an error in the modal rather than merging the two collections;
merge is a destructive interpretation of a typo. `--collection` also references
names, so uniqueness is a load-time invariant.

**Atomic writes, reusing the cache pattern.** Marshal, write to a temp file in
the same directory, sync, close, rename (`internal/cache/details.go:70-108`). A
missing file loads as an empty store; the directory is created on first write.

## Risks / Trade-offs

- [A curated collection silently loses members when a provider is renamed in
  `config.yaml`] → Members become dangling rather than deleted, stay in the file,
  and render dim, so the user can fix the config or re-add them. Documented in
  the spec as a scenario.
- [`selected:only` in a scoped list looks like the basket but is not] → The
  global count in the position indicator always shows the true total, and the
  spec pins that behavior.
- [`Tab` in filter-edit mode could be captured by the textinput or by terminal
  completion habits] → It is intercepted in the `m.filtering` switch before
  `m.filter.Update(msg)`, and unit-tested for "query unchanged, cursor advanced,
  row selected".
- [Deferring Home Manager layering means the file may later need to move or gain
  a wrapper] → `version: 1` plus the list-shaped schema makes a declared layer
  additive (a second file merged under the writable one), so v1 files stay valid.
- [Five keys added to an already dense footer] → The collections keys (`d`, `R`)
  only apply at the collections level and inside a collection, so the footer is
  level-dependent, as it already is for the detail view.
