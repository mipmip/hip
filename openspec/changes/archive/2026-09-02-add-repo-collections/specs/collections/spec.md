## ADDED Requirements

### Requirement: Collection store file

The system SHALL persist user-defined repository collections in a writable YAML
file in the configuration directory, `collections.yaml` alongside `config.yaml`
(`$XDG_CONFIG_HOME/huphop/collections.yaml`, otherwise
`~/.config/huphop/collections.yaml`). The file SHALL carry a schema version and
SHALL represent collections and their members as ordered lists, so member order
is the order in which repositories were added and the file stays stable across
writes. A missing file SHALL load as an empty store without error; the directory
SHALL be created on the first write. Writes SHALL be atomic (write a temporary
file in the same directory, then rename).

#### Scenario: Missing file loads as an empty store

- **WHEN** no `collections.yaml` exists
- **THEN** the store loads successfully with zero collections, and no file is
  created until a collection is written

#### Scenario: Member order is insertion order

- **WHEN** repositories are added to a collection in a given order and the store
  is reloaded
- **THEN** the members are returned in the order they were added

#### Scenario: Write is atomic

- **WHEN** a store is saved
- **THEN** the file is replaced by rename from a temporary file in the same
  directory, so an interrupted write cannot leave a partial `collections.yaml`

#### Scenario: Invalid file is reported, not silently ignored

- **WHEN** `collections.yaml` cannot be parsed, or contains two collections with
  the same name
- **THEN** loading returns an error naming the file, rather than silently
  discarding collections

### Requirement: Cross-provider member identity

A collection member SHALL be identified by the provider name (the unique
`providers[].name` key from the configuration), the owner, and the repository
name, so that one collection can hold repositories from different owners,
organizations and providers. Owner and repository names SHALL be stored as
supplied and compared case-insensitively.

#### Scenario: One collection spans owners and providers

- **WHEN** repositories from two different owners on two different providers are
  added to the same collection
- **THEN** all of them are members of that collection and are resolved against
  their respective providers

#### Scenario: Membership comparison is case-insensitive

- **WHEN** a member is stored as `Org1/Terraform-AWS` and a repository resolves
  as `org1/terraform-aws` on the same provider
- **THEN** they are treated as the same member

### Requirement: Collection operations

The store SHALL support creating a collection, adding repositories to it,
removing members from it, renaming it, and deleting it. Collection names SHALL
be unique: creating or renaming onto an existing name SHALL fail with an error
rather than merging the two collections. Adding a repository that is already a
member SHALL be a no-op, and the operation SHALL report how many of the
requested repositories were added and how many were already members.

#### Scenario: Create and add

- **WHEN** a collection is created and repositories are added to it
- **THEN** the collection exists with those repositories as members, and the
  store is persisted

#### Scenario: Adding an existing member does not duplicate it

- **WHEN** repositories are added to a collection and some are already members
- **THEN** the collection contains each repository exactly once, and the result
  reports the added count and the already-member count

#### Scenario: Remove members

- **WHEN** members are removed from a collection
- **THEN** they are no longer members, the remaining members keep their relative
  order, and the collection continues to exist even when it becomes empty

#### Scenario: Rename refuses an existing name

- **WHEN** a collection is renamed to a name another collection already uses
- **THEN** the rename fails with an error and neither collection is modified

#### Scenario: Rename keeps position and members

- **WHEN** a collection is renamed to an unused name
- **THEN** its members and its position in the file are unchanged

#### Scenario: Delete removes the collection

- **WHEN** a collection is deleted
- **THEN** it is absent from the store and from the persisted file, and other
  collections are untouched

### Requirement: Unresolvable members are preserved

The store SHALL preserve members that cannot currently be resolved to a cached
repository — for example a repository that was renamed or deleted upstream, whose
owner has not been fetched, or whose provider entry was renamed in the
configuration. Such members SHALL remain in the file and be reported as
unresolved rather than being dropped.

#### Scenario: Member whose provider is no longer configured

- **WHEN** a collection has a member on a provider name that is not in the
  current configuration
- **THEN** the member remains in `collections.yaml` and is reported as
  unresolved, and saving the store does not remove it

#### Scenario: Member whose repository is not in the cache

- **WHEN** a member's repository is absent from the cache (never fetched, or
  removed upstream)
- **THEN** the member remains a member and is reported as unresolved
