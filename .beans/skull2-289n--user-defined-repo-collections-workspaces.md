---
# skull2-289n
title: user-defined repo collections (workspaces)
status: todo
type: epic
priority: normal
created_at: 2026-08-18T08:31:19Z
updated_at: 2026-09-02T15:08:43Z
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
