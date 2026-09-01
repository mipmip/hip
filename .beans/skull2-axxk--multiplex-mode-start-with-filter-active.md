---
# skull2-axxk
title: multiplex mode start with filter active
status: completed
type: feature
priority: normal
created_at: 2026-09-01T18:36:40Z
updated_at: 2026-09-01T20:12:11Z
---

## Summary of Changes

Added a `hup tui --filter` launch flag that boots the TUI straight into the
combined cross-provider flat repo list with the fuzzy filter active and focused
— an fzf-style picker: launch, type to narrow, Enter to act.

- `--filter` implies the flat view (the fuzzy filter only searches repositories
  at the repos level) and composes with `--mode multiplex`, so typing a repo and
  pressing Enter clones-if-needed and runs the `switch_command`.
- Esc keeps the existing two-stage unwind (first Esc clears the query + stops
  filtering, second Esc backs the flat view out to the provider list); this is
  now pinned by a spec scenario rather than changed.
- Threaded a `filter bool` through `tui.Run`; extracted `startFiltering` (shared
  by the `/` key and the flag) and an `applyLaunchOptions` seam for testing.

OpenSpec change `multiplex-start-filtering` archived
(`2026-09-01-multiplex-start-filtering`); shipped in commit `7f68086`.
