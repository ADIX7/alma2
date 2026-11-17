#!/bin/sh
set -eu
unformatted=$(find . -type f -name '*.go' -exec gofmt -l {} +)
test -z "$unformatted" || { printf 'Unformatted Go files:\n%s\n' "$unformatted" >&2; exit 1; }
go test ./... -count=1
go vet ./...
go build -o /tmp/alma ./cmd/alma
