# Proposal

## Why

`hip sync` runs unattended as the daily backup in mipnix, where a non-zero exit
is the only failure signal. When a forge repository has no commits, cloning
succeeds but the next fast-forward fails because `origin/<default-branch>` does
not exist. Six empty repositories currently turn every run into a reported
failure and mask real ones (bean skull2-f4m3).

## What Changes

- After fetching an existing clone, detect that the remote has no branches
  (ignoring `origin/HEAD`) and report the repository as skipped with a warning
  that names it, instead of failed.
- Resolve the default branch against the fetched remote refs: use the cached
  default branch when it exists upstream, otherwise the remote HEAD, otherwise
  fail.
- Preserve clone behavior: a missing empty repository is still cloned, and once
  upstream receives its first commit the next sync fast-forwards the unborn
  local branch.
- Update the skip action's documentation and cover the empty-upstream,
  first-commit, and CLI exit-status cases in tests.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sync`: "Fast-forward pull existing repositories" now treats an empty upstream
  as a skipped repository and resolves the default branch from the fetched
  remote refs.

## Impact

- `internal/syncer/syncer.go`: existing-clone flow after fetch, skip warning.
- `internal/syncer/git.go`: the `Git` interface and `ExecGit` gain remote-branch
  listing; the test fake is updated to match.
- Tests: `internal/syncer/syncer_test.go`, `e2e/e2e_test.go`,
  `cmd/hip/main_test.go`.
- No configuration, provider API, cache format, or CLI flag changes.
