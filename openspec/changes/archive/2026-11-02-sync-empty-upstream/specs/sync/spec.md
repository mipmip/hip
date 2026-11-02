# Spec Delta

## MODIFIED Requirements

### Requirement: Fast-forward pull existing repositories

The system SHALL update existing clones with a fetch and fast-forward-only pull
on the default branch, never forcing. It SHALL resolve the default branch
against the fetched remote refs, and SHALL treat a remote with no branches as
skipped rather than failed.

#### Scenario: Clean repo is fast-forwarded

- **WHEN** a local clone exists and is clean and behind its remote
- **THEN** the system fetches and fast-forward-pulls the default branch

#### Scenario: Dirty repo is skipped

- **WHEN** a local clone has uncommitted changes
- **THEN** the system skips it and records a warning instead of modifying it

#### Scenario: Empty upstream is skipped

- **WHEN** a local clone exists and is clean but the fetched remote has no
  branches
- **THEN** the system records it as skipped with a warning naming the
  repository, not as failed

#### Scenario: Stale default branch falls back to remote HEAD

- **WHEN** the cached default branch does not exist upstream but the fetched
  remote has other branches and a remote HEAD
- **THEN** the system fast-forward-pulls the branch named by the remote HEAD

#### Scenario: Unresolvable default branch fails

- **WHEN** the fetched remote has branches but neither the cached default branch
  nor a remote HEAD names one
- **THEN** the system records a failure for that repository
