# Alma (Go migration)

The Go CLI is being built in steps tracked by [PLAN.md](PLAN.md). The current
executable is a bootstrap only: it reports that command dispatch is not yet
implemented and exits nonzero. It does not read configuration or modify files.

Module path: `github.com/ADIX7/alma2`. Prerequisite: Go 1.23 or newer. From the
repository root:

```sh
go build -o alma ./cmd/alma
go test ./...
```

The build command writes `alma` to the current directory; use `-o` with another
path if desired. See [MIGRATION.md](MIGRATION.md) for the agreed YAML-only
behavioral contract (`repository.yml` and `.alma.yml`).
