# Design

## Context

See `proposal.md` for motivation. `syncRepo` currently clones a missing target,
skips dirty trees, fetches, resolves a branch from the cached default branch or
`origin/HEAD`, then runs `git merge --ff-only origin/<branch>`. On a remote with
no commits, the fetch succeeds but leaves no remote refs, so the merge fails.

`Git` has two implementations: `ExecGit` and the test fake. jj checkouts are
colocated, so the same `.git` remote refs are available to both.

## Goals / Non-Goals

**Goals:**

- Detect an empty upstream from local remote refs after fetch.
- Resolve the default branch against those refs, recovering from a stale cached
  name.
- Report an empty upstream as skipped with an identifiable warning.

**Non-Goals:**

- Change how missing repositories are cloned.
- Use provider API metadata such as repository size.
- Change TUI clone-on-demand behavior.
- Push local commits anywhere.

## Decisions

**List remote branches with one `Git` method.** Add
`RemoteBranches(ctx, dir) ([]string, error)`, implemented with
`git for-each-ref --format='%(refname:short)' refs/remotes/origin` and filtering
out `origin/HEAD`. `for-each-ref` exits 0 when nothing matches, so there is no
exit code to interpret. One call answers both "is the remote empty?" and "does
the cached branch still exist?".

Alternatives: a count-only method is smaller but leaves a renamed default branch
failing forever. A per-branch `rev-parse` check needs two methods and still
fails before branch resolution when the cached default branch is empty.

**Check immediately after fetch, before resolving the branch.** Zero remote
branches means skip with `empty repository upstream: <owner>/<repo>`. This comes
before `CurrentDefaultBranch`, which cannot resolve on an empty remote and also
covers a null GitLab `default_branch`.

**Resolve in three steps.** Use the cached default branch when it appears in the
remote branch list; otherwise use `CurrentDefaultBranch`; otherwise fail with
`resolving default branch`. This keeps today's behavior when the cache is
correct, recovers renames, and still reports a genuinely broken state.

**Name the repository in the warning.** The CLI stderr line for skips does not
include the repository name, so the warning text must carry it or six empty
repos produce six identical journal lines.

## Risks / Trade-offs

- **Extra git subprocess per existing repo** -> Accept the cost for correctness;
  it is one `for-each-ref` call alongside the existing fetch and merge, and can
  be narrowed later if a large portfolio shows measurable overhead.
- **`origin/HEAD` missing or stale** -> Exclude it from the branch list and fall
  back to the cached branch, then fail if neither names a branch. No silent
  wrong-branch merge.
- **Remote gains its first commit between fetch and merge** -> The sync uses the
  refs it just fetched, so it skips this run and fast-forwards on the next one.
- **Local commits in an unborn clone of an empty remote** -> The repo is still
  skipped, since sync never pushes. This matches existing never-force behavior.
