# Alma Go migration plan

## Goal

Rewrite the C# Alma CLI in Go, preserving the agreed functionality described in
[MIGRATION.md](MIGRATION.md), with YAML-only configuration:
`repository.yml` and `.alma.yml`. No JSON configuration fallback.

This file tracks implementation across agent sessions. Creating this plan does
not resolve open behavior decisions or authorize changes to real dotfiles.

## Session protocol

1. Read this file, `MIGRATION.md`, and applicable repository instructions.
2. Inspect the working tree; preserve existing user/agent work.
3. Choose one unchecked task whose dependencies are complete. Each task is
   intended to fit one agent session, including focused tests and handoff.
4. Check behavior against approved decisions. Ask the user when policy is
   unresolved; do not silently fix legacy defects or add features.
5. Implement only the selected scope and its tests. If the task cannot fit,
   split it into smaller stable IDs before continuing; document remaining work.
6. Run focused tests plus available project-wide checks. Record actual results,
   including skipped/platform-dependent tests; do not claim unrun checks passed.
7. Mark `[x]` only when all completion checks pass. For partial work, leave `[ ]`
   and update the handoff section with files, tests, blockers, and next action.
8. Update documentation when an approved decision changes the contract. Do not
   mark downstream tasks complete merely because scaffolding exists.

Use temporary fixture roots for home, config, app data, repositories, and link
targets. Never invoke install/configure strings copied from real dotfiles during
validation. Keep `/home/adam/Workshop/Alma-Go` and `/home/adam/dotconfig` read-only.
No real-dotfile rollout, destructive unlink, or legacy history import is included
without separate approval.

## Progress

All tasks start unchecked. `S01` gates behavior-dependent implementation, but
project scaffolding and isolated test helpers can proceed independently.

### Foundation and contract

- [x] **S01 — Resolve migration behavior policies**
  - **Depends on:** none.
  - **Scope:** review `MIGRATION.md` §8 with the user. Record approved choices
    for mapping, merges, platform applicability, paths, dry-run side effects,
    conflicts, validation/exit codes, scripts, and legacy metadata-file handling.
    Specify deferred features explicitly. Settle YAML null/default semantics,
    malformed/unknown-field handling, and history compatibility needs.
  - **Done when:** `MIGRATION.md` distinguishes approved behavior from legacy
    observations; each open item is resolved or explicitly deferred with affected
    tasks identified. Do not check this off on recommendations alone.

- [x] **S02 — Bootstrap the Go executable**
  - **Depends on:** none.
  - **Scope:** initialize the Go module and minimal `alma` entry point; document
    the module path/toolchain choice and reproducible local build/test commands.
    Do not implement application behavior yet.
  - **Done when:** the executable builds and `go test ./...` succeeds from a
    clean checkout with documented prerequisites.

- [x] **S03 — Add isolated filesystem/CLI test helpers**
  - **Depends on:** S02.
  - **Scope:** temporary fixture repositories, target roots, controlled
    environment/home/data values, captured output, and safe subprocess helpers.
    Define how tests avoid inherited user configuration and real shell effects.
  - **Done when:** helper tests prove fixture cleanup and environment isolation;
    tests cannot write into the operator's home/config/data directories.

- [x] **S04 — Implement command dispatch and argument handling**
  - **Depends on:** S01, S03.
  - **Scope:** command registry, `help`/`--help`/`-h`, no/unknown-command cases,
    positional conventions, and approved output/exit behavior. Unimplemented
    handlers must report that honestly rather than simulate success.
  - **Done when:** table-driven tests cover command case sensitivity, help
    ordering, missing arguments, flags, and extra arguments per the contract.

### Configuration and context

- [x] **S05 — Resolve startup directories and environment precedence**
  - **Depends on:** S01, S03.
  - **Scope:** effective working directory, configuration root, application-data
    root, provider diagnostics data, and approved directory-creation timing.
  - **Done when:** tests cover every provider, invalid/missing values, fallback
    precedence, and side effects without touching real user directories.

- [x] **S06 — Resolve path expressions**
  - **Depends on:** S01, S03.
  - **Scope:** home/Documents tokens, relative and absolute paths, slash handling,
    and any additionally approved token/Windows normalization behavior.
  - **Done when:** tests cover base-directory combination, home-only paths,
    platform-specific cases, and unsupported tokens according to policy.

- [x] **S07 — Load and validate repository YAML**
  - **Depends on:** S01, S02, S03.
  - **Scope:** `repository.yml`, repository entries, optional paths, missing-file
    defaults, and approved validation. Select/document the YAML dependency.
  - **Done when:** tests cover valid examples, empty/missing/malformed files,
    invalid entries, null/type handling, and refusal to read JSON or unapproved
    filename aliases. Preserve repository declaration order.

- [x] **S08 — Load and validate module YAML**
  - **Depends on:** S07.
  - **Scope:** `.alma.yml` with ordered `platforms`, link pairs, targets,
    exclusions, README control, install/configure strings, and field presence
    needed for defaulting/merging.
  - **Done when:** tests cover defaults versus explicit values, malformed types,
    quoted home paths versus YAML null, `%` paths, script strings, and JSON-only
    modules. Parsing does not execute scripts or create links.

- [x] **S09 — Detect compatible platform identifiers**
  - **Depends on:** S01, S03.
  - **Scope:** OS identifiers, legacy-compatible architecture names, Linux
    distribution detection, and approved handling of quoted/missing `ID` values.
  - **Done when:** fixture-driven tests cover Linux distributions, missing or
    unreadable release information, Windows/macOS/FreeBSD identifiers, unknown
    platforms, and Go-to-legacy architecture-name translation.

- [x] **S10 — Select and merge platform settings**
  - **Depends on:** S08, S09.
  - **Scope:** selector matching, declaration order, scalar precedence, link
    combination, exclusions/README settings, and no-matching-section behavior.
  - **Done when:** tests cover default + OS + distro + architecture sections,
    duplicate sources, explicit false values, nulls, and approved merge rules.

- [ ] **S11 — Resolve repository and module operation context**
  - **Depends on:** S04, S05, S06, S07, S10.
  - **Scope:** named repositories, working-directory fallbacks, nested module
    names, final destination roots, unknown names, and approved boundary rules.
  - **Done when:** tests cover one/two positional forms, absent path fields,
    relative paths, target overrides, missing directories, and unsafe paths.

### Discovery and link workflow

- [ ] **S12 — Implement repository/module discovery and `ls`**
  - **Depends on:** S11.
  - **Scope:** repository names and recursive YAML-configured module discovery,
    including nested/root modules and approved hidden-directory behavior.
  - **Done when:** CLI tests cover empty repositories, unknown names, missing
    sources, nested modules, slash-separated output, and JSON-only directories.

- [ ] **S13 — Plan automatically discovered links without writing**
  - **Depends on:** S11.
  - **Scope:** recursive file candidates, hidden files, unmapped directories,
    empty directories, and destination-relative structure. Keep planning separate
    from filesystem mutation; define traversal of existing source symlinks per
    approved safety policy.
  - **Done when:** fixture tests verify source/target candidates and termination
    for applicable symlink cases; planning creates no directories or links.

- [ ] **S14 — Apply explicit file and directory mappings**
  - **Depends on:** S13.
  - **Scope:** renamed/absolute/home-relative targets, approved source matching,
    and single-symlink directory mappings that suppress child expansion.
  - **Done when:** tests cover nested paths, repeated basenames, directory
    mappings, target-root overrides, and approved collision behavior.

- [ ] **S15 — Apply configuration and payload exclusions**
  - **Depends on:** S14.
  - **Scope:** `.alma.yml`, approved legacy-config exclusions, source subtrees,
    default/explicit README settings, and directory-link exclusion limitations.
  - **Done when:** tests cover configured/unconfigured modules, mixed JSON/YAML
    files, all recognized README names, nested exclusions, and merged settings.

- [ ] **S16 — Wire `link` dry run and conflict reporting**
  - **Depends on:** S04, S15.
  - **Scope:** `link --help`, `-d`/`--dry-run`, complete candidate output,
    destination checks, and approved dry-run/startup side-effect policy.
  - **Done when:** CLI tests verify proposed links, existing files/directories,
    ordinary/broken symlinks, missing sources, flags, and exact allowed effects.
    No symlink or history write occurs in a dry run.

- [ ] **S17 — Create links and handle filesystem failures**
  - **Depends on:** S16.
  - **Scope:** real file/directory symlinks, approved parent-directory creation,
    non-overwrite behavior, partial failures, and actionable platform errors.
    Return successful creations for subsequent history persistence.
  - **Done when:** temporary-filesystem tests verify link destinations/content,
    preservation of existing items, permissions/failure paths, and approved
    stop/continue semantics. Permission-dependent skips are explicit.

- [ ] **S18 — Persist and integrate successful link history**
  - **Depends on:** S17.
  - **Scope:** define the new history format; source/destination pair identity,
    deduplication, prior records, partial successes, and `link` integration.
    Do not add old-format imports unless separately approved.
  - **Done when:** tests verify repeated runs, separate target roots, skipped
    conflicts, zero/partial successes, persistence failures, and dry-run absence
    of history mutations. Document the chosen storage contract.

### Information and utility commands

- [ ] **S19 — Implement general and repository `info`**
  - **Depends on:** S12.
  - **Scope:** version, directories, platform, repository availability,
    configured module counts, sorting, and install/configure capabilities.
  - **Done when:** CLI tests cover no config, empty/root/nested modules, missing
    source folders, explicit mapping counts, and approved legacy display fixes.

- [ ] **S20 — Implement module `info`**
  - **Depends on:** S19, S14.
  - **Scope:** resolved module settings, target display, explicit mappings,
    capability indicators, and approved existence/color behavior.
  - **Done when:** tests cover absent configs, missing source/target items,
    remapped destinations, relative roots, and distinction between existence
    and actual symbolic-link ownership/correctness.

- [ ] **S21 — Implement safe-to-test shell execution support**
  - **Depends on:** S01, S03, S05.
  - **Scope:** approved Linux/Windows shells, cwd, stdin/stdout/stderr, lifecycle,
    unsupported platforms, and exit-code handling. Use fake runners for unit
    tests; limit process smoke tests to harmless fixture commands.
  - **Done when:** tests verify shell arguments, effective cwd, output, start
    failure, child failure, and approved multiline/interaction restrictions.

- [ ] **S22 — Wire `install` and `configure`**
  - **Depends on:** S11, S21.
  - **Scope:** selected merged commands, missing module/config/command cases,
    command display, execution, and approved multiline validation.
  - **Done when:** CLI tests prove only the requested command runs, no implicit
    linking or install/configure chaining occurs, and real dotfile scripts are
    never executed by tests.

- [ ] **S23 — Implement `readme`**
  - **Depends on:** S11.
  - **Scope:** approved filename behavior, search priority, case handling,
    plain-text printing, and no-config support.
  - **Done when:** tests cover each supported filename, priority collisions,
    missing files/modules, nested module paths, and literal Markdown output.

- [ ] **S24 — Implement diagnostics and explicit unsupported `unlink`**
  - **Depends on:** S04, S05, S06.
  - **Scope:** diagnostic help/list/provider values and special-path values;
    recognize `unlink` but report unsupported behavior using approved errors.
  - **Done when:** CLI tests cover helper ordering, missing/unknown helpers,
    null provider values, and prove `unlink` never deletes filesystem items.

### Verification and migration documentation

- [ ] **S25 — Complete end-to-end migration coverage**
  - **Depends on:** S18, S20, S22, S23, S24.
  - **Scope:** map every `MIGRATION.md` §7 scenario and approved §8 decision to
    tests. Add sanitized, corrected fixture examples representative of
    `~/dotconfig` modules; leave original WIP files unchanged.
  - **Done when:** subprocess-level tests cover the full config-to-plan-to-link
    workflow, repeat runs, history, YAML-only discovery, and failure cases.
    Record any remaining coverage gaps explicitly.

- [ ] **S26 — Verify supported platforms and repeatable checks**
  - **Depends on:** S25.
  - **Scope:** automated formatting, vetting, tests, and build checks; add CI
    appropriate to the repository. Verify Linux and Windows behavior natively
    where available; build other agreed targets without implying runtime support.
  - **Done when:** clean-checkout commands and CI are documented and pass;
    native tests, cross-compilation, permissions-dependent skips, and untested
    platforms are clearly distinguished. Shell restrictions remain explicit.

- [ ] **S27 — Publish usage guide and rehearse migration safely**
  - **Depends on:** S26.
  - **Scope:** populate `README.md` with build/use instructions, valid YAML
    examples, environment precedence, deliberate legacy differences, limitations,
    and a migration checklist. Rehearse on disposable copies/targets only.
  - **Done when:** documented commands work against fixtures; quoted paths,
    exclusions, conflicts, dry-run effects, history, and script trust are covered.
    `MIGRATION.md`, tests, and implementation agree. Real rollout remains gated
    on separate user authorization.

## Handoff / blockers

Update this section at the end of each implementation session. Preserve useful
history as dated entries; replace stale next-action guidance explicitly.

### Initial state

- Completed: functional inventory in `MIGRATION.md`; this checklist created.
- Implementation: not started; no Go module yet.
- Next actions: S01 user policy review, or independent S02 scaffolding.
- Blockers: behavior-dependent tasks require decisions from `MIGRATION.md` §8.
- Verification: documentation-only checks; no Go build/tests exist yet.

### 2026-11-17 / S01

- Outcome: complete. Approved policies recorded in `MIGRATION.md` §8; no Go implementation or real-dotfile changes.
- Files changed: `MIGRATION.md` (textual-prefix matching, unavailable-module command behavior, source-symlink exclusion, root-module `ls`, and `info` fixes); `PLAN.md` (S01 completion and handoff). Earlier approved S01 decisions were already present in the clean baseline.
- Checks: S01 documentation assertions and `git diff --check` passed. No Go module or executable tests exist yet; no executable tests were run or claimed to pass.
- Decisions: all S01 policy gates resolved in `MIGRATION.md` §8; old-history import and `unlink` deferred explicitly. Unspecified legacy behavior is not silently changed.
- Next ready step: S02 (Go executable bootstrap). Do not begin it in this session.

### 2026-11-17 / S02

- Outcome: complete. Bootstrapped `github.com/ADIX7/alma2` with a Go 1.23 module and an `alma` entry point that explicitly fails until command dispatch is implemented; no configuration or filesystem behavior added.
- Files changed: `go.mod`, `cmd/alma/main.go`, `README.md` (prerequisite and reproducible build/test commands), `PLAN.md` (progress and handoff).
- Checks: `gofmt` check, `go build -o <temporary path> ./cmd/alma`, disposable-binary smoke check (exit 1, expected stderr, empty stdout), `go test ./...`, and `git diff --check` passed. `go test` reported no test files; executable test helpers start in S03.
- Decisions: no new behavior policies; YAML-only contract unchanged. No real dotfiles or C# files touched.
- Next ready step: S03 (isolated filesystem/CLI test helpers). Do not begin it in this session.

### 2026-11-17 / S03

- Outcome: complete. Added disposable fixture roots, guarded fixture-file writes, and direct CLI subprocess execution with controlled environment, cwd, timeout, captured output, and exit status. No application command behavior added.
- Files changed: `internal/testutil/fixture.go`, `internal/testutil/fixture_test.go`, `README.md` (test isolation guidance), `PLAN.md` (progress and handoff).
- Checks: focused `go test ./internal/testutil -count=1 -v` passed (fixture cleanup, home/config/data isolation, captured streams and exit code, invalid executable paths); full `go test ./... -count=1`, `go vet ./...`, `go build ./cmd/alma` to a temporary output, `gofmt` check, and `git diff --check` passed. No skips; `cmd/alma` still has no test files.
- Decisions: no new behavior policies; YAML-only contract unchanged. No real dotfiles or C# files touched.
- Next ready step: S04 (command dispatch and argument handling). Do not begin it in this session.

### 2026-11-17 / S04

- Outcome: complete. Added case-sensitive command selection, ordered help, positional/flag parsing, approved argument errors, and explicit nonzero placeholders for handlers not yet implemented.
- Files changed: `internal/cli/cli.go`, `internal/cli/cli_test.go`, `cmd/alma/main.go`, `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, temporary-binary build and help/unimplemented-handler smoke checks, `gofmt`, and `git diff --check` passed. No skips; `cmd/alma` has no package-local test files.
- Decisions: no new behavior policies; link-specific help remains an explicit placeholder until S16, diagnostics and unlink until S24. No real dotfiles or C# files touched.
- Next ready step: S05 (startup directories and environment precedence).

### 2026-11-17 / S05

- Outcome: complete. Added side-effect-free working/config/app-data provider selection and diagnostics values; explicit app-data preparation skips all writes for dry runs. CLI wiring remains for later command steps.
- Files changed: `internal/startup/startup.go`, `internal/startup/startup_test.go`, `PLAN.md`.
- Checks: focused `go test ./internal/startup -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips.
- Decisions: no new behavior policies; config provider selection tests the base directory, not `repository.yml`. No real dotfiles or C# files touched.
- Next ready step: S06 (path expressions). CLI startup integration must use this resolver and avoid preparation for dry runs.

### 2026-11-17 / S06

- Outcome: complete. Added side-effect-free path expression resolution with injectable special directories, literal home/Documents expansion, Windows-only `%APPDATA%`, relative/absolute paths, and YAML-only backslash rejection.
- Files changed: `internal/pathresolve/pathresolve.go`, `internal/pathresolve/pathresolve_test.go`, `PLAN.md`.
- Checks: focused `go test ./internal/pathresolve -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips; Windows `%APPDATA%` behavior was simulated on Linux, not natively run on Windows.
- Decisions: no new behavior policies; unknown tokens remain literal. No real dotfiles or C# files touched.
- Next ready step: S07 (repository YAML loader). Callers must flag YAML path fields when resolving them.

### 2026-11-17 / S07

- Outcome: complete. Added strict, ordered `repository.yml` loading with missing-file empty default; rejects nulls, unknown/duplicate fields, malformed types, missing names, backslash paths, and multiple documents. Legacy JSON and `.yaml` aliases are never read.
- Files changed: `internal/config/repository.go`, `internal/config/repository_test.go`, `go.mod`, `go.sum`, `README.md` (YAML dependency), `PLAN.md`.
- Checks: focused `go test ./internal/config -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `go mod tidy`, `gofmt`, and `git diff --check` passed. No skips. An initial focused test caught yaml.v3 accepting a numeric scalar into a string; node-type validation fixed it before completion.
- Decisions: pinned `gopkg.in/yaml.v3 v3.0.1`; no JSON fallback or filename aliases. No real dotfiles or C# files touched.
- Next ready step: S08 (module YAML loader).

### 2026-11-17 / S08

- Outcome: complete. Added strict `.alma.yml` parsing with ordered sections, optional-field presence, required link fields, source-boundary and YAML separator checks. Parsing does not execute scripts or create links; JSON-only modules remain unconfigured.
- Files changed: `internal/config/module.go`, `internal/config/module_test.go`, shared strict decoder `internal/config/yaml.go`, refactored `internal/config/repository.go` and its test expectation, `PLAN.md`.
- Checks: focused `go test ./internal/config -count=1`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips. Initial run exposed a changed repository root-type error message after sharing the decoder; test expectation was updated without changing rejection behavior, then all tests passed.
- Decisions: no new behavior policies; multiline shell strings parse but execution remains restricted to S21/S22. No real dotfiles or C# files touched.
- Next ready step: S09 (platform identifiers).

### 2026-11-17 / S09

- Outcome: complete. Added OS/distro/process-architecture identifiers with fixture-driven `/etc/os-release` parsing, quoted/unquoted IDs, missing/unreadable fallback, and Go-to-legacy architecture names.
- Files changed: `internal/platform/platform.go`, `internal/platform/platform_test.go`, `PLAN.md`. Inspected the C# `OsInformation.cs` read-only to preserve case-insensitive `ID=` lookup.
- Checks: focused `go test ./internal/platform -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips; non-Linux OS identifiers are fixture-simulated, not natively run.
- Decisions: no new behavior policies; approved quoted Linux ID normalization retained. No real dotfiles or C# files changed.
- Next ready step: S10 (platform selection and merging).

### 2026-11-17 / S10

- Outcome: complete. Added textual-prefix selector matching, least-to-most-specific stable merging, later scalar/mapping overrides, appended exclusions, explicit README boolean handling, and configured-module unavailability on no match.
- Files changed: `internal/config/merge.go`, `internal/config/merge_test.go`, `PLAN.md`.
- Checks: focused `go test ./internal/config -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips; YAML null validation remains covered by the loader tests.
- Decisions: no new behavior policies; duplicate link mappings keep their first output position but use the last value. No real dotfiles or C# files touched.
- Next ready step: S11 (repository/module operation context).

### Session entry template

```text
Date / task ID:
Outcome: complete | partial | blocked
Files changed:
Checks run and results (including skips):
Approved decisions / remaining questions:
Next action:
```
