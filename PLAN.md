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

- [x] **S11 — Resolve repository and module operation context**
  - **Depends on:** S04, S05, S06, S07, S10.
  - **Scope:** named repositories, working-directory fallbacks, single direct-child
    module names, final destination roots, unknown names, and approved boundary rules.
  - **Done when:** tests cover one/two positional forms, absent path fields,
    relative paths, target overrides, missing directories, and unsafe paths.

### Discovery and link workflow

- [x] **S12 — Implement repository/module discovery and `ls`**
  - **Depends on:** S11.
  - **Scope:** repository names and direct-child YAML-configured module discovery,
    including hidden child directories but excluding root/deeper folders.
  - **Done when:** CLI tests cover empty repositories, unknown names, missing
    sources, direct child modules, and JSON-only directories.

- [x] **S13 — Plan automatically discovered links without writing**
  - **Depends on:** S11.
  - **Scope:** recursive file candidates, hidden files, unmapped directories,
    empty directories, and destination-relative structure. Keep planning separate
    from filesystem mutation; define traversal of existing source symlinks per
    approved safety policy.
  - **Done when:** fixture tests verify source/target candidates and termination
    for applicable symlink cases; planning creates no directories or links.

- [x] **S14 — Apply explicit file and directory mappings**
  - **Depends on:** S13.
  - **Scope:** renamed/absolute/home-relative targets, approved source matching,
    and single-symlink directory mappings that suppress child expansion.
  - **Done when:** tests cover nested paths, repeated basenames, directory
    mappings, target-root overrides, and approved collision behavior.

- [x] **S14a — Align module layout with approved flat repository model**
  - **Depends on:** S11, S12, S14.
  - **Scope:** record direct-child-only module names; correct earlier context and
    `ls` behavior/tests without changing recursive *payload* planning.
  - **Done when:** names with `/` or `.` are refused; `ls` only discovers
    immediate configured children; fixtures confirm root/deeper configs are
    not modules and the full test suite passes.

- [x] **S15 — Apply configuration and payload exclusions**
  - **Depends on:** S14, S14a.
  - **Scope:** `.alma.yml`, approved legacy-config exclusions, source subtrees,
    default/explicit README settings, and directory-link exclusion limitations.
  - **Done when:** tests cover configured/unconfigured modules, mixed JSON/YAML
    files, all recognized README names, nested exclusions, and merged settings.

- [x] **S16 — Wire `link` dry run and conflict reporting**
  - **Depends on:** S04, S15.
  - **Scope:** `link --help`, `-d`/`--dry-run`, complete candidate output,
    destination checks, and approved dry-run/startup side-effect policy.
  - **Done when:** CLI tests verify proposed links, existing files/directories,
    ordinary/broken symlinks, missing sources, flags, and exact allowed effects.
    No symlink or history write occurs in a dry run.

- [x] **S17 — Create links and handle filesystem failures**
  - **Depends on:** S16.
  - **Scope:** real file/directory symlinks, approved parent-directory creation,
    non-overwrite behavior, partial failures, and actionable platform errors.
    Return successful creations for subsequent history persistence.
  - **Done when:** temporary-filesystem tests verify link destinations/content,
    preservation of existing items, permissions/failure paths, and approved
    stop/continue semantics. Permission-dependent skips are explicit.

- [x] **S18 — Persist and integrate successful link history**
  - **Depends on:** S17.
  - **Scope:** define the new history format; source/destination pair identity,
    deduplication, prior records, partial successes, and `link` integration.
    Do not add old-format imports unless separately approved.
  - **Done when:** tests verify repeated runs, separate target roots, skipped
    conflicts, zero/partial successes, persistence failures, and dry-run absence
    of history mutations. Document the chosen storage contract.

### Information and utility commands

- [x] **S19 — Implement general and repository `info`**
  - **Depends on:** S12.
  - **Scope:** version, directories, platform, repository availability,
    configured module counts, sorting, and install/configure capabilities.
  - **Done when:** CLI tests cover no config, empty/direct-child modules, missing
    source folders, explicit mapping counts, and approved legacy display fixes.

- [x] **S20 — Implement module `info`**
  - **Depends on:** S19, S14.
  - **Scope:** resolved module settings, target display, explicit mappings,
    capability indicators, and approved existence/color behavior.
  - **Done when:** tests cover absent configs, missing source/target items,
    remapped destinations, relative roots, and distinction between existence
    and actual symbolic-link ownership/correctness.

- [x] **S21 — Implement safe-to-test shell execution support**
  - **Depends on:** S01, S03, S05.
  - **Scope:** approved Linux/Windows shells, cwd, stdin/stdout/stderr, lifecycle,
    unsupported platforms, and exit-code handling. Use fake runners for unit
    tests; limit process smoke tests to harmless fixture commands.
  - **Done when:** tests verify shell arguments, effective cwd, output, start
    failure, child failure, and approved multiline/interaction restrictions.

- [x] **S22 — Wire `install` and `configure`**
  - **Depends on:** S11, S21.
  - **Scope:** selected merged commands, missing module/config/command cases,
    command display, execution, and approved multiline validation.
  - **Done when:** CLI tests prove only the requested command runs, no implicit
    linking or install/configure chaining occurs, and real dotfile scripts are
    never executed by tests.

- [x] **S23 — Implement `readme`**
  - **Depends on:** S11.
  - **Scope:** approved filename behavior, search priority, case handling,
    plain-text printing, and no-config support.
  - **Done when:** tests cover each supported filename, priority collisions,
    missing files/modules, single-component module names, and literal Markdown output.

- [x] **S24 — Implement diagnostics and explicit unsupported `unlink`**
  - **Depends on:** S04, S05, S06.
  - **Scope:** diagnostic help/list/provider values and special-path values;
    recognize `unlink` but report unsupported behavior using approved errors.
  - **Done when:** CLI tests cover helper ordering, missing/unknown helpers,
    null provider values, and prove `unlink` never deletes filesystem items.

### Verification and migration documentation

- [x] **S25a — Subprocess success-path fixture and lifecycle**
  - **Depends on:** S18, S20, S22, S23, S24.
  - **Scope:** sanitized, corrected YAML examples and a disposable binary test for
    config discovery, dry-run, real links, history, and repeat-run conflicts.
  - **Done when:** a subprocess test proves successful end-to-end behavior and
    read-only dry-run effects without invoking real dotfile scripts.

- [ ] **S25b — Subprocess validation and failure-path matrix**
  - **Depends on:** S25a.
  - **Scope:** malformed/JSON-only configuration, unavailable modules, unsafe
    names, final-plan collisions, partial failure, and command restrictions.
  - **Done when:** subprocess tests verify nonzero failures and exact allowed
    fixture effects for representative rejected workflows.

- [ ] **S25c — Audit scenario/policy coverage**
  - **Depends on:** S25b.
  - **Scope:** map every `MIGRATION.md` §7 scenario and §8 approved policy to
    a test; close gaps or record platform-dependent limitations.
  - **Done when:** a reviewable coverage matrix and full suite checks pass.

- [ ] **S25 — Complete end-to-end migration coverage**
  - **Depends on:** S25a, S25b, S25c.
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

### 2026-11-17 / S11 (blocked: module-name boundary policy)

- Outcome: blocked; S11 remains unchecked. No implementation started. The S11 acceptance check requires unsafe-path handling, but `MIGRATION.md` explicitly approves boundary rejection for configured link sources only, not CLI module names.
- Files changed: `PLAN.md` (this blocker only). Inspected read-only C# `LinkCommand.cs`: it combines the repository source with the CLI module token without a containment check; absolute names and `..` may leave the repository, and symlinked module directories are not explicitly rejected.
- Checks: no development tests required or run for this blocked step; documentation consistency and `git diff --check` should be run after recording this entry. The preceding S10 commit passed its tests.
- Decision needed: approve a boundary rule for CLI module names (absolute names, `..` escapes, and symlinked module directories/components). Do not implement or mark S11 complete until answered.
- Next action: record the user's approved rule in `MIGRATION.md`, implement S11 with fixture tests, run focused and full checks, then commit only if complete.

### 2026-11-17 / S11 (approval received; complete)

- Outcome: complete. User approved rejecting absolute/escaping CLI module names and symlinked module directories/components. Recorded rule in `MIGRATION.md`; added read-only repository/module context resolution with case-sensitive repository lookup, working-directory fallbacks, YAML path expressions, nested and root modules, unavailable settings, and missing/unsafe path errors. The earlier S11 blocker is resolved.
- Files changed: `MIGRATION.md`, `internal/operation/context.go`, `internal/operation/context_test.go`, `PLAN.md` (including pre-existing uncommitted blocker entry).
- Checks: focused `go test ./internal/operation -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. A first focused run exposed a test using `os.IsNotExist` on a wrapped error; updated the assertion to `errors.Is`, then reran all checks. Symlink tests run on this Linux host; platform-native Windows tests deferred to S26. No real dotfiles or C# files touched.
- Next ready step: S12 (repository/module discovery and `ls`).

### 2026-11-17 / S12

- Outcome: complete. `ls` now reads only `repository.yml`, lists configured names in declaration order, resolves named repository source paths, and recursively lists valid `.alma.yml` modules including nested/root (`.`), hidden, and unavailable modules. JSON-only directories are ignored; directory symlinks are skipped; invalid YAML and unknown/missing repository paths fail nonzero.
- Files changed: `internal/operation/discovery.go`, `internal/operation/discovery_test.go`, `internal/cli/list.go`, `internal/cli/list_test.go`, `internal/cli/cli.go`, `internal/cli/cli_test.go`, `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/cli ./internal/operation -count=1`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check`, and a temporary executable `ls fixture` smoke test passed (no app-data write). Initial focused run exposed the old CLI placeholder test for `ls`; updated it to test the still-unimplemented `info` handler and reran all checks. Symlink tests run on Linux, with explicit skips if creation is unavailable. No real dotfiles or C# files changed.
- Next ready step: S13 (automatic link candidate planning).

### 2026-11-17 / S13

- Outcome: complete. Added side-effect-free automatic candidate planning for regular files, preserving nested destination structure and hidden files while traversing unmapped directories. Empty directories yield no links; source symlinks, symlinked root modules, and nonregular items are skipped/rejected without following them. Mapping and payload exclusions remain S14–S15.
- Files changed: `internal/linkplan/automatic.go`, `internal/linkplan/automatic_test.go`, `PLAN.md`.
- Checks: focused `go test ./internal/linkplan -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host; symlink creation tests explicitly skip if unavailable on other platforms. No real dotfiles or C# files touched.
- Next ready step: S14 (explicit file/directory mappings).

### 2026-11-17 / S14 (blocked: distinct sources sharing one destination)

- Outcome: blocked; S14 remains unchecked and no mapping implementation was started. `MIGRATION.md` approves later-wins for duplicate source keys and skipping pre-existing targets, but does not settle two different candidate sources resolving to one destination.
- Files changed: `PLAN.md` (blocker only). Read-only C# `LinkCommand.CreateLinks` indicates sequential creation: on a real run, first candidate creates the destination and later candidates skip it as existing; a dry run can show both. Candidate traversal order comes from directory enumeration, not a specified stable order.
- Checks: no development tests for blocked S14; documentation consistency and `git diff --check` should be run after this note. S13 passed full checks and was committed separately.
- Decision needed: retain sequential first-created-wins/second-conflicts behavior (including both dry-run candidates), or reject a duplicate planned destination before any writes? If retaining sequential behavior, should candidate order be made deterministic (for example lexical source order)?
- Next action: record the approved policy in `MIGRATION.md`, then implement/test S14, run focused and full checks, and commit only when complete.

### 2026-11-17 / S14 (approval received; complete)

- Outcome: complete. User approved rejecting duplicate final destinations before any writes. Recorded policy in `MIGRATION.md`; added module-relative explicit file/directory mappings, last-wins source aliases, mapped directory child suppression, path-expression targets, and separate destination validation for the final candidate set. The earlier S14 blocker is resolved. S15 must filter excluded payloads before calling `ValidateDestinations`; S16 must run that validation before app-data or destination preparation. Existing destination conflicts remain S16/S17.
- Files changed: `MIGRATION.md`, `internal/config/module.go` (export shared source-path validator), `internal/linkplan/automatic.go`, `internal/linkplan/mapped.go`, `internal/linkplan/mapped_test.go`, `PLAN.md` (including pre-existing uncommitted blocker entry).
- Checks: focused `go test ./internal/linkplan ./internal/config -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. An initial vet run found unkeyed external-package test literals; fixed them and reran checks. Symlink tests run on Linux, with explicit skips if unavailable. No real dotfiles or C# files touched.
- Next ready step: S15 (configuration and payload exclusions).

### 2026-11-17 / S15 (blocked: protected config inside mapped directory)

- Outcome: blocked; S15 remains unchecked and no exclusion implementation started. `MIGRATION.md` says always omit `.alma.yml` and `.alma-config.json` from link payloads, but a whole-directory symlink can expose either file nested within it. The same contract says excluding a child cannot remove it from an already planned whole-directory link.
- Files changed: `PLAN.md` (blocker only). No real dotfiles or C# files touched.
- Checks: no development tests for blocked S15; documentation consistency and `git diff --check` should be run after this note. S14 focused/full tests and vet passed before its separate commit.
- Decision needed (superseded by later flat-layout clarification): when an explicitly mapped directory contains a protected config filename anywhere below it, reject that mapping/plan, expand the directory into individual links instead, or permit those files through the whole-directory symlink as an exception to “always exclude”? Ordinary nested `exclude` paths retain the documented whole-directory caveat unless separately revised.
- Next action: record the approved rule in `MIGRATION.md`, implement S15 with fixture tests, validate destinations after exclusions and before writes, run focused/full checks, then commit if complete.

### 2026-11-17 / S14a (flat module layout correction)

- Outcome: complete. User clarified repository modules are single direct child directories, with no module hierarchy or `/` in names; module payload may still contain nested folders. Corrected previous S11/S12 behavior and planned downstream acceptance checks. The S15 blocker is superseded for valid layouts.
- Files changed: `MIGRATION.md`, `PLAN.md`, `internal/operation/context.go`, `internal/operation/discovery.go`, related operation/CLI tests.
- Checks: `go test ./internal/operation ./internal/cli -count=1`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host; symlink tests explicitly skip if creation unavailable.
- Next action: implement S15, keeping the historical S11/S12 handoff notes as records of prior behavior.

### 2026-11-17 / S15 (flat layout clarified; complete)

- Outcome: complete. Flat modules place protected config files at module root; the earlier nested-module blocker is resolved. Added read-only final planning: always filter `.alma.yml`/`.alma-config.json` candidates (including explicitly mapped files), filter exact/descendant source exclusions, apply configured root README default/explicit false, and reject collisions only after filtering. A mapped directory remains one link: excluding a child cannot change that link, while excluding the directory itself removes it. No CLI linking or filesystem writes added.
- Files changed: `internal/linkplan/filtered.go`, `internal/linkplan/filtered_test.go`, `PLAN.md` (including earlier uncommitted blocker entry).
- Checks: focused `go test ./internal/linkplan -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host. No real dotfiles or C# files touched.
- Next ready step: S16 (wire `link` dry run, checking the final plan before any startup/app-data write).

### 2026-11-17 / S16

- Outcome: complete. Wired `link --help` and read-only `-d`/`--dry-run`: resolve repository/module and final excluded plan, reject duplicate final destinations before any writes, list proposals, and report existing file/directory/ordinary/broken-symlink conflicts using `Lstat`. Invalid names/configs, unavailable modules, and missing directories return nonzero. Real link creation stays an explicit nonzero placeholder for S17.
- Files changed: `internal/cli/link.go`, `internal/cli/link_test.go`, `internal/cli/cli.go`, `internal/cli/cli_test.go`, `internal/cli/list_test.go` (renamed shared fixture helper), `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, `git diff --check`, and a temporary CLI smoke test verifying no target/app-data write passed. Symlink tests run on Linux and explicitly skip if creation unavailable. No real dotfiles or C# files touched.
- Next ready step: S17 (real symlink creation and failure handling).

### 2026-11-17 / S17

- Outcome: complete. Real `link` now plans/validates before startup/app-data writes, creates parent directories and file/directory symlinks, preserves existing destinations (including broken symlinks), skips source symlinks/changed sources, continues after per-item failures, and exits nonzero on any failure. The apply layer returns only newly created links for S18 history; no history is persisted yet.
- Files changed: `internal/linkapply/apply.go`, `internal/linkapply/apply_test.go`, `internal/cli/link.go`, `internal/cli/link_test.go`, `internal/cli/cli_test.go`, `internal/cli/list_test.go` (fully fixture-scoped environment for real writes), `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/linkapply ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, `git diff --check`, and temporary-binary real-link/repeat-conflict smoke passed. First CLI run exposed test fixture app-data environment incompleteness; fixed it before full checks. POSIX permission test passed on this host; Windows symlink permission behavior remains untested natively. No real dotfiles or C# files touched.
- Next ready step: S18 (new-format link history and integration).

### 2026-11-17 / S18

- Outcome: complete. Added versioned Go-only per-pair YAML records in `history-v1` named by SHA-256 of absolute source/NUL/destination; retains prior destinations, deduplicates new successes, records zero-link real runs, and writes via a temporary file. Real `link` persists successful creations even after partial link failures; history errors return nonzero without rollback. Old C# root-level history files remain untouched. Dry runs never write history.
- Files changed: `internal/history/history.go`, `internal/history/history_test.go`, `internal/cli/link.go`, `internal/cli/link_test.go`, `MIGRATION.md`, `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/history ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. Native Windows atomic replacement remains unverified until S26. No real dotfiles or C# files touched.
- Next ready step: S19 (general/repository info), with S21/S23/S24 independently ready.

### 2026-11-17 / S19

- Outcome: complete. Added general `info` (Go build version, selected directories, platform, ordered repository availability) and named repository `info` (sorted direct-child configured modules, effective explicit mapping counts, install/configure capability, unavailable marker). Empty repositories and missing sources are handled without panics; module-level `info` remains an explicit nonzero S20 placeholder.
- Files changed: `internal/cli/info.go`, `internal/cli/info_test.go`, `internal/cli/cli.go`, `internal/cli/cli_test.go`, `README.md`, `PLAN.md`. Inspected the C# `InfoCommand` and `VersionService` read-only.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host. No real dotfiles or C# files touched.
- Next ready step: S20 (module info).

### 2026-11-17 / S20

- Outcome: complete. Module `info` now handles unconfigured and unavailable direct-child modules, resolves the final module destination and each explicit mapping target, displays install/configure capability, and colors source/actual target paths by filesystem existence. An ordinary existing file is marked present; coloring is not proof of Alma link ownership. Missing modules/unknown repositories fail nonzero.
- Files changed: `internal/cli/module_info.go`, `internal/cli/module_info_test.go`, `internal/cli/info.go`, `internal/cli/info_test.go`, `internal/cli/cli_test.go`, `README.md`, `PLAN.md`. Inspected the C# `ColorCodes` source read-only.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host. No real dotfiles or C# files touched.
- Next ready step: S21 (safe-to-test shell execution support); S23/S24 independently ready.

### 2026-11-17 / S21

- Outcome: complete. Added injectable runner and real process adapter for trusted, explicitly requested single-line commands: Linux `sh -c`, Windows `pwsh -c`, effective cwd, redirected EOF stdin, forwarded stdout/stderr, propagated child/start/cancellation errors, and unsupported-platform rejection. No CLI command wiring or real dotfile scripts executed.
- Files changed: `internal/shell/shell.go`, `internal/shell/shell_test.go`, `PLAN.md`. Inspected read-only C# `ShellService.cs`.
- Checks: focused `go test ./internal/shell -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. Native `sh` fixture smoke ran on Linux; Windows selection was simulated, native Windows execution remains S26. No real dotfiles or C# files touched.
- Next ready step: S22 (wire install/configure).

### 2026-11-17 / S22

- Outcome: complete. `install` and `configure` now select only their respective merged YAML command, print it, and run in effective working-directory cwd via S21's supported shell. Missing config/command reports without execution; unavailable modules fail; multiline commands fail before execution; child nonzero status propagates. No implicit link, install/configure chaining, or real-dotfile script execution in tests.
- Files changed: `internal/cli/commands.go`, `internal/cli/commands_test.go`, `internal/cli/cli.go`, `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. Initial multiline YAML flow-scalar test folded the newline; changed to a literal block scalar and reran checks. Native process smoke ran on Linux with harmless fixture commands; Windows command selection is simulated by S21 tests, native Windows execution remains S26. No real dotfiles or C# files touched.
- Next ready step: S23 (readme).

### 2026-11-17 / S23

- Outcome: complete. `readme` reads only a direct-child module root and prints the first exact-case `README.md`, `README.txt`, or `README`, literally and without requiring YAML config. Unavailable configured modules remain readable; missing modules and unsafe names fail nonzero.
- Files changed: `internal/cli/readme.go`, `internal/cli/readme_test.go`, `internal/cli/cli.go`, `README.md`, `PLAN.md`.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host. No real dotfiles or C# files touched.
- Next ready step: S24 (diagnostics and explicit unsupported unlink).

### 2026-11-17 / S24

- Outcome: complete. Added `diag --help`, sorted helper list, config/app-data provider values with `<null>` handling, special path values including Windows-only `%APPDATA%`, and nonzero missing/unknown helper errors. `unlink` now explicitly reports unsupported and never removes anything.
- Files changed: `internal/cli/diag.go`, `internal/cli/diag_test.go`, `internal/cli/cli.go`, `internal/cli/cli_test.go`, `internal/cli/list_test.go` (fixture APPDATA isolation), `README.md`, `PLAN.md`. Inspected the C# diagnostic sources read-only.
- Checks: focused `go test ./internal/cli -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. No skips on this Linux host. No real dotfiles or C# files touched.
- Next ready step: S25 (end-to-end coverage and corrected sanitized fixtures).

### 2026-11-17 / S25a

- Outcome: complete. Added sanitized YAML-only `repository.yml` and `.alma.yml` fixture examples plus a built-CLI subprocess test for repository/module listing, info mapping count, literal README, read-only dry run, real file/directory mappings, exclusions, Go history, and repeat-run conflicts. The CLI subprocess inherits only fixture-valued environment variables; no real dotfile shell commands are invoked. S25 parent remains unchecked pending S25b/S25c.
- Files changed: `internal/e2e/testdata/repository.yml`, `internal/e2e/testdata/mod/.alma.yml`, `internal/e2e/e2e_test.go`, `PLAN.md` (split S25 into session-sized S25a/b/c).
- Checks: focused `go test ./internal/e2e -count=1 -v`, full `go test ./... -count=1`, `go vet ./...`, `gofmt`, and `git diff --check` passed. Symlink preflight explicitly skips if unavailable. No real dotfiles or C# files touched.
- Next ready step: S25b (subprocess failure-path matrix).

### Session entry template

```text
Date / task ID:
Outcome: complete | partial | blocked
Files changed:
Checks run and results (including skips):
Approved decisions / remaining questions:
Next action:
```
