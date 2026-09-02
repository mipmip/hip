## 1. Collection store package

- [x] 1.1 Create `internal/collections/` with a `doc.go` package comment and the
      data types: `Store{Version int, Collections []Collection}`,
      `Collection{Name string, Members []Member}`,
      `Member{Provider, Owner, Repo string}` (yaml tags `provider`/`owner`/`repo`)
- [x] 1.2 Add `DefaultPath()` resolving `$XDG_CONFIG_HOME/huphop/collections.yaml`
      with the `~/.config/huphop/collections.yaml` fallback, mirroring
      `config.DefaultPath`
- [x] 1.3 Add `Load(path string) (*Store, error)`: a missing file yields an empty
      store with `Version = 1` and no error; a parse failure or duplicate
      collection name returns an error naming the file
- [x] 1.4 Add `Save(path string, s *Store) error` writing atomically (marshal →
      temp file in the same dir → sync → close → rename), creating the directory
      on first write, following `internal/cache/details.go:70-108`
- [x] 1.5 Add member identity helpers: `Member.SameAs(other Member) bool`
      comparing provider exactly and owner/repo case-insensitively, and
      `Collection.Has(Member) bool`
- [x] 1.6 Add the operations: `Get(name)`, `Create(name)`,
      `Add(name string, members []Member) (added, existing int, err error)`
      (idempotent, appends in order), `Remove(name, members)`,
      `Rename(old, new)` (in-place, refuses an existing name),
      `Delete(name)`; `Create`/`Rename` reject empty and duplicate names
- [x] 1.7 Add `internal/collections` to `CORE_PKGS` in `scripts/coverage.sh`

## 2. Collection store tests

- [x] 2.1 Unit test load: missing file → empty store, no file created; malformed
      YAML → error naming the path; duplicate names → error
- [x] 2.2 Unit test round-trip in a temp dir: save → load preserves collection
      order, member insertion order and the version field
- [x] 2.3 Unit test `Add` idempotency: adding a mix of new and existing members
      reports the right added/existing counts and never duplicates a member
- [x] 2.4 Unit test case-insensitive membership (`Org1/Terraform-AWS` matches
      `org1/terraform-aws` on the same provider) and provider-sensitive identity
      (same owner/repo on two providers are two members)
- [x] 2.5 Unit test `Rename` (in-place, members intact, collision refused),
      `Remove` (order preserved, empty collection still exists) and `Delete`
- [x] 2.6 Unit test atomicity: a temp file left in the directory is not loaded as
      a collection file, and `Save` replaces the target by rename

## 3. Select while filtering, basket facet and count

- [x] 3.1 Add `tea.KeyTab` / `tea.KeyShiftTab` handling to the `m.filtering`
      switch in `handleKey` (`internal/tui/update.go`), *before* the
      `m.filter.Update(msg)` call: toggle the highlighted repo and move the
      cursor down / up, leaving the query and the scroll window untouched
- [x] 3.2 Bind `tab` / `shift+tab` in the list-mode key switch to toggle and
      advance / retreat (`space` keeps toggle-in-place)
- [x] 3.3 Extract the toggle+move behavior into a helper on the model reused by
      both key paths; keep it inert when `m.multiplexActive()` and at the
      provider/owner levels (as `toggleSelect` already is)
- [x] 3.4 Add `selected triState` to `facets` (`internal/tui/filters.go`) with
      `active()`, `status()` (`selected:only`) and cycling support
- [x] 3.5 Change `facets.match(r provider.Repo)` to
      `facets.match(r provider.Repo, selected bool)` and update the call site in
      `model.go:548` to pass `it.Selected`
- [x] 3.6 Bind `x` to cycle the selected facet through `applyFacetChange`
- [x] 3.7 Append the global selection count to `indicatorText` when non-zero
      (`41-60 of 213 · 4 selected`) in `internal/tui/view.go`, using
      `selectedItems()` for the total
- [x] 3.8 Update `footerText` with the new keys (`tab: select`, `x: selected`,
      `A: add to collection`)

## 4. Selection and facet tests

- [x] 4.1 Unit test: with the filter focused and a query typed, `Tab` selects the
      highlighted repo, advances the cursor, and leaves `filter.Value()` and the
      scroll offset unchanged; `shift+Tab` toggles and retreats
- [x] 4.2 Unit test: `space` while filtering inserts a space into the query and
      selects nothing
- [x] 4.3 Unit test: select under query A, replace with query B, select again →
      `selectedItems()` returns the union across both queries and across
      providers
- [x] 4.4 Unit test: `Tab`/`shift+Tab` are inert in multiplex mode and at the
      provider/owner levels
- [x] 4.5 Unit test: the `selected` facet narrows to the selection, composes with
      a fuzzy query and another facet, and renders in the facet status line
- [x] 4.6 Unit test: toggling a row off while `selected:only` removes it from the
      visible list
- [x] 4.7 Unit test: the indicator reports the global selection count while a
      single owner's list is in scope

## 5. Add-to-collection modal

- [x] 5.1 Add picker state to the model (target collection list, cursor, an
      optional name `textinput`, and an error string) plus a key gate at the top
      of `handleKey` alongside the existing `m.cloning` gate
- [x] 5.2 Bind `A` to open the picker with the current selection (falling back to
      the highlighted repo when nothing is selected; report and no-op when
      neither exists); inert in multiplex mode
- [x] 5.3 Render the picker as a centered modal over the dimmed list, reusing the
      `lipgloss.Place` overlay approach from the clone popup
      (`internal/tui/view.go:100`): existing collections with member counts, then
      a new-collection entry
- [x] 5.4 Handle picker keys: up/down to move, Enter to commit, Esc to cancel;
      choosing the new-collection entry switches to name entry with Enter to
      create-and-add
- [x] 5.5 On commit, map the selected `repoItem`s to `collections.Member`
      (provider `Name`, owner, repo), call `Add`, persist, set a status line
      reporting added/already-member counts, and clear the selection; on a name
      collision or write error show the error in the modal and change nothing

## 6. Collections navigation and management

- [x] 6.1 Load the store in `New` (tolerating a missing file) and hold it on the
      model together with the selected collection name
- [x] 6.2 Add the `Collections` synthetic row to `providerRows()` after
      `allReposLabel`, and a `levelCollections` level listing collections with
      member counts
- [x] 6.3 Wire `enter`/`goBack`: providers → collections → a collection's repo
      list (flat, cross-provider, `selColl` set) and back out symmetrically,
      clearing the filter on level change as the other levels do
- [x] 6.4 Scope `visibleRepos` to the selected collection's members, resolving
      each member against `reposByProvider`; keep unresolved members as dim
      placeholder rows marked unresolved
- [x] 6.5 Bind `d` inside a collection to drop the selected (or highlighted)
      members, persisting immediately and reporting the outcome
- [x] 6.6 Bind `R` at the collections level to rename via a name-entry modal
      (refusing an existing name) and `d` to delete behind a confirmation modal
- [x] 6.7 Make the footer level-aware for the collections keys, as the detail view
      already is

## 7. Launch into a collection

- [x] 7.1 Add a `collection string` parameter to `tui.Run`
      (`internal/tui/run.go`): when set, resolve the collection and start at its
      repo list, composing with the mode, flat-list and filter options
- [x] 7.2 Add a `--collection` string flag to the `tui` flag set in
      `cmd/hup/main.go`, passing it into `tui.Run`; an unknown name returns an
      error so the command exits non-zero

## 8. TUI and CLI tests

- [x] 8.1 Unit test the picker: open with a selection, commit to an existing
      collection → store updated, selection cleared, status reports counts;
      cancel → store and selection unchanged
- [x] 8.2 Unit test creating a collection from the picker, including the
      name-collision error path leaving the store unchanged
- [x] 8.3 Unit test navigation: providers → `Collections` → a collection lists its
      members across providers; two back navigations return to the provider level
- [x] 8.4 Unit test that an unresolved member renders as a dim placeholder row and
      is not dropped from the store when it is saved
- [x] 8.5 Unit test member removal, rename (success and collision) and delete
      (confirm and cancel) against a temp-dir store
- [x] 8.6 Unit test `Run` with `collection` set: starts in the collection's list,
      composes with `--mode multiplex` and `--filter`, and errors on an unknown
      name

## 9. Documentation and verify

- [x] 9.1 Document collections in `README.md`: the `collections.yaml` location and
      format, the pick-while-filtering keys, the basket facet, and the
      `--collection` launch flag (noting that Home Manager layering and automatic
      taxonomies are not yet implemented)
- [x] 9.2 `gofmt -l .` is empty and `nix flake check` passes (build + tests +
      coverage gate, including `internal/collections` ≥80%)
