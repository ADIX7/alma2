# Alma (Go migration)

The Go CLI is being built in steps tracked by [PLAN.md](PLAN.md). It currently
supports `help`, `ls`, `link`, `install`, `configure`, `readme`, all `info`
views, and a read-only `link -d`/`link --dry-run` preview. `ls` reads `repository.yml` and lists direct child modules with
`.alma.yml`. Real linking creates symlinks in selected destinations and stores Go-only YAML
history under `<app-data>/history-v1/`; legacy history files are not read or
modified. `install` and `configure` execute explicitly selected, trusted
single-line shell commands—inspect YAML before running either. Other recognized
commands still report that they are not implemented and exit nonzero.

Module path: `github.com/ADIX7/alma2`. Prerequisite: Go 1.23 or newer.
Configuration parsing uses the pinned `gopkg.in/yaml.v3` dependency (see
`go.mod`/`go.sum`); a clean build needs it in the module cache or access to a
Go module proxy. From the repository root:

```sh
go build -o alma ./cmd/alma
go test ./...
```

The build command writes `alma` to the current directory; use `-o` with another
path if desired. See [MIGRATION.md](MIGRATION.md) for the agreed YAML-only
behavioral contract (`repository.yml` and `.alma.yml`).

Tests can use `internal/testutil.NewFixture(t)` for disposable home, config,
app-data, repository, target, work, binary, and temp directories. Its `RunCLI`
helper invokes an explicit absolute executable directly from the fixture work
folder, captures stdout/stderr and exit status, imposes a timeout, and supplies
only fixture-valued environment variables (except Windows `SystemRoot`). `PATH`
points only to the fixture binary folder; no inherited user configuration or
shell is used. Place fake commands there and use harmless fixture inputs: never
execute scripts from real dotfiles or write to the operator's home/config/data.
