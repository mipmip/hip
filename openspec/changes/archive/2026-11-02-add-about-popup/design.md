# Design

## Context

The TUI already has three centered modals (clone progress, collection picker,
rename/delete prompt). All key handling flows through `Model.handleKey`, with
each modal owning the keyboard while open. `View` checks modal state before the
list and detail views. `centerBox` provides the shared bordered, centered frame.

The CLI resolves its version through `resolveVersion()` in `cmd/hip`: the
ldflags value wins, otherwise the embedded `VERSION` file is used. The TUI has
no version field today, and `tui.Run` has no way to receive the resolved value.

## Goals / Non-Goals

**Goals:**

- One about modal with the five requested fields.
- The same resolved version the `hip version` command prints.
- A single global key that works from every level and mode without disturbing
  the filter or other modals.
- Footer hints that make the key discoverable.

**Non-Goals:**

- Commit hash, build date, or VCS provenance.
- Full license text or a scrollable license view.
- Opening the GitHub URL from the popup.
- A general help screen or configurable about content.

## Decisions

### Modal state and rendering

Add `about bool` and `version string` to `Model`, plus an `aboutView` method.
`View` returns it after the existing clone, picker, and prompt checks, before
the detail and list views. `aboutView` builds a short body and passes it to
`centerBox`, matching the other modals.

The body is static except for the version:

```
hip 1.4.0
version   1.4.0
build     go1.26.8 linux/amd64
license   MIT, (c) 2026 Pim Snel
github    github.com/mipmip/hip
credits   Pim Snel, et al

? or esc: close
```

The build line comes from `runtime.Version()`, `runtime.GOOS`, and
`runtime.GOARCH`. No `debug.ReadBuildInfo` call is used: Nix builds without a
`.git` directory would produce inconsistent VCS fields, and the proposal
excludes provenance.

### Resolved version injection

Change `tui.New(cfg)` to `tui.New(cfg, version)` and `tui.Run(cfg, ...)` to
`tui.Run(cfg, version, ...)`. `cmd/hip` passes `resolveVersion()`; the smoke test
passes a fixed test version. This makes the resolved value explicit at the only
production entrypoint instead of importing the root `hip` package, whose
`Version` is only the embedded fallback.

Alternative considered: import `github.com/mipmip/hip` and default to
`hip.Version`. Rejected because a release built with `-X main.version` would
show the wrong value.

### Key routing

`handleKey` checks modal state in this order: clone, picker, prompt, about.
While `about` is true, `?` and `esc` close it and every other key is ignored.
The filter branch stays before the about toggle, so `?` is typed into the query
while filtering. Outside the filter, a `?` check before the detail/list branch
opens the popup, making the key global without duplicating it in
`handleDetailKey`.

### Footer hints

Add `?: about` to `footerText`, `detailFooterText`, and the `switch_hint`
element used by multiplex mode. Custom mode headers and footers remain
unchanged; the key still works there, and no new chrome element or config
schema is introduced.

## Risks / Trade-offs

- **`?` may later be wanted for a help overlay** -> If that happens, move about
  to `i` and leave `?` for help; the modal state is independent of the key.
- **Tiny terminal clips the modal** -> The body is eight lines; `centerBox`
  centers what fits, consistent with the existing modals.
- **A future call site forgets the version argument** -> The constructor takes
  it explicitly, so the code does not compile without a value.
- **Custom footers do not advertise the key** -> The key remains global; only
  discoverability varies by user configuration.

## Migration Plan

None. No configuration, cache, CLI, or provider behavior changes.

## Open Questions

None.
