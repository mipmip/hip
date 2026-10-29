# HIP

**Hub Is Plural.** HIP is a multi-provider git portfolio manager: a uniform
on-disk layout for all your git repos across GitHub, Codeberg/Forgejo and
GitLab, a TUI to browse and clone them, and a cron-friendly CLI (`hip`) to keep
backup clones in sync.

> Status: PoC / alpha base. See [BRIEFING.md](./BRIEFING.md) for the full
> design and [CLAUDE.md](./CLAUDE.md) for the build guide. The roadmap lives in
> `beans` (`beans list`).

## Quick start

```bash
nix develop            # dev shell
nix build              # build
nix run . -- version   # run
nix flake check        # build + tests
```

## What it does

- **Uniform layout** via a configurable clone-path template, e.g. GitHub
  `TechNative-B-V/foo` → `~/gh.technative-b-v/foo`.
- **TUI**: browse provider → owner → repos with fuzzy search; single/bulk
  clone; open in browser.
- **CLI sync**: clone missing repos and fast-forward-pull existing ones;
  dirty-tree safe, cron-friendly.
- **Collections**: hand-curated, named groups of repos that span owners,
  organizations and forges.

Configuration lives in `~/.config/hip/config.yaml`; cached repo metadata in
`~/.cache/hip/`.

## Collections

A collection is a named group of repositories you assemble yourself, say the
terraform modules you maintain across three organizations and two forges. It is
a view, never storage: adding or removing members never touches a clone on disk.

Build one from the TUI:

| Key         | Where                | What it does                            |
|-------------|----------------------|-----------------------------------------|
| `/`         | repo list            | filter (fzf-style query)                |
| `tab`       | repo list, filtering | select the row and move down            |
| `shift+tab` | repo list, filtering | select the row and move up              |
| `space`     | repo list            | select the row in place                 |
| `x`         | repo list            | cycle the selection facet (see below)   |
| `A`         | repo list            | add the selection to a collection       |
| `d`         | inside a collection  | remove the selected/highlighted members |
| `R`         | collections list     | rename a collection                     |
| `d`         | collections list     | delete a collection (asks first)        |

`space` is a term separator in the filter query, so selection has its own keys:
`tab` and `shift+tab` work while you are still typing, and the query is left
untouched. Selections live on the repository set rather than the visible list,
so they survive re-filtering: narrow, pick, retype, pick again. Because the
repos you picked scroll out of view as the query changes, `x` cycles a
`selected` facet (`all` → `only` → `hide`) that shows just what is in your
basket; the position indicator also reports how many are selected in total.

`A` opens a picker listing the existing collections with their member counts,
plus a `+ new collection` entry that asks for a name. Browse collections from
the `Collections` entry in the provider list.

Collections are stored in `~/.config/hip/collections.yaml`, beside
`config.yaml` but writable: the TUI edits it and rewrites it atomically:

```yaml
version: 1
collections:
  - name: terraform-modules
    members:
      - provider: github-mipmip   # the providers[].name key from config.yaml
        owner: org1
        repo: terraform-aws
      - provider: codeberg-pim
        owner: pim
        repo: tf-modules-dns
```

Members are kept in the order you added them, and a member that matches no
cached repository (renamed upstream, owner not fetched yet, provider renamed in
`config.yaml`) is kept and shown dim as `(unresolved)` rather than being
dropped.

Launch straight into one:

```bash
hip tui --collection terraform-modules                  # scoped repo list
hip tui --collection terraform-modules --mode multiplex --filter
```

The second form is the tmux jump list scoped to a collection: type, press Enter,
and the repo is cloned if needed before the mode's `switch_command` runs.

The active collection is available to that command as `{{.Collection}}`, so a
collection can name its own tmux session while each repo gets a window:

```yaml
modes:
  multiplex:
    switch_command: >-
      tmux-helper open-or-new-session-with-name="{{.Collection}}"
      open-or-new-window-with-name="{{.Repo}}" layout-template="3-pane-coding"
```

Outside a collection (the combined view, or an owner drill-down)
`{{.Collection}}` is empty; no substitute is invented for you. If the same
keybinding should also work there, say so in the template:

```yaml
      open-or-new-session-with-name="{{if .Collection}}{{.Collection}}{{else}}{{.Short}}->{{.OwnerLower}}{{end}}"
```

The clone-path template deliberately does *not* get `{{.Collection}}`: a repo can
belong to several collections, and a path that varied by collection would clone
it twice and leave `sync` without a canonical copy.

Not yet implemented: automatic collections computed from language or forge
topics, and a Home Manager-style declared (read-only) layer beneath the
writable file.
