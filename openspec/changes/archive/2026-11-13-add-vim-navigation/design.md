# Design

## Context

See `proposal.md` for motivation and the `tui` delta for the behavior
contract. The list key handling already has `Home`/`End` setting `m.cursor`
directly and `PgDn`/`PgUp` moving by `m.pageStep()`. The detail pane forwards
unclaimed keys to a `bubbles/viewport`, which supports `j`/`k`, `f`/`b`, and
paging but not `g`/`G`, `Home`/`End`, or `Ctrl+F`/`Ctrl+B`.

## Goals / Non-Goals

**Goals:**

- Add the keys at the two places that already own navigation: the list switch
  and `handleDetailKey`.
- Keep filtering behavior unchanged; the filter captures keys before these
  handlers run.
- Keep the footer hints short enough for narrow terminals.

**Non-Goals:**

- Changing the collection picker or making it scroll.
- Adding `h`/`l` as back/enter, `gg`, or other Vim chords.
- Changing viewport internals or the bubbles dependency.

## Decisions

- **Reuse existing movement primitives.** In the list switch, `g` and `Home`
  share the first-row path, `G` and `End` share the last-row path, and
  `Ctrl+F`/`Ctrl+B` share the `PgDn`/`PgUp` paths. This keeps cursor clamping
  and prefetch scheduling identical across keys.
- **Intercept detail keys before the viewport.** `handleDetailKey` handles
  `g`, `G`, `home`, `end`, `ctrl+f`, and `ctrl+b` explicitly, calling
  `m.readme.GotoTop()`, `GotoBottom()`, `PageDown()`, and `PageUp()`. The
  alternative, extending the viewport keymap, would not cover the viewport's
  unexported keymap cleanly and would make the behavior depend on a dependency
  version.
- **Keep `Ctrl+F`/`Ctrl+B` as aliases, not replacements.** Existing `f`/`b`
  keep working. `f` in the list still cycles the fork facet, because detail
  keys are dispatched separately.
- **Leave filter routing untouched.** Filtering is handled before the global
  `?` and detail/list dispatch, so `g` and `G` remain query characters without
  extra conditions.

## Risks / Trade-offs

- **Footer crowding** -> Use compact `g/G: ends` and `^f/^b: page` forms; keep
  existing hints and let narrow terminals truncate as they already do.
- **Key collision in the detail pane** -> `g`/`G` are unbound there today, and
  `Ctrl+F`/`Ctrl+B` are distinct from `f`/`b`, so no existing action is lost.
- **Tests coupling to viewport offsets** -> Assert first/last and page movement
  through the model's viewport offset, matching existing detail tests.

## Migration Plan

None. This is a keybinding addition with no data or configuration changes.
