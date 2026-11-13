# Spec Delta

## MODIFIED Requirements

### Requirement: Paging and jump keys

The TUI SHALL provide keys to move by a screen, half a screen, and to the first
or last row, in addition to single-row movement. In the main navigation lists,
`g` and `G` SHALL jump to the first and last row, and `Ctrl+F` and `Ctrl+B`
SHALL page down and up. In the detail pane, `g`, `G`, `Home`, and `End` SHALL
jump, and `Ctrl+F` and `Ctrl+B` SHALL page down and up as aliases for `f` and
`b`. While filtering, `g` and `G` SHALL remain query characters.

#### Scenario: Page down and page up

- **WHEN** the user presses `PgDn`, `PgUp`, `Ctrl+F`, or `Ctrl+B` in a main
  navigation list
- **THEN** the cursor moves down or up by roughly one visible screen and the
  window follows

#### Scenario: Half-page down and up

- **WHEN** the user presses `Ctrl-D` or `Ctrl-U`
- **THEN** the cursor moves down or up by roughly half a visible screen

#### Scenario: Jump to first and last

- **WHEN** the user presses `g`, `G`, `Home`, or `End` in a main navigation list
- **THEN** the cursor jumps to the first or last row and the window follows

#### Scenario: Detail pane jumps and pages

- **WHEN** the user presses `g`, `G`, `Home`, or `End` in the detail pane
- **THEN** the README scrolls to the first or last line
- **WHEN** the user presses `Ctrl+F` or `Ctrl+B` in the detail pane
- **THEN** the README scrolls down or up by one page

#### Scenario: Filter keeps g and G literal

- **WHEN** the filter is active and the user presses `g` or `G`
- **THEN** the character is added to the query and no jump occurs
