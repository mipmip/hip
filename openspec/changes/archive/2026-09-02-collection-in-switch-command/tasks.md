## 1. Expose the active collection in the switch context

- [x] 1.1 Add `Collection string` to `switchData` (`internal/tui/multiplex.go`)
      with a doc comment stating it is empty outside a collection and that the
      clone-path template deliberately does not get this field
- [x] 1.2 Populate it from `m.selColl` in `renderSwitchCommand`

## 2. Tests

- [x] 2.1 Unit test: with `selColl` set, a `switch_command` referencing
      `{{.Collection}}` renders that collection's name (alongside `{{.Repo}}` and
      `{{.Target}}`)
- [x] 2.2 Unit test: with no collection scope, `{{.Collection}}` renders empty and
      the documented conditional
      (`{{if .Collection}}…{{else}}{{.Short}}->{{.OwnerLower}}{{end}}`) selects the
      fallback branch
- [x] 2.3 Unit test: a collection name containing a space survives `splitArgs` as
      one argument (no injection, no split)
- [x] 2.4 Unit test: the end-to-end multiplex activation for a repo inside a
      collection runs the argv rendered with the collection name

## 3. Documentation and verify

- [x] 3.1 Document `{{.Collection}}` in the README collections section: the
      session-per-collection example and the conditional fallback form
- [x] 3.2 `gofmt -l .` is empty and `nix flake check` passes (build + tests +
      coverage gate)
