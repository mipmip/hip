---
# skull2-uwha
title: create new repo in one of the preconfigured git providers
status: completed
type: feature
priority: normal
created_at: 2026-11-13T13:43:46Z
updated_at: 2026-11-13T15:51:48Z
---

- When more then one providers select which.
- Auto update cache by adding it.
- ask to add readme. open template in EDITOR
- ask to add license
- Ask to checkout as well.

## Summary of Changes

- Added the `n` repository creation wizard to the TUI: provider and owner, name and description, visibility, optional README in `$EDITOR`, optional license, and optional checkout.
- Added repository creation for GitHub, GitLab and Codeberg, including namespace resolution and scope-aware permission errors.
- Added temporary worktree scaffolding, initial commit and push to `main`, additive cache insertion, and optional clone with the configured path and VCS.
- Added provider, cache, creator, TUI and end-to-end coverage; overall coverage is 79.9% with all core packages above 80%.
