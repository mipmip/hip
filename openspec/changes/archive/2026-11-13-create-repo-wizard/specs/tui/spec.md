# Spec Delta

## ADDED Requirements

### Requirement: Launch the create-repository wizard

The TUI SHALL open a create-repository wizard from the provider, owner and
repository list levels with `n`. It SHALL preselect the provider and owner
already in scope. The key SHALL be inactive in the detail view, while filtering,
and while another modal owns the screen.

#### Scenario: Launch from the provider level
- **WHEN** the user presses `n` on the provider list
- **THEN** the wizard opens and asks for the provider

#### Scenario: Launch from the owner level
- **WHEN** the user presses `n` on an owner's repository list
- **THEN** the wizard opens with that provider and owner preselected

#### Scenario: Launch from the repository level
- **WHEN** the user presses `n` on a repository row
- **THEN** the wizard opens with that repository's provider and owner preselected

#### Scenario: Detail view
- **WHEN** the user presses `n` in the detail view
- **THEN** nothing happens

#### Scenario: Filtering
- **WHEN** the filter input is active
- **THEN** `n` remains a query character

### Requirement: Choose the repository owner and visibility

The wizard SHALL offer only the self account, configured owners and discovered
owners as the repository owner. It SHALL offer public and private visibility,
defaulting to private.

#### Scenario: Owner picker
- **WHEN** the wizard asks for the owner
- **THEN** the choices are limited to the known owners for the selected provider

#### Scenario: Visibility default
- **WHEN** the wizard reaches the visibility step
- **THEN** private is selected

#### Scenario: Public visibility
- **WHEN** the user selects public
- **THEN** the creation request uses public visibility

### Requirement: Prepare README and license choices

The wizard SHALL offer an embedded README template opened in the user's editor,
and a license choice of none, MIT, Apache-2.0 or GPL-3.0. The license holder
SHALL come from the git user name and the current year, prompting for the name
only when it is unavailable.

#### Scenario: Edit the README
- **WHEN** the user accepts the README step
- **THEN** the embedded template is opened in `$EDITOR`, falling back to `vi`

#### Scenario: Choose a license
- **WHEN** the user selects MIT, Apache-2.0 or GPL-3.0
- **THEN** the corresponding license text is written with the resolved holder and year

#### Scenario: Decline both
- **WHEN** the user declines README and license
- **THEN** the wizard proceeds without preparing local files

#### Scenario: Missing git user name
- **WHEN** a license is selected but no git user name is configured
- **THEN** the wizard prompts for the holder name before continuing

### Requirement: Review and execute creation

The wizard SHALL show a review of the chosen provider, owner, name, visibility,
README, license and checkout before creating. It SHALL show progress while
creating, add the new repository to the current list on success, and show the
error without discarding the wizard on failure.

#### Scenario: Confirm creation
- **WHEN** the user confirms the review
- **THEN** the wizard creates the remote, pushes any initial commit and updates the cache

#### Scenario: Creation succeeds
- **WHEN** creation succeeds
- **THEN** the wizard closes and the new repository is selected in the current list

#### Scenario: Creation fails
- **WHEN** remote creation or the initial push fails
- **THEN** the wizard remains open with the error and no cache entry is added

#### Scenario: Optional checkout
- **WHEN** creation succeeds with checkout requested
- **THEN** the repository is cloned after it has been added to the cache
