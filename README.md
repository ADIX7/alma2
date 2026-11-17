# Alma (Go migration)

The Go CLI is being built in steps tracked by [PLAN.md](PLAN.md). It currently
supports command selection and `help`; other recognized commands report that
they are not implemented and exit nonzero. It does not read configuration or
modify files.

Module path: `github.com/ADIX7/alma2`. Prerequisite: Go 1.23 or newer. From the
repository root:

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
