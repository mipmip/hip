# Tasks

## 1. Provider creation API

- [x] 1.1 Add `CreateRepoOptions`, visibility normalization, a shared authenticated `postJSON` helper, and provider-specific 403 permission hints; verify `go test ./internal/provider` compiles and helper tests cover payload, status, and scope hints.
- [x] 1.2 Implement GitHub `CreateRepo` for `POST /user/repos` and `POST /orgs/{owner}/repos`, normalize the response into `provider.Repo`, and verify httptest cases for self, organization, public/private, and creation errors.
- [x] 1.3 Implement GitLab `CreateRepo` for `POST /projects`, resolve group owners through `GET /namespaces?search=` with an exact `full_path` match, omit `namespace_id` for the self namespace, and verify tests for personal, group, missing namespace, and visibility.
- [x] 1.4 Implement Codeberg `CreateRepo` for `POST /user/repos` and `POST /orgs/{owner}/repos`, normalize the response, and verify httptest cases for self, organization, private, and creation errors.
- [x] 1.5 Add `CreateRepo` to the `provider.Provider` interface and verify the default registry builds all three clients and a compile-time fake satisfies the interface.

## 2. Creator pipeline and cache

- [x] 2.1 Export the clone URL rule from `internal/syncer` as `CloneURL`, reuse it in all existing clone paths, and verify `go test ./internal/syncer` still covers HTTPS and SSH selection.
- [x] 2.2 Add `cache.AddRepo` with case-insensitive owner/name replacement, preservation of sibling repositories and fetched state, and atomic saving; verify tests cover insert, replace, unfetched owner, and sibling preservation.
- [x] 2.3 Add `internal/creator`'s temporary worktree (`git init -b main`, write files, add all, commit, push, cleanup) with an injectable git runner; verify real-git tests cover files, commit, push to a bare remote, cleanup, and command errors.
- [x] 2.4 Add the creator engine that creates the remote, commits and pushes only when content exists, adds the repository to the cache, and returns a typed error naming the retained worktree when a push fails; verify fake-provider and fake-store tests cover success, no-content, remote-create failure, push failure, and cache failure.

## 3. TUI wizard

- [x] 3.1 Add wizard state and modal rendering, bind `n` from the provider, owner, and repository lists, preselect the current provider/owner, and verify tests that `n` is ignored in detail, filter, and other modals.
- [x] 3.2 Implement provider, owner, name, description, and visibility steps, restrict owners to self/configured/discovered owners, default to private, and verify transition and validation tests.
- [x] 3.3 Implement README and license steps with embedded templates, `$EDITOR` fallback to `vi` via `tea.ExecProcess`, license choice (`none`, MIT, Apache-2.0, GPL-3.0), holder from git config or a prompt, and verify editor cancel, missing holder, and license rendering tests.
- [x] 3.4 Implement review and execution: prepare the worktree before remote creation, call the creator engine, show progress and errors, insert and select the cache row, and optionally clone using the rendered target and configured VCS; verify tests for success, push failure, cache failure, and optional-clone failure.
- [x] 3.5 Add the `n: new repo` hint to the list footer and document the wizard in the README key table; verify footer tests and the documented keys match the bindings.

## 4. Integration verification

- [x] 4.1 Add an integration test that creates a repository through a fixture provider, pushes the initial commit to a local bare remote, updates the cache, and verifies the optional clone target and VCS selection; run `go test ./e2e`.
- [x] 4.2 Run `bash scripts/coverage.sh` and verify overall coverage is at least 70% with every core package at least 80%.

## Workflow follow-up

- Add an `### Added` bullet for the repository creation wizard under `## [Unreleased]` in `CHANGELOG.md`.
- Archive the change with `scripts/ship-change.sh` after the gate passes.
- Close bean `skull2-uwha` with a `## Summary of Changes` section and commit the closure as Pim Snel without self-promotion.
