# Tasks

## 1. Remote branch listing

- [x] 1.1 Add `RemoteBranches(ctx, dir)` to the `Git` interface and implement it
  on `ExecGit` with `for-each-ref`, excluding `origin/HEAD`; verify with an
  `ExecGit` unit test covering an empty remote and a remote with a branch.
- [x] 1.2 Extend `fakeGit` to serve remote branch lists so it still satisfies
  `Git`; verify `go test ./internal/syncer` passes.

## 2. Sync empty-upstream handling

- [x] 2.1 After `Fetch`, return `ActionSkipped` with
  `empty repository upstream: <owner>/<repo>` when `RemoteBranches` is empty;
  verify a unit test gets one skipped result and no failure.
- [x] 2.2 Resolve a non-empty remote's branch from the cached default, then
  `CurrentDefaultBranch`, then fail; verify unit tests for a stale cached name
  and an unresolvable branch.
- [x] 2.3 Broaden the `ActionSkipped` comment and document the empty-upstream
  skip in `BRIEFING.md` section 10; verify the section describes the skip and
  matches the implemented code path.

## 3. Integration and CLI behavior

- [x] 3.1 Extend the e2e test to clone an empty bare remote, sync once and
  expect skipped, add the first upstream commit, sync again and expect updated;
  verify `go test ./e2e`.
- [x] 3.2 Add a `runSync` test for an empty upstream that expects no error and a
  skipped count; verify `go test ./cmd/hip`.
- [x] 3.3 Run `bash scripts/coverage.sh` to confirm the coverage gate and the
  full suite, including e2e, still pass.

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Commit the change together with the bean closure.
