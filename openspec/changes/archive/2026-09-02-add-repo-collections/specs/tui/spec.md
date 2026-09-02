## ADDED Requirements

### Requirement: Select repositories while filtering

The TUI SHALL let the user toggle a repository's multi-selection while the fuzzy
filter input is focused, without altering the query. A dedicated
selection-and-advance key (`Tab`) SHALL toggle the highlighted repository and
move the cursor down; its reverse (`shift+Tab`) SHALL toggle and move up. These
keys SHALL work both while filtering and in list navigation, where `space` keeps
its existing toggle-in-place behavior. Because the multi-selection is held on the
underlying repository set rather than on the filtered view, selections SHALL
survive query changes, facet changes, sorting and navigation, so the user can
narrow, pick, re-narrow and pick again. The keys SHALL be inert at the provider
and owner levels and in multiplex mode.

#### Scenario: Toggle selection with the query focused

- **WHEN** the filter input is focused with a query typed and the user presses
  the selection-and-advance key
- **THEN** the highlighted repository is selected, the cursor moves to the next
  row, and the query text is unchanged (no character is inserted and the visible
  window is not reset)

#### Scenario: Space stays a filter character

- **WHEN** the filter input is focused and the user presses `space`
- **THEN** a space is inserted into the query as a term separator and nothing is
  selected

#### Scenario: Selections survive re-filtering

- **WHEN** repositories are selected under one query and the user then replaces
  the query with a different one and selects more
- **THEN** all repositories selected under either query remain selected

#### Scenario: Reverse toggle moves up

- **WHEN** the user presses the reverse selection key
- **THEN** the highlighted repository is toggled and the cursor moves to the
  previous row

### Requirement: Selection basket facet

The TUI SHALL provide a tri-state `selected` facet that narrows the repository
list to the current multi-selection (`only`), excludes it (`hide`), or applies no
constraint (`all`), so the user can review the repositories they have picked
after the queries that found them are gone. It SHALL compose with the other
facets and the fuzzy query like any facet, appear in the active-facet display,
and leave the selection togglable from within the narrowed list.

#### Scenario: Review the selection

- **WHEN** repositories have been selected under several different queries and
  the user cycles the `selected` facet to `only`
- **THEN** the list shows the selected repositories that are in the current
  scope, regardless of the query that selected them

#### Scenario: Un-pick from the basket

- **WHEN** the `selected` facet is `only` and the user toggles a visible row off
- **THEN** that repository is deselected and leaves the narrowed list

#### Scenario: Basket facet is displayed

- **WHEN** the `selected` facet is not `all`
- **THEN** the active-facet display includes it (e.g. `selected:only`)

### Requirement: Add the selection to a collection

The TUI SHALL provide an action that adds the multi-selected repositories to a
user-defined collection. The action SHALL open a modal listing the existing
collections with their member counts plus an entry that creates a new collection
by prompting for a name. Committing SHALL persist the membership, report how many
repositories were added and how many were already members, and clear the
multi-selection. Cancelling SHALL leave both the store and the selection
unchanged.

#### Scenario: Add to an existing collection

- **WHEN** repositories are selected and the user chooses an existing collection
  in the modal
- **THEN** the repositories are added to that collection, the store is persisted,
  the status reports the added and already-member counts, and the selection is
  cleared

#### Scenario: Create a collection while adding

- **WHEN** the user chooses the new-collection entry, types a name and commits
- **THEN** the collection is created with the selected repositories as its
  members

#### Scenario: New name collides with an existing collection

- **WHEN** the typed name is already used by a collection
- **THEN** the modal reports the collision and the store is unchanged

#### Scenario: Cancelling changes nothing

- **WHEN** the modal is dismissed without committing
- **THEN** no collection is created or modified and the repositories stay
  selected

#### Scenario: No selection

- **WHEN** the action is triggered with no repository selected
- **THEN** the highlighted repository is offered instead, or the action reports
  that there is nothing to add, rather than committing an empty change

### Requirement: Collections navigation level

The TUI SHALL expose the user-defined collections as a navigation level reached
from a synthetic `Collections` entry at the provider level. That level SHALL list
the collections with their member counts, and entering a collection SHALL show
its members as a flat cross-provider repository list scoped to that collection,
with the same filtering, facets, sorting, cloning and primary action as the
combined view. Navigating back SHALL return from a collection to the collections
list and from there to the provider level. Members that cannot be resolved
against the cache SHALL be rendered dim and identified as unresolved, not
omitted.

#### Scenario: Browse a collection

- **WHEN** the user selects the `Collections` entry and enters a collection
- **THEN** the repositories in that collection are listed across providers and
  owners, and the filter, facets and sort apply to them

#### Scenario: Back navigation unwinds the collection levels

- **WHEN** the user navigates back from a collection's repository list
- **THEN** the collections list is shown, and a further back navigation returns
  to the provider level

#### Scenario: Unresolved members are visible

- **WHEN** a collection contains a member that is not in the cache
- **THEN** the row is shown dim and marked as unresolved rather than being hidden

#### Scenario: Collection scoping composes with the primary action

- **WHEN** a collection's repository list is shown in multiplex mode and the user
  presses Enter on a member
- **THEN** the repository is cloned if needed and the `switch_command` runs, as in
  the combined view

### Requirement: Manage collections from the TUI

The TUI SHALL support the full lifecycle of a collection without leaving the
application: removing members from within a collection's repository list,
renaming a collection, and deleting a collection. Removal SHALL act on the
multi-selection when one exists and otherwise on the highlighted member.
Deleting a collection SHALL require confirmation. Every mutation SHALL persist
immediately and report its outcome in the status line.

#### Scenario: Remove members from a collection

- **WHEN** the user removes the selected (or highlighted) members inside a
  collection's repository list
- **THEN** they are no longer members, the list refreshes, the change is
  persisted, and the repositories themselves are untouched on disk

#### Scenario: Rename a collection

- **WHEN** the user renames a collection from the collections list and commits an
  unused name
- **THEN** the collection is renamed in place with its members intact; a name
  already in use is refused with an error

#### Scenario: Delete a collection with confirmation

- **WHEN** the user triggers delete on a collection
- **THEN** a confirmation is required, and only on confirmation is the collection
  removed and the store persisted; cancelling leaves it untouched

#### Scenario: Deleting a collection does not touch clones

- **WHEN** a collection is deleted
- **THEN** no repository is removed from disk — a collection is a view, not
  storage

### Requirement: Launch into a collection

The TUI SHALL support launching directly into a collection's repository list via
a `--collection <name>` launch option, composing with the mode selection and the
filter-active launch options. An unknown collection name SHALL fail with a clear
error and a non-zero exit rather than opening an empty list.

#### Scenario: Launch scoped to a collection

- **WHEN** the TUI is launched with `--collection terraform-modules`
- **THEN** it opens in that collection's flat cross-provider repository list

#### Scenario: Composes with multiplex and filter-active launch

- **WHEN** the TUI is launched with `--collection`, `--mode multiplex` and
  `--filter`
- **THEN** it opens the collection's list filtering in multiplex mode, so typing
  and pressing Enter clones-if-needed and runs the `switch_command` within that
  collection

#### Scenario: Unknown collection name

- **WHEN** the TUI is launched with a collection name that does not exist
- **THEN** it exits non-zero with an error naming the unknown collection

## MODIFIED Requirements

### Requirement: Interactive facet filters

The TUI SHALL provide tri-state facet filters for fork status, archived status
and multi-selection membership, and a value-cycle filter for visibility, applied
to the currently cached repositories and composed with the fuzzy filter.

#### Scenario: Fork facet narrows the list

- **WHEN** the user cycles the fork facet to `only` or `hide`
- **THEN** the list shows only forks, or excludes forks, respectively

#### Scenario: Archived facet narrows the list

- **WHEN** the user cycles the archived facet to `only` or `hide`
- **THEN** the list shows only archived, or excludes archived, respectively

#### Scenario: Selected facet narrows the list

- **WHEN** the user cycles the selected facet to `only` or `hide`
- **THEN** the list shows only multi-selected repositories, or excludes them,
  respectively

#### Scenario: Visibility facet cycles values

- **WHEN** the user cycles the visibility facet through `all` → `public` →
  `private` → `internal`
- **THEN** the list is limited to repositories of the selected visibility (all
  when `all`)

#### Scenario: Facets compose with fuzzy search

- **WHEN** one or more facets are active and a fuzzy query is entered
- **THEN** the visible repositories satisfy every active facet AND the fuzzy
  query

#### Scenario: Active facets are visible

- **WHEN** any facet is not `all`
- **THEN** the UI indicates the active facet states (e.g. `fork:hide vis:private`)

### Requirement: Position indicator

The TUI SHALL display the position of the visible window within the current
list, and SHALL report the size of the current multi-selection whenever at least
one repository is selected. The reported selection count SHALL be the total
across all providers, not only the repositories visible in the current scope, so
the full basket size stays known while browsing a narrowed list.

#### Scenario: Indicator shows visible range and total

- **WHEN** a list is displayed and windowed
- **THEN** an indicator shows the visible row range and the total count
  (e.g. `41-60 of 213`)

#### Scenario: Indicator reports the selection count

- **WHEN** one or more repositories are multi-selected
- **THEN** the indicator also reports how many are selected
  (e.g. `41-60 of 213 · 4 selected`)

#### Scenario: Selection count is global

- **WHEN** repositories are selected across several owners and the user is
  viewing a single owner's list
- **THEN** the reported selection count is the total selection across all
  providers, even though only some selected rows are in scope

### Requirement: Multiplex mode hides multiselect

In multiplex mode the repository list SHALL NOT show the multi-select checkbox,
and the selection-toggle keys — including the selection-and-advance keys usable
while filtering — SHALL have no effect, because the primary action operates on a
single repository.

#### Scenario: No checkbox in multiplex

- **WHEN** the repository list is shown in multiplex mode
- **THEN** rows are rendered without the `[ ]`/`[x]` multi-select column

#### Scenario: Selection toggle is inert in multiplex

- **WHEN** the user presses any selection-toggle key in multiplex mode, whether
  in list navigation or while filtering
- **THEN** nothing is selected and the list is unchanged

#### Scenario: Multiselect remains in management

- **WHEN** the repository list (including the combined view) is shown in
  management mode
- **THEN** the multi-select checkbox and toggle continue to work
