---
# skull2-289n
title: user-defined repo collections (workspaces)
status: completed
type: epic
priority: normal
created_at: 2026-08-18T08:31:19Z
updated_at: 2026-09-02T16:00:56Z
parent: skull2-ok4c
---

My brain doesn't always think in forge organizations. A project or personal
topic has repos spread over multiple organizations and sometimes multiple
forges. I want hand-curated, named collections of repos — e.g. several
terraform modules living in different orgs and on different forges, browsable
as one list called `terraform-modules`.

The core of this bean is **making the selection cheap**: multiselect a bunch of
repos and add them to a collection, either an existing one or a new one created
on the spot. Everything else (browsing, renaming, launching into a collection)
follows from that.

Automatic/predicate collections (language, forge topics, archived, forked) were
originally mixed into this bean and are now split out into `skull2-ewgn`.
Symlink-based collections were considered and dropped: collections live in the
TUI only, and get their effect by scoping the multiplex jump list.

## Scope

- Storage: `~/.config/huphop/collections.yaml`, writable, hand-editable.
  Members are provider name + owner + repo, so a collection spans forges.
  Member order = insertion order. Home Manager layering (declared read-only
  layer + upstream export for my `up_home` task) is deliberately deferred.
- Selection must work while filtering. `space` is a filter term separator, so
  it cannot toggle selection in filter-edit mode — `Tab`/`shift+Tab` do it
  fzf-style (toggle + move), leaving the query untouched. Selections already
  survive re-filtering because they live on the backing repo set.
- A "basket" facet (`selected:only`) to review what has been picked across
  several different queries, plus a selection count in the position indicator.
- CRUD from inside the TUI: create, add, remove members, rename, delete.
- `--collection <name>` launch flag, composing with `--mode multiplex` and
  `--filter`: a collection becomes a scoped tmux jump list.

## Checklist

- [x] collection store package (`internal/collections`) + coverage gate entry
- [x] `Tab`/`shift+Tab` selection while filtering
- [x] `selected` basket facet + selection count
- [x] add-to-collection modal (pick existing / create new)
- [x] collections navigation level + remove/rename/delete
- [x] `--collection` launch flag
- [x] README documentation

## OpenSpec change

Captured as `2026-09-02-add-repo-collections`
(`openspec/changes/2026-09-02-add-repo-collections/`) — new `collections`
capability (store file, cross-provider member identity, CRUD, dangling members
preserved) plus `tui` deltas (select-while-filtering, basket facet, selection
count, add-to-collection modal, collections nav level and management,
`--collection` launch). No `config.yaml` schema change, so the config-example
gate is untouched. Validated. Ship with: `/ship 2026-09-02-add-repo-collections`.


## Summary of Changes

Shipped as `2026-09-02-add-repo-collections` (commit 11b798a). New
`internal/collections` package (91% covered): `Store`/`Collection`/`Member` as
ordered lists so `~/.config/huphop/collections.yaml` stays diff-stable and
insertion-ordered; `Load` (missing file -> empty store without creating it,
malformed or duplicate-name file -> error naming the path), atomic `Save`,
provider-exact + owner/repo case-insensitive member identity, and
Create/Add/Remove/Rename/Delete with case-insensitive unique names. Every TUI
mutation snapshots via `Store.Clone()` and rolls back when the write fails, so
the in-memory store never drifts from the file.

TUI: `Tab`/`shift+Tab` toggle-and-move selection intercepted before the filter
input sees them (query and scroll window untouched), `space` unchanged; a
`selected` tri-state basket facet on `x`; the global selection count folded into
`indicatorText` (no new chrome element, so the config-example gate is untouched);
`A` opens the add-to-collection modal (existing collections with counts plus a
create-new row); a `Collections` synthetic provider row -> `levelCollections` ->
a collection's cross-provider repo list; `d` removes members, `R` renames, `d` at
the collections level deletes behind a y/n confirm; `--collection <name>` launch
flag composing with `--mode` and `--filter`. Unresolved members (uncached repo,
or a provider no longer in config.yaml) are kept, rendered dim as
`(unresolved)`, guarded against in clone/open/detail/prefetch/select, and
removable via the highlight fallback in `d`.

No `config.yaml` schema change and no new dependencies. `internal/collections`
added to the coverage gate's core packages. Overall coverage 82.3%.

Follow-ups split out: `skull2-ewgn` (automatic predicate-based collections) and
`skull2-u63k` (expose the active collection in the multiplex `switch_command`).
