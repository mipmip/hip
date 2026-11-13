# Design

## Context

HIP currently reads from three providers and never writes to them. Creating a
repository means leaving the TUI, creating it on the forge, then refreshing the
cache and cloning it by hand. The cache is organized per owner: replacing an
owner's entries is supported, but adding one repository without disturbing the
others is not.

The wizard spans four boundaries: the TUI, the provider API, a temporary local
git repository, and the cache. The initial commit must be prepared before the
remote exists, otherwise cancelling the editor would leave an orphan remote.
The final checkout is optional and must follow the configured clone path and
VCS.

## Goals / Non-Goals

**Goals:**

- Create an empty repository through the selected provider from the TUI.
- Prepare and push an initial commit with an optional README and license.
- Add the created repository to the cache without a full refresh.
- Optionally clone the new repository using the existing clone path and VCS
  rules.
- Keep failures explicit: no cache entry for a remote that was not created and
  pushed, and a clear message when a remote may exist after a failed push.

**Non-Goals:**

- A CLI create command.
- Renaming or deleting repositories.
- GitLab `internal` visibility.
- Provider-side README or license generation.
- Configurable remote templates beyond the embedded README and license set.
- Deleting a remote automatically after a failed local step.

## Decisions

### Provider contract

Add a method to `provider.Provider`:

```go
type CreateRepoOptions struct {
    Owner       string
    Name        string
    Description string
    Visibility  Visibility
}

CreateRepo(ctx context.Context, opts CreateRepoOptions) (Repo, error)
```

An options struct keeps the signature stable when more forge fields are added.
The method returns the normalized `Repo`, including clone URLs and the default
branch, so the caller does not need a follow-up API call.

The provider decides between a personal and an organization endpoint by
comparing `opts.Owner` with the configured `Username`, case-insensitively.

### Provider endpoints

| Provider | Personal | Organization |
| --- | --- | --- |
| GitHub | `POST /user/repos` | `POST /orgs/{owner}/repos` |
| GitLab | `POST /projects` without `namespace_id` | `POST /projects` with `namespace_id` |
| Codeberg | `POST /user/repos` | `POST /orgs/{owner}/repos` |

All three create an empty repository: `auto_init` is false, and GitLab's
`initialize_with_readme` is false. GitLab group ownership is resolved with
`GET /namespaces?search=<owner>` and an exact `full_path` match. A missing
namespace is an error before any create call.

Visibility maps to `private` on GitHub and Codeberg, and to `visibility` on
GitLab. Only `public` and `private` are exposed.

A shared `postJSON` helper mirrors the existing `get` helpers: authenticated
request, JSON body, status and body in errors. A 403 error adds a scope hint:
GitHub `repo` for private and `public_repo` for public, GitLab `api`, Codeberg
`write:repository`.

### Temporary worktree

A new `internal/creator` package owns the temporary git worktree. It exposes a
session with:

- `NewWorktree()`: creates `/tmp/hip-create-*` and runs `git init -b main`.
- `WriteREADME(content)` and `WriteLicense(content)`: write files in that
  directory.
- `CommitAll(message)`: stages and commits.
- `Push(remoteURL)`: adds `origin` and pushes `main`.
- `Cleanup()`: removes the directory.

The worktree is plain git, regardless of the VCS configured for the final
checkout. The wizard holds the session across editor and review steps, so the
editor can run through `tea.ExecProcess` without losing state.

The push URL uses the same rule as sync: HTTPS when the provider's
`clone_protocol` is `https`, otherwise SSH. The existing `cloneURL` helper in
`internal/syncer` is exported as `CloneURL` and reused, so the rule has one
implementation.

If neither README nor license is requested, the worktree is not committed and
nothing is pushed. The remote is created empty and the empty-upstream sync
behavior already covers it.

### Create sequence

```text
prepare temp worktree and files
    |
    v
provider.CreateRepo(auto_init=false)
    |
    v
push main (only when there is initial content)
    |
    v
cache.AddRepo
    |
    v
optional clone
```

Preparing first means editor cancellation and write errors cannot leave a
remote behind. If remote creation fails, the worktree is removed and the cache
is untouched. If the push fails, the remote exists; the error says so, the
worktree is retained for recovery, and the cache is untouched.

The cache is updated only after the remote exists and, when applicable, the
initial push succeeded. A failed optional clone does not remove the cache row;
the row stays and the clone error is shown.

### Cache insertion

Add `AddRepo(repo provider.Repo) error` to `cache.Cache`. It replaces a case-
insensitive match on owner and name, preserves all other repositories and the
owner's fetched state, and saves atomically with the existing temp-file and
rename path. It does not mark an unfetched owner as fetched.

### TUI wizard

Add a wizard modal owned by a `wizard` struct on `Model`. It is checked after
the clone, picker, and prompt modals and before the about modal. The steps are:

1. Provider, when the launch level does not already identify one.
2. Owner, restricted to the self account, configured owners, and discovered
   owners.
3. Name and description.
4. Visibility, private by default.
5. README yes/no, then the editor.
6. License, one of none, MIT, Apache-2.0, GPL-3.0.
7. Checkout yes/no.
8. Review and create.

The wizard is launched with `n` from the provider, owner, and repository lists.
Provider and owner are prefilled from the current level. It is ignored in the
detail view, during filtering, and while another modal owns the screen.

The README template is embedded and written before the editor opens. The editor
comes from `$EDITOR`, falling back to `vi`, and runs through `tea.ExecProcess`.
The license holder comes from `git config user.name` and the current year. If
the name is empty, the wizard asks for a holder before review.

Execution runs asynchronously so the TUI can show progress. After success the
cache row is inserted, the current list is refreshed from the cache, and the
new row is selected. If checkout was requested, the existing clone engine runs
with the rendered target and configured VCS.

## Risks / Trade-offs

- **Partial remote:** a push failure leaves an empty remote. The alternative,
  creating the remote only after a successful local commit, does not help
  because the push still needs the remote. Keeping the worktree and naming it
  in the error is the safest recovery.
- **Token scopes:** read-only tokens will fail at create time. The scope hints
  reduce guesswork, but permissions differ between classic and fine-grained
  GitHub tokens.
- **GitLab namespaces:** namespace lookup adds a request and can fail
  independently. It is done before `POST /projects` so a failed lookup cannot
  create a personal repository by accident.
- **Editor suspension:** `tea.ExecProcess` hands the terminal to the editor.
  The wizard must resume cleanly and treat a non-zero editor exit as a cancel.
- **Temp files:** a crash can leave `/tmp/hip-create-*` behind. Cleanup runs on
  cancel, success, and remote-create failure; a push failure deliberately keeps
  the directory and reports its path.
- **Cache consistency:** adding one row must not mark an owner fetched. The new
  method is tested against sibling repositories and fetched state.

## Migration Plan

None. Existing configuration, clone paths, and cached repositories are
unchanged. The feature only uses provider permissions that the user already has
or must grant.

## Open Questions

None.
