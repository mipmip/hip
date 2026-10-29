## Context

Third wholesale rename (golgotha → huphop → hip), following the playbook of
the archived `2026-08-18-rename-to-huphop` change. Unlike that one, the tool is
released (v1.4.0) and used daily, so existing installs have config, tokens and
cache under the old names. The GitHub repo is already `mipmip/hip`; only the
local remote needs repointing.

## Goals / Non-Goals

**Goals:** complete, gate-green rebrand huphop→hip / hup→hip with no stray
references in code, tooling or active docs; released as v2.0.0 with clear
migration steps.

**Non-Goals:** behavior changes; a fallback to or auto-migration of
`~/.config/huphop`, `~/.cache/huphop` or `HUPHOP_*`; touching archived changes;
renaming `.beans/` files or the `skull2-` bean prefix; renaming the local
checkout directory (the user does that outside the change).

## Decisions

- **Name and brand:** product `HIP`, expanded as "Hub Is Plural" in README and
  `--help`/description; command `hip`.
- **Rename map:** module `github.com/mipmip/huphop`→`github.com/mipmip/hip`;
  dir `cmd/hup`→`cmd/hip`; binary/CLI `hup`→`hip`; env `HUPHOP_`→`HIP_`;
  config `~/.config/huphop`→`~/.config/hip`; cache `~/.cache/huphop`→
  `~/.cache/hip`; root package `huphop`→`hip`; goreleaser `project_name` and
  ids `hip`; flake `pname`/`mainProgram` `hip`, coverage check `hip-coverage`.
- **Hard cut:** the old paths and env vars are not read. Rationale: a single
  daily user who can migrate by hand in a minute; a fallback adds code and
  tests that must be removed later.
- **Version:** the next release is `2.0.0` because the change breaks config
  and env for v1 users. `VERSION` stays `1.4.0` in this change:
  `scripts/release.sh` bumps it and promotes the `[Unreleased]` CHANGELOG
  section. The module path changes anyway, so no `/v2` suffix is needed.
- **CHANGELOG migration block:**
  `mv ~/.config/huphop ~/.config/hip`, rename `HUPHOP_*_TOKEN` to
  `HIP_*_TOKEN` in shell profile and cron, replace `hup` with `hip` in cron
  entries and aliases, optionally `rm -rf ~/.cache/huphop`.
- **Remote:** repoint `origin` to `git@github.com:mipmip/hip.git` first so the
  ship push lands.
- **vendorHash:** keep; the rename adds no dependencies. Recompute only if nix
  reports a mismatch.
- **Spec deltas:** `configuration`, `collections`, `repo-cache`,
  `provider-clients` carry the path and env changes. Command-name references
  (`hup sync`, `hup refresh`, `hup version`) in other requirements of the main
  specs (`sync`, `repo-cache` refresh, `version-embedding`) are swept directly
  as doc-only edits, never inside a requirement a delta replaces.

## Risks / Trade-offs

- [Daily cron breaks after upgrade] → CHANGELOG migration steps; release notes
  marked BREAKING; v2.0.0 signals it.
- [Name search collides with AMD ROCm HIP] → accepted; the command does not
  collide (`hipcc`/`hipconfig` only).
- [Third rename / churn] → accepted; name declared final.
