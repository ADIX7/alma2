#!/bin/sh
# POSIX fixture-only rehearsal: no real dotfiles or configured shell commands.
set -eu
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
sandbox=$(mktemp -d)
trap 'rm -rf "$sandbox"' EXIT HUP INT TERM
for dir in home config/alma data repository/mod/editor repository/mod/folder repository/mod/scratch target work bin; do
  mkdir -p "$sandbox/$dir"
done
cp "$project/internal/e2e/testdata/repository.yml" "$sandbox/config/alma/repository.yml"
cp "$project/internal/e2e/testdata/mod/.alma.yml" "$sandbox/repository/mod/.alma.yml"
printf 'editor\n' > "$sandbox/repository/mod/editor/config"
printf 'child\n' > "$sandbox/repository/mod/folder/child"
printf 'excluded\n' > "$sandbox/repository/mod/scratch/skip"
printf 'unmapped\n' > "$sandbox/repository/mod/unmapped"
printf '# Fixture README\n' > "$sandbox/repository/mod/README.md"
go build -o "$sandbox/bin/alma" "$project/cmd/alma"
run() (
  cd "$sandbox/work"
  env -i HOME="$sandbox/home" XDG_CONFIG_HOME="$sandbox/config" XDG_DATA_HOME="$sandbox/data" \
    ALMA_CONFIG="$sandbox/config" ALMA_APP_DATA="$sandbox/data" ALMA_WORKDIR="$sandbox/work" \
    PATH="$sandbox/bin" "$sandbox/bin/alma" "$@"
)
run ls
run ls fixtures
run info fixtures
run readme fixtures mod
run link fixtures mod --dry-run
if test -e "$sandbox/data/alma" || test -e "$sandbox/target/renamed"; then
  echo 'dry run wrote to sandbox' >&2
  exit 1
fi
run link fixtures mod
for name in renamed/config mapped-folder unmapped; do
  test -L "$sandbox/target/$name" || { echo "missing symlink: $name" >&2; exit 1; }
done
test ! -e "$sandbox/target/scratch"
test ! -e "$sandbox/target/README.md"
test ! -e "$sandbox/target/.alma.yml"
test "$(find "$sandbox/data/alma/history-v1" -name '*.yml' | wc -l)" -eq 1
repeat=$(run link fixtures mod)
printf '%s\n' "$repeat"
test "$(printf '%s\n' "$repeat" | grep -c 'Item already exists:')" -eq 3
# A quoted home-only YAML target stays inside the disposable home.
mkdir -p "$sandbox/repository/home-mod"
printf 'platforms:\n  - platform: default\n    target: "~"\n    links:\n      - {source: "space file", target: "space dir/file"}\n' > "$sandbox/repository/home-mod/.alma.yml"
printf 'home payload\n' > "$sandbox/repository/home-mod/home-only"
printf 'spaced payload\n' > "$sandbox/repository/home-mod/space file"
run link fixtures home-mod --dry-run
test ! -e "$sandbox/home/home-only"
test ! -e "$sandbox/home/space dir"
run link fixtures home-mod
test -L "$sandbox/home/home-only"
test -L "$sandbox/home/space dir/file"
test "$(find "$sandbox/data/alma/history-v1" -name '*.yml' | wc -l)" -eq 2
echo 'Fixture rehearsal passed; sandbox removed on exit.'
