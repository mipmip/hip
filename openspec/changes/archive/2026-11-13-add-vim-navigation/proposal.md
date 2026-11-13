# Proposal

## Why

Navigation uses Vim-style `j`/`k` and half-page keys in lists and the detail
pane, but jumps to the first or last item still require `Home`/`End`, and
paging in lists lacks the familiar `Ctrl+F`/`Ctrl+B` pair. The detail pane
also has no first/last jump at all.

## What Changes

- Add `g`/`G` for first/last in the main navigation lists.
- Add `Ctrl+F`/`Ctrl+B` as page down/page up in the main navigation lists.
- Add `g`/`G`, `Home`/`End`, and `Ctrl+F`/`Ctrl+B` to the detail pane, with
  `Ctrl+F`/`Ctrl+B` aliasing the existing `f`/`b` page keys.
- Update the built-in footer hints and README key table.
- Leave filtering unchanged: `g` and `G` remain query characters there.
- Leave the collection picker unchanged; it renders all rows and does not
  scroll.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tui`: The paging and jump requirement gains Vim first/last and
  `Ctrl+F`/`Ctrl+B` paging in lists and the detail pane.

## Impact

- `internal/tui/update.go`: list and detail key handling.
- `internal/tui/view.go`: footer hints.
- `internal/tui/viewport_test.go`, `internal/tui/detail_test.go`: navigation
  tests.
- `README.md`: key table.
- `openspec/specs/tui/spec.md`: delta for paging and jump keys.
