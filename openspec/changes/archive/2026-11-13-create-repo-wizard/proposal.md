# Proposal

## Why

HIP can browse and clone a large portfolio, but creating a new repository still
means leaving the TUI for each provider's web UI and then refreshing the cache.
A TUI wizard that creates the remote, scaffolds README/license content, updates
the cache and optionally checks out the repository keeps the whole workflow in
one place.

## What Changes

- Add an `n` action in the provider, owner and repository lists that opens a
  create-repository wizard.
- The wizard collects provider, owner, name, description, visibility,
  README, license and checkout choices, with known provider/owner context
  prefilled and a review step before creating.
- Create the remote with no provider-side initialization, build the initial
  commit in a temporary git worktree, push `main`, then add the repository to
  the provider cache without dropping siblings.
- Open the README template in `$EDITOR` (fallback `vi`); offer `none`, MIT,
  Apache-2.0 and GPL-3.0 license texts.
- Optionally clone the new repository through the configured clone-path
  template and clone VCS.
- Add provider create-repository calls for GitHub, GitLab and Codeberg,
  including personal versus organization targets and actionable auth errors.

## Capabilities

### New Capabilities

- `repo-creation`: end-to-end repository creation workflow, from remote
  creation through local scaffolding, cache insertion and optional checkout.

### Modified Capabilities

- `provider-abstraction`: the provider interface gains repository creation.
- `provider-clients`: each provider client creates repositories with the
  correct owner target and visibility.
- `repo-cache`: the cache can add one repository without replacing an owner's
  existing repositories.
- `tui`: the TUI gains the create-repository wizard and its launch key.

## Impact

- `internal/provider`: new create method, options type, POST helper and
  per-provider implementations.
- `internal/cache`: additive repository insertion.
- New `internal/creator`: temp worktree, README/license templates, commit,
  push, cache update and optional checkout orchestration.
- `internal/tui`: wizard state, modal rendering, key handling, editor
  integration and tests.
- `README.md`, `CHANGELOG.md`; no configuration, CLI command or dependency
  changes.
