## ADDED Requirements

### Requirement: Filter-active launch

The TUI SHALL support launching directly into the combined cross-provider flat
view with the fuzzy filter already active and focused, via a `--filter` launch
option. Because the fuzzy filter only means "search repositories" at the repos
level, `--filter` SHALL imply the flat combined view (equivalent to passing the
flat-list option as well); the first keystroke after launch narrows the
repository list. The filter-active launch SHALL reuse the existing filter input,
key handling, and mode chrome unchanged.

#### Scenario: Filter-active launch opens the flat list filtering

- **WHEN** the TUI is launched with the filter option (`--filter`)
- **THEN** it opens in the combined flat repository list with the fuzzy filter
  active and focused, so typing immediately narrows the list without pressing
  `/` first

#### Scenario: Filter option implies the flat list

- **WHEN** the TUI is launched with `--filter` but without the flat-list option
- **THEN** it opens in the combined flat repository list (as if the flat-list
  option were also given), not the provider list

#### Scenario: Composes with multiplex mode

- **WHEN** the TUI is launched with both `--filter` and `--mode multiplex`
- **THEN** it opens the flat list filtering in multiplex mode, so the user can
  type a repository name and press Enter to clone-if-needed and run the
  `switch_command`

#### Scenario: First Esc stops filtering, second Esc goes one level up

- **WHEN** the TUI was launched filter-active and the user presses Esc
- **THEN** the first Esc clears the query and stops filtering, leaving the user
  in the combined flat repository list, and a second Esc backs the flat view out
  to the provider list
