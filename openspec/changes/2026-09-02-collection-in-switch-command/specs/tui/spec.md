## MODIFIED Requirements

### Requirement: switch_command is rendered and executed shell-safely

The TUI SHALL render the active mode's `switch_command` as a Go text/template
against the selected repository — including a `Target` field holding the resolved
local clone path, and a `Collection` field holding the name of the collection
whose repository list is active, alongside the clone-path fields — then execute
it without a shell by splitting the rendered string into arguments using
shell-quoting rules, so repository, owner or collection names cannot inject shell
behavior.

The `Collection` field SHALL be the active collection's name whether that scope
was entered by the launch option or by navigating into a collection in the TUI,
and SHALL be the empty string wherever no collection scopes the list (the
combined view or an owner drill-down). No substitute value SHALL be invented for
that case; a configuration that wants a different name per scope expresses it
with a template conditional.

#### Scenario: Template has access to the local path

- **WHEN** a `switch_command` references `{{.Target}}`
- **THEN** it receives the repository's resolved local clone path

#### Scenario: Template has access to the active collection

- **WHEN** a `switch_command` references `{{.Collection}}` while the user is in a
  collection's repository list
- **THEN** it receives that collection's name, so a command can name a session
  after the collection and a window after the repository

#### Scenario: Collection is empty outside a collection

- **WHEN** a `switch_command` references `{{.Collection}}` while the user is in
  the combined view or an owner's repository list
- **THEN** it renders as the empty string, and a template conditional (for
  example falling back to the provider short and owner) is how a configuration
  chooses another name

#### Scenario: Names with metacharacters do not inject

- **WHEN** a repository, owner or collection name contains spaces or shell
  metacharacters and appears in the rendered command
- **THEN** it is passed as a single literal argument, not interpreted by a shell

#### Scenario: Command result is surfaced

- **WHEN** the executed command exits non-zero
- **THEN** the TUI reports the failure rather than silently continuing
