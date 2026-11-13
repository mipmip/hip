# repo-creation Specification

## Purpose
Creates a new repository across a supported provider, gives it an initial
commit, records it in the local cache and optionally checks it out.

## Requirements

### Requirement: Prepare initial content before creating the remote

The system SHALL build the initial commit in a temporary git worktree before
creating the remote, so cancelling local preparation never leaves an orphan
remote. The worktree SHALL use `main` as the initial branch.

#### Scenario: README and license selected
- **WHEN** the user requests a README and a license
- **THEN** the temporary worktree contains both files, staged in one commit on `main`

#### Scenario: No initial content selected
- **WHEN** the user declines both README and license
- **THEN** no initial commit is prepared and remote creation proceeds without one

#### Scenario: Editor is cancelled
- **WHEN** the user cancels the README editor
- **THEN** the workflow stops without creating a remote

### Requirement: Create an empty remote and push the initial commit

The system SHALL create the remote with provider-side initialization disabled,
then push the prepared initial commit to `main` when one exists.

#### Scenario: Remote creation succeeds
- **WHEN** the provider accepts the create request
- **THEN** the remote exists with no provider-created files

#### Scenario: Initial push succeeds
- **WHEN** an initial commit was prepared and the remote was created
- **THEN** the commit is pushed to the remote's `main` branch

#### Scenario: Initial push fails
- **WHEN** the remote was created but the initial push fails
- **THEN** the system reports that the remote may exist and does not add the repository to the cache

### Requirement: Add the created repository to the cache

The system SHALL add a successfully created repository to its provider cache
without replacing other repositories for the same owner and without changing
that owner's fetch state. The cache SHALL be written atomically.

#### Scenario: Owner already has cached repositories
- **WHEN** a repository is added for an owner with existing cached repositories
- **THEN** the existing repositories remain and the new repository is added

#### Scenario: Owner was never fetched
- **WHEN** a repository is added for an owner whose fetch state is unfetched
- **THEN** the owner remains unfetched after the insertion

#### Scenario: Repository already exists in the cache
- **WHEN** a repository with the same owner and name is added
- **THEN** its cached fields are replaced rather than duplicated

### Requirement: Optionally check out the created repository

After creation succeeds, the system SHALL clone the repository only when the
user requested a checkout, using the configured clone-path template and clone
VCS. A checkout failure SHALL leave the repository in the cache and report the
clone error.

#### Scenario: Checkout requested
- **WHEN** creation succeeds and the user requested a checkout
- **THEN** the repository is cloned to its rendered target path with the configured VCS

#### Scenario: Checkout declined
- **WHEN** creation succeeds and the user declined a checkout
- **THEN** no clone is attempted

#### Scenario: Checkout fails
- **WHEN** creation succeeds but cloning fails
- **THEN** the cached repository remains and the clone error is shown
