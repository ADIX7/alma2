# Alma (Go)

Alma links dotfile modules from a repository into selected destinations.
This migration reads **YAML only**: `repository.yml` and per-module `.alma.yml`.
It never imports legacy JSON or legacy link history. The implemented behavior
and approved differences are specified in [MIGRATION.md](MIGRATION.md); progress
and remaining rollout gates are in [PLAN.md](PLAN.md).

## Build and safe rehearsal

Go 1.23+ is required. `gopkg.in/yaml.v3` is pinned in `go.mod`/`go.sum`;
a clean build needs it cached or a reachable Go module proxy.

```sh
go build -o alma ./cmd/alma
go test ./... -count=1
sh scripts/rehearse.sh
```

The POSIX rehearsal builds a temporary CLI, copies only sanitized
[`repository.yml`](internal/e2e/testdata/repository.yml) and
[`.alma.yml`](internal/e2e/testdata/mod/.alma.yml) fixtures, adds quoted home-only and spaced YAML paths, previews and links **inside a disposable sandbox**,
checks history and repeat conflicts, then removes the sandbox. It does not run `install`/`configure` or use real dotfiles.
Windows behavior is covered by the fixture subprocess tests and native CI.

## Configuration and module layout

A repository contains **one-level modules**: each module is a direct child
directory with a single-component name (including hidden directory names).
The repository root is not a module; directories inside a module are payload,
not nested modules. Neither config file is eligible to become link payload,
even if explicitly mapped. A module without `.alma.yml` is valid and links
its discovered files using destination defaults. A configured module whose
platform sections do not match is unavailable for linking or commands.

`<config-base>/alma/repository.yml` (paths below are relative to an effective
working directory such as the rehearsal's `work/`):

```yaml
repositories:
  - name: fixtures
    repositoryPath: "../repository"
    linkPath: "../target"
```

`<repository>/mod/.alma.yml` (the [tested example](internal/e2e/testdata/mod/.alma.yml)):

```yaml
platforms:
  - platform: default
    links:
      - source: editor/config
        target: renamed/config
      - source: folder
        target: mapped-folder
    exclude:
      - scratch
    excludeReadme: true
```

`editor/config` selects that exact module-relative file, not any other
`config`; `folder` creates **one directory symlink**, not child links.
Unmapped files are also linked recursively, preserving relative paths.
`exclude: [scratch]` removes that source subtree; excluding a child *inside*
an explicitly linked directory cannot hide it—exclude the directory itself.
Configured modules exclude root `README.md`, `README.txt`, and `README` by
default; explicitly set `excludeReadme: false` to include them. Unconfigured
modules do not auto-exclude README files. Source symlinks are skipped.

Matching platform sections merge from least to most specific (`default`, then
OS/distribution/architecture selectors); later equal-specificity sections win
for scalars and mappings, while exclusions append. Optional fields include
`target`, `install`, and `configure`. Paths in YAML use `/`, never `\`.
`target: "~"` means the home directory; bare `target: ~` is YAML null and is
rejected. `%DOCUMENTS%` expands to Documents; `%APPDATA%` expands only on
Windows, to roaming application data. Unknown fields, malformed types, and
explicit YAML nulls are errors. `.yaml` filename aliases are not supported.

## Commands and effects

```text
alma help
alma --version                # build version; no configuration needed
alma ls                       # configured repository names
alma ls fixtures              # direct-child modules, including unavailable ones
alma info                     # paths, platform, repository availability
alma info fixtures            # module summaries
alma info fixtures mod        # resolved module settings/destinations
alma readme fixtures mod      # first README.md, README.txt, README; literal text
alma link fixtures mod -d     # same as --dry-run; no filesystem writes
alma link fixtures mod        # real symlinks and Go history
alma install fixtures mod     # ONLY when its trusted YAML command is intended
alma configure fixtures mod   # ONLY when its trusted YAML command is intended
alma diag list                # diagnostic helper names
```

For module operations, `alma link mod` uses the working directory as the
source repository and its parent as the default destination. Named
repositories use configured paths; unknown names and invalid module names
fail rather than falling back. Commands and names are case-sensitive.
`readme` reads without running a script, even when a configured module is
unavailable. `install` and `configure` run only on explicit request: inspect
and trust the YAML first. They accept **single-line** commands, run
non-interactively with `sh -c` on Linux or `pwsh -c` on Windows in the
effective working directory, and propagate failure; other OSes do not have
shell execution support. `unlink` is unsupported and deletes nothing.

A dry run shows candidates and existing-target conflicts without making
links, target directories, app-data, or history. Real linking preserves
existing destinations (including broken symlinks) and reports skips; it
rejects distinct sources targeting the same final destination **before any
write**. If an individual link fails, later candidates still run, successful
links are recorded, and the command exits nonzero—there is no rollback.
New YAML link history lives under `<app-data>/alma/history-v1/`; legacy
`moduleHashes.txt` and legacy hash records are neither read nor changed.
Go history is not a supported unlink mechanism.

## Environment and supported platforms

- Working directory: first existing `ALMA_WORKDIR`, then `WORKDIR`, then the
  invocation directory. Invalid overrides warn and are skipped.
- Configuration base: first existing `ALMA_CONFIG`, `XDG_CONFIG_HOME`,
  `~/.config`, then `ALMA_CONFIG_FALLBACK`; append `alma/repository.yml`.
  This selects a **base directory**, not a YAML file or the `alma` directory.
  Selecting an existing base with no YAML file does not search the next base.
- App-data base: first existing `ALMA_APP_DATA`, OS-local application-data
  directory, then `ALMA_APP_DATA_FALLBACK`; append `alma`. If all are absent,
  use the working directory. Real operations may create this directory;
  dry runs never do.

For repeatable Linux checks:

```sh
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./... -count=1
go vet ./...
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
go build -o "$tmp/alma" ./cmd/alma
GOOS=windows GOARCH=amd64 go build -o "$tmp/alma.exe" ./cmd/alma
GOOS=darwin GOARCH=arm64 go build -o "$tmp/alma-darwin" ./cmd/alma
GOOS=freebsd GOARCH=amd64 go build -o "$tmp/alma-freebsd" ./cmd/alma
```

[CI](.github/workflows/ci.yml) tests/vets/builds on native Ubuntu and Windows,
and runs the disposable rehearsal on Ubuntu and in a Dockerized Linux job using [`Dockerfile.ci`](Dockerfile.ci)
and [`scripts/ci.sh`](scripts/ci.sh). Run that job locally with:

```sh
docker build -f Dockerfile.ci -t alma2-ci .
docker run --rm alma2-ci
```

The native Linux, native Windows, and Docker jobs
[passed with fixture rehearsal](https://github.com/ADIX7/alma2/actions/runs/44154244613).
Linux tests include native shell/symlink coverage; Windows includes native
PowerShell and history checks. Windows symlink tests may skip without creation
privileges; permission tests may skip as root or where permissions cannot be
enforced. macOS and FreeBSD are **cross-built only**, not run natively.
See the [coverage audit](docs/coverage.md) for scenario-to-test references.

## Migration checklist (copies only; no real rollout authorization)

1. Inventory source modules and scripts **without executing them**. Never
   point a test at a real dotfile repository or destination.
2. Copy only reviewed, sanitized payload and YAML into a disposable
   repository; keep one direct-child directory per module. Convert any JSON
   offline into `repository.yml`/`.alma.yml`, then remove JSON from the copy.
3. Set fixture-only home/config/app-data/work/destination directories, or run
   `sh scripts/rehearse.sh` to verify the whole flow automatically.
4. Check `alma ls`, `info`, and `readme`; preview each module with
   `alma link <repo> <module> --dry-run`. Confirm absolute and quoted home
   targets, nested mappings, exclusions, README policy, and any collisions.
5. Only within the disposable fixture, run real `link`; inspect created
   symlinks and Go history, then repeat to confirm conflicts preserve existing
   destinations. Do **not** run copied install/configure scripts.
6. Obtain separate, explicit authorization and a backup/rollback plan before
   any actual home-directory rollout. `unlink` cannot provide rollback.

Tests can use `internal/testutil.NewFixture(t)` for disposable home, config,
app-data, repository, target, work, binary, and temp directories. Its `RunCLI`
helper invokes an absolute binary directly from fixture work, captures exit
status, and supplies only fixture-valued environment variables (except Windows
`SystemRoot`). `PATH` points only to fixture binaries: no inherited user
configuration or shell is used.
