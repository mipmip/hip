## MODIFIED Requirements

### Requirement: Collection store file

The system SHALL persist user-defined repository collections in a writable YAML
file in the configuration directory, `collections.yaml` alongside `config.yaml`
(`$XDG_CONFIG_HOME/hip/collections.yaml`, otherwise
`~/.config/hip/collections.yaml`). The file SHALL carry a schema version and
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
