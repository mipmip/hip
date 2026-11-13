# Spec Delta

## ADDED Requirements

### Requirement: Add a single repository without disturbing an owner

The cache SHALL support adding or replacing one repository for an owner while
preserving that owner's other repositories and existing fetch state, and SHALL
persist the result atomically.

#### Scenario: Owner has other cached repositories
- **WHEN** a repository is added for an owner with existing cached repositories
- **THEN** the other repositories remain present

#### Scenario: Owner was never fetched
- **WHEN** a repository is added for an owner whose fetch state is unfetched
- **THEN** the owner's fetch state remains unfetched

#### Scenario: Repository already exists
- **WHEN** a repository with the same owner and name is added
- **THEN** its fields are replaced without creating a duplicate entry

#### Scenario: Persist after insertion
- **WHEN** a repository is added
- **THEN** the cache file is written atomically with the new contents
