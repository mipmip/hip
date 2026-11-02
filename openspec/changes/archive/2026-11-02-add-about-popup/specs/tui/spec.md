# Spec Delta

## ADDED Requirements

### Requirement: About popup

The TUI SHALL provide an about popup that shows the running binary's resolved
version, its Go runtime and platform, the license and copyright, the project
source URL, and credits.

#### Scenario: About popup shows build and project details

- **WHEN** the user opens the about popup
- **THEN** it shows the resolved version, the Go runtime and `GOOS/GOARCH`,
  `MIT` with the copyright line, `github.com/mipmip/hip`, and
  `Pim Snel, et al`

#### Scenario: Resolved version matches the CLI

- **WHEN** the binary was built with a version override
- **THEN** the popup shows that resolved version, matching `hip version`,
  rather than the embedded fallback

#### Scenario: About popup opens from every level

- **WHEN** the user presses `?` outside the filter and outside another modal
- **THEN** the popup opens at the current navigation level and TUI mode

#### Scenario: About popup closes and swallows other keys

- **WHEN** the popup is open and the user presses `?` or `esc`
- **THEN** it closes; any other key is ignored

#### Scenario: Filter input keeps the question mark

- **WHEN** the filter is active and the user types `?`
- **THEN** the character is added to the query and the popup does not open

#### Scenario: Built-in footers advertise the about key

- **WHEN** a built-in list, detail, or multiplex footer is shown
- **THEN** it includes `?: about`
