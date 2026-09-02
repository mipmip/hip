## Why

The on-disk layout and the TUI hierarchy are both organized by forge and owner,
but real work is not. A set of terraform modules lives in several
organizations and across forges; a client project spans a GitHub org and a
Codeberg account. There is no way to say "these twelve repositories are one
thing" and come back to that grouping tomorrow.

The value is in *hand-curated* groupings — the ones the user assembles
themselves and can trust. Automatic taxonomies (language, forge topics) are a
separate, later idea and are explicitly out of scope here.

Building such a grouping is a selection problem, and selection is where the TUI
currently blocks: the fuzzy filter is the only practical way to find repos
scattered across owners, but while the filter input is focused `space` is a
query separator, so nothing can be multi-selected. Leaving filter-edit mode is
no help either — Enter blurs the input but also drills into the highlighted
repo, and Esc blurs it but wipes the query. There is no way to narrow, pick,
re-narrow and pick again.

## What Changes

- **A collection store.** A new writable file `~/.config/huphop/collections.yaml`
  holds named, user-defined collections of repositories. Members are identified
  by provider name + owner + repository name, so a collection spans owners,
  organizations and forges. Member order is preserved (insertion order).
- **Select while filtering.** `Tab` toggles the highlighted repository's
  selection and moves down; `shift+Tab` toggles and moves up. Both work while
  the filter input is focused, leaving the query untouched, so the loop
  "type → Tab Tab → retype → Tab" works. `space` remains a filter character and
  keeps its existing toggle behavior in list mode.
- **A selection basket.** A fourth tri-state facet, `selected`, narrows the list
  to (or hides) the current selection, so the user can review what they have
  picked after the queries that picked it are gone. The position indicator gains
  a selection count.
- **Add to collection.** With a selection active, a new action opens a modal that
  lists existing collections with member counts plus a "new collection" entry
  that prompts for a name. Committing adds the selected repositories to that
  collection (already-present members are not duplicated) and clears the
  selection.
- **Browse and manage collections.** A synthetic `Collections` row at the
  provider level opens a collections list; entering a collection shows its
  members as a flat cross-provider repository list. From there members can be
  removed; from the collections list a collection can be renamed or deleted
  (with confirmation).
- **Launch into a collection.** A `--collection <name>` launch option opens
  directly in that collection's repository list, composing with `--mode
  multiplex` and `--filter` so a collection doubles as a scoped tmux jump list.
- Members that no longer resolve against the cache (renamed, deleted, or
  never-fetched repositories) are kept in the file and rendered dim rather than
  silently dropped.

## Capabilities

### New Capabilities
- `collections`: the persisted store of named, user-defined repository
  collections — file location and format, cross-provider member identity,
  insertion-ordered membership, atomic writes, and the create/add/remove/rename/
  delete operations with name uniqueness and dangling-member preservation.

### Modified Capabilities
- `tui`: gains selection-while-filtering keys, a `selected` facet, a
  selection count in the position indicator, the add-to-collection action and
  modal, a collections navigation level with member/collection management, and a
  `--collection` launch option. The facet-filter, position-indicator and
  multiplex-hides-multiselect requirements are amended.

## Impact

- New `internal/collections/`: store type, load/save (atomic write), and the
  CRUD operations. Added to `CORE_PKGS` in `scripts/coverage.sh` (≥80%).
- `internal/tui/filters.go`: `selected` tri-state facet; `facets.match` gains the
  selection flag as a parameter (`repoItem.Selected` lives on the item, not on
  `provider.Repo`).
- `internal/tui/update.go`: `Tab`/`shift+Tab` handling in both filter-edit mode
  and list mode; new keys for the facet, the add-to-collection action, and
  member/collection management; a modal key gate alongside the existing clone
  popup gate.
- `internal/tui/model.go`: `levelCollections`, the `Collections` synthetic row,
  collection scoping of the repo list, and the loaded store on the model.
- `internal/tui/view.go`: collection rows, the picker/confirm modals (reusing the
  existing centered-overlay approach), dim rendering for dangling members, and
  the selection count in `indicatorText`.
- `internal/tui/run.go` + `cmd/hup/main.go`: `--collection` launch option.
- **No `config.yaml` schema change** — collections live in their own file, so the
  config-example gate is untouched. No new dependencies (`gopkg.in/yaml.v3` is
  already a dependency).
