# Agent instructions

## Development and tests

- Read `MIGRATION.md` and `PLAN.md` before migration work.
- After completing development changes, always run relevant tests and the full
  available test suite. Once the Go module exists, run `go test ./...`.
- Fix failures before declaring work complete. If tests cannot run, report why,
  record the blocker in `PLAN.md`, and leave the migration step unchecked.
- For documentation-only changes, run applicable documentation checks. If no
  executable tests exist yet, state that explicitly; do not claim tests passed.

## Migration steps and commits

- Complete one session-sized migration step at a time, following its dependencies
  and acceptance checks in `PLAN.md`.
- After a step is finished and its tests pass, update its checkbox and handoff
  notes, then commit the step's changes. Include the progress updates in the commit.
- Do not commit an incomplete step or one with failing/unrun required tests.
- Keep commits scoped to the completed step; do not include unrelated changes.
- Use `jj` when the repository supports it; otherwise use Git.
