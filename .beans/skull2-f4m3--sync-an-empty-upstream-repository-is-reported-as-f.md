---
# skull2-f4m3
title: 'sync: an empty upstream repository is reported as failed'
status: completed
type: bug
priority: high
created_at: 2026-11-02T11:13:20Z
updated_at: 2026-11-02T21:10:24Z
---

`hip sync` reports a repository as failed when the repository has no commits
on the forge. The clone succeeds (git only warns "You appear to have cloned an
empty repository"), but every later sync fails at the fast-forward step,
because `origin/<default_branch>` does not exist:

```
hip: github: mipmip/traefik-oc: fast-forwarding main: git merge --ff-only origin/main: exit status 1: merge: origin/main - not something we can merge
```

Inside such a checkout:

```
$ git log --oneline -1
fatal: your current branch 'main' does not have any commits yet
$ git status -sb
## No commits yet on main...origin/main [gone]
```

## Why it matters

hip runs unattended in mipnix as a daily backup on dapperehaan
(`hip-repo-backup.service`, mipnix change `add-github-repo-backup`). `hip sync`
exits non-zero when any repository fails, and that exit code is the only
failure signal the backup has. Six empty repositories now make the unit fail
on every run, so a real failure no longer stands out.

Run on 2026-11-02: `0 cloned, 725 updated, 0 skipped, 6 failed`. All six
report `size: 0` in the GitHub API:

- mipmip/traefik-oc
- mipmip/Fruits-And-Vegetables-Cards-svg-stencil
- mipmip/docker-jekyll-gulp-rsync
- TechNative-B-V/wtoorren-test
- TechNative-B-V/Static-Website-Assignment
- wearetechnative/elastixcan

## Where

`internal/syncer/syncer.go`, `syncRepo`: after `Fetch`, the branch comes from
`r.DefaultBranch` (or `CurrentDefaultBranch`), and `FastForward` in
`internal/syncer/git.go` runs `git merge --ff-only origin/<branch>`. Nothing
checks whether `origin/<branch>` exists first.

## Expected

- An existing checkout whose remote has no commits is not a failure. Report
  it as skipped with a warning such as `empty repository upstream`, and keep
  the exit code at 0 when nothing else failed.
- Once the upstream repository gets its first commit, the next sync
  fast-forwards the unborn local branch as usual.
- The clone path for a new empty repository keeps working as it does now.

## Approach to consider

Check for the remote branch before fast-forwarding, e.g.
`git rev-parse --verify --quiet refs/remotes/origin/<branch>`, and return
`ActionSkipped` when it is missing. Prefer this over relying on the API's
`size == 0`, which is approximate and only updated periodically by GitHub.
Covers git checkouts and jj colocated checkouts alike, since both have a
`.git` with remote refs.

## Tests

- Syncer unit test with a fake `Git` whose remote branch does not exist:
  result is `ActionSkipped`, not `ActionFailed`.
- e2e: create a bare empty repo as the remote, clone, sync again; expect
  skipped and exit 0. Then push a commit to the bare repo and sync; expect
  updated.



## Summary of Changes

Shipped as OpenSpec change `sync-empty-upstream` (archived as
`2026-11-02-sync-empty-upstream`).

`hip sync` now lists the fetched remote branches before fast-forwarding. A
repository whose upstream has no branches is reported as skipped with
`empty repository upstream: <owner>/<repo>` instead of failing, so an empty
forge repository no longer makes the backup unit exit non-zero. The default
branch is taken from the cache when it exists upstream, otherwise from the
remote HEAD, and an unresolvable branch is still a hard failure.

Covered by syncer unit tests (empty remote, stale default branch, cached
branch, unresolvable branch, branch-listing error), an e2e flow (empty clone,
skipped sync, first upstream commit, updated sync), and a `runSync` test that
asserts exit 0 and the skipped count. Coverage: 82.4% overall, syncer 83.1%.
