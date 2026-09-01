## 1. Thread the filter option through Run

- [x] 1.1 Add a `filter bool` parameter to `tui.Run` (`internal/tui/run.go`),
      updating the doc comment to describe the filter-active launch
- [x] 1.2 When `filter` is true, call `m.startFlat()` (so `--filter` alone still
      opens the flat combined list) even if `flatList` is false, then activate
      the filter input (`m.filtering = true`, `m.filter.Focus()`); ensure it
      also composes when both `filter` and `flatList` are true (idempotent)

## 2. Wire the CLI flag

- [x] 2.1 Add a `--filter` bool flag to the `tui` flag set in
      `cmd/hup/main.go` (help: "start with the fuzzy filter active in the flat
      repo list")
- [x] 2.2 Pass the flag value into `tui.Run(cfg, *mode, *flatList, *filter)`

## 3. Tests

- [x] 3.1 Unit test: `Run` (or an extracted helper on the model) with
      `filter=true` yields a model that is in the flat combined view
      (`flatAll` true, `nav == levelRepos`) with `filtering` true and the filter
      focused
- [x] 3.2 Unit test: `filter=true, flatList=false` still lands in the flat view
      (filter implies flat), and `filter=true` composes with `--mode multiplex`
      (switch command active, multiselect suppressed)
- [x] 3.3 Unit test: from a filter-active start, a first Esc clears the query and
      stops filtering (in the flat list), and a second Esc backs out to the
      provider level; an Enter on the highlighted repo triggers the level-repos
      primary action (multiplex switch when a `switch_command` is configured)

## 4. Verify

- [x] 4.1 `gofmt -l .` is empty and `nix flake check` passes (build + tests +
      coverage gate)
