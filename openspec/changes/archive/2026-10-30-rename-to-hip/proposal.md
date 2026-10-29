## Why

The tool is used daily and needs a name that lasts. **HIP** ("Hub Is Plural")
is short, nods to `hub`, and its backronym states the multi-forge idea. The
command becomes `hip`, which is easier to type than `hup` and pairs better with
the name. The GitHub repo has already been renamed to `mipmip/hip`.

## What Changes

- **BREAKING** Go module path `github.com/mipmip/huphop` → `github.com/mipmip/hip`.
- **BREAKING** CLI binary/entrypoint `cmd/hup` (`hup <cmd>`) → `cmd/hip`
  (`hip <cmd>`).
- **BREAKING** Auth env-var prefix `HUPHOP_<PROVIDER>_TOKEN` →
  `HIP_<PROVIDER>_TOKEN`.
- **BREAKING** Config dir `~/.config/huphop/` → `~/.config/hip/` (holds
  `config.yaml` and `collections.yaml`) and cache dir `~/.cache/huphop/` →
  `~/.cache/hip/` (XDG-honored).
- Hard cut: no fallback to the old paths or env vars. The CHANGELOG documents
  the manual migration steps.
- Released as **v2.0.0**, because v1 users must migrate config and env.
- Packaging/tooling: `flake.nix`, `.goreleaser.yaml`,
  `.github/workflows/release.yml`, `scripts/release.sh`, `scripts/coverage.sh`,
  root `version.go` package name.
- Docs/config rebranded: `README.md`, `BRIEFING.md`, `CLAUDE.md`,
  `RELEASING.md`, `CHANGELOG.md`, `config.example.yaml`, `openspec/config.yaml`
  context, and `hup`/`huphop` references in the main specs.
- Local `jj`/git remote repointed to `git@github.com:mipmip/hip.git`.

## Capabilities

### New Capabilities
<!-- none: no new behavior -->

### Modified Capabilities
- `configuration`: default config path changes to `~/.config/hip/config.yaml`.
- `collections`: store file moves to `~/.config/hip/collections.yaml`.
- `repo-cache`: cache path changes to `~/.cache/hip/<provider>.json`.
- `provider-clients`: token env-var names change to the `HIP_` prefix.

## Impact

- Every Go import path, `cmd/` directory, test fixtures and e2e tests.
- Existing installs: after upgrading, `hip` finds no config and no tokens until
  the user moves `~/.config/huphop` and renames the `HUPHOP_*` env vars (shell
  profile, cron). The cache is rebuilt by the next refresh.
- Release pipeline produces `hip` binaries; the Nix package and
  `mainProgram` become `hip`.
- No new dependencies; `vendorHash` should stay the same.
