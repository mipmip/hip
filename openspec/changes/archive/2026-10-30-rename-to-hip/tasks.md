## 1. Remote & module

- [x] 1.1 Repoint local `jj`/git remote `origin` to `git@github.com:mipmip/hip.git`
- [x] 1.2 `go.mod` module → `github.com/mipmip/hip`; rewrite every import across all `*.go` (incl. tests + `e2e`)
- [x] 1.3 Move `cmd/hup` → `cmd/hip`; root `version.go` package `huphop` → `hip`; update its importer

## 2. Code rename

- [x] 2.1 CLI/brand strings in `cmd/hip` (+ tests): help/usage/version print `hip`, description mentions "Hub Is Plural"
- [x] 2.2 Config dir `~/.config/huphop` → `~/.config/hip` (`internal/config`, `internal/collections` + fixtures)
- [x] 2.3 Cache dir `~/.cache/huphop` → `~/.cache/hip` (`internal/cache` + fixtures)
- [x] 2.4 Env prefix `HUPHOP_` → `HIP_` in `internal/provider` + all `*_test.go`/e2e
- [x] 2.5 Sweep remaining `huphop`/`hup` strings in `internal/tui` and other packages (titles, messages, temp-dir prefixes) (TUI test fixtures keep `mipmip/huphop` as sample repo data; fuzzy cases depend on it)

## 3. Packaging & tooling

- [x] 3.1 `flake.nix`: description, `pname`/`mainProgram` `hip`, `subPackages = [ "cmd/hip" ]`, `hip-coverage`, test git identity; keep vendorHash
- [x] 3.2 `.goreleaser.yaml` (project name, ids, binary `hip`, main `./cmd/hip`), `.github/workflows/release.yml`
- [x] 3.3 `scripts/coverage.sh` import paths → `github.com/mipmip/hip/...`; `scripts/release.sh` brand/URL strings; check `scripts/ship-change.sh`
- [x] 3.4 If `nix build` reports a vendorHash mismatch, update from the printed hash

## 4. Docs, specs & release

- [x] 4.1 Rebrand `README.md` (HIP, "Hub Is Plural"), `BRIEFING.md`, `CLAUDE.md`, `RELEASING.md`, `config.example.yaml`, `openspec/config.yaml` context
- [x] 4.2 Sweep `hup` command references in main specs outside delta-carried requirements: `openspec/specs/{sync,version-embedding}/spec.md` and the `Refresh command` requirement in `repo-cache` (do NOT touch `openspec/changes/archive/`)
- [x] 4.3 `CHANGELOG.md` `[Unreleased]` entry (release as 2.0.0 via `scripts/release.sh`, which bumps `VERSION`) marked BREAKING with the migration steps (move config dir, rename `HUPHOP_*` env vars, replace `hup` in cron/aliases, optional cache cleanup)
- [x] 4.4 Leave `.beans.yml` prefix `skull2-`; do not rename `.beans/` files

## 5. Verify

- [x] 5.1 No stray refs: `rg -i 'huphop|\bhup\b'` outside `openspec/changes/archive/`, `.beans/` and the CHANGELOG history/migration text is empty
- [x] 5.2 `gofmt -l .` empty; `go vet`, `go build`, `go test ./...` pass; `go run ./cmd/hip version` prints `hip 1.4.0`
- [x] 5.3 `nix build` + `./result/bin/hip version` prints `hip 1.4.0`; `bash scripts/coverage.sh` PASS; `nix flake check` green
- [x] 5.4 `openspec validate rename-to-hip` passes
