# Alma Go migration: functional specification

## Goal and scope

Rewrite Alma (Advanced Link Manager Application), a dotfile symbolic-link CLI,
in Go in this repository. Reproduce its useful user-visible functionality,
with **YAML replacing JSON entirely**.

This document describes behavior, not an implementation design. It does not
prescribe Go packages, libraries, architecture, or build tooling.

Sources inspected:

- Existing C# application: `/home/adam/Workshop/Alma-Go`, revision
  `5892210f9d2a6107940c0f0991441832f1623d30`.
- Work-in-progress YAML examples: `/home/adam/dotconfig/*/.alma.yml`.
- Conversion intent: `/home/adam/dotconfig/migrate_alma.py`.

Behavior below is source-derived; the old executable was not run. Existing
failures are documented, not implicitly endorsed as requirements.

### Confirmed migration decisions

- Retain the `alma` CLI concept and existing functional scope.
- Module configuration is named **`.alma.yml`**.
- Repository configuration is named **`repository.yml`**.
- Read YAML only. No JSON fallback or automatic conversion at runtime.
- Use the WIP module schema: a `platforms` sequence containing `platform`
  selectors and `links` sequences of `source`/`target` pairs.
- Describe legacy defects separately; do not silently treat fixes as approved
  behavior changes. Unresolved policy choices appear at the end.

Only `.yml` filenames are confirmed. Supporting `.yaml` aliases is not specified.

## 1. User model

A **repository** is a directory containing dotfile modules. A configured
repository has a name, an optional source directory, and an optional default
link destination.

A **module** is a directory of files and subdirectories to expose through
symbolic links. Modules may be nested; CLI module names use `/`, for example
`shell/fish`.

A module configuration can specify:

- A destination root.
- Explicit file or directory destination mappings.
- Source paths to exclude.
- Whether root README files are excluded.
- An install command and a configure command.
- Different settings for different operating systems, distributions, and
  process architectures.

The `links` collection is **not an allowlist** in the existing application.
Unmapped files are also linked through recursive discovery. An explicitly
mapped directory is linked as one directory rather than expanded into files.

Repository and module names are case-sensitive. Command names are also
case-sensitive.

## 2. Startup, environment, and path handling

### Working directory

Before processing any command, choose the first existing directory from:

1. `ALMA_WORKDIR`.
2. `WORKDIR`.
3. The invocation working directory, if neither override is usable.

An override pointing to a nonexistent directory produces a message and is
skipped. Relative repository paths and shell commands are affected by this
working directory.

### Configuration directory

Search these base-directory providers in order, preferring an existing base:

1. `ALMA_CONFIG`.
2. `XDG_CONFIG_HOME`.
3. The user's `~/.config` directory.
4. `ALMA_CONFIG_FALLBACK`.

Append `alma` to the chosen base. Repository settings therefore live at
`<base>/alma/repository.yml`; `ALMA_CONFIG` is a base directory, not the full
path to a configuration file or necessarily the `alma` subdirectory.

The legacy search tests the base directory, not whether it contains Alma's
configuration. It does not continue searching just because `repository.json`
is absent there. Preserve that distinction when defining YAML discovery.

If no base exists, legacy behavior retains the final provider's value, even
if nonexistent; a null final value means no configuration root. Without a
readable configuration file, the normal missing-file result is an empty
repository list. Invalid configuration is not handled as a missing file.

### Application data

Choose an existing base from:

1. `ALMA_APP_DATA`.
2. The OS user's local application-data directory.
3. `ALMA_APP_DATA_FALLBACK`.

Legacy behavior retains a non-null final fallback even if nonexistent;
otherwise it uses the working directory. Append `alma`, and create the result
if needed. This startup side effect can occur even for informational commands.

Application data stores the record of successfully created links. It is
separate from repository configuration.

### Path expressions

Existing path resolution supports:

- `~`: substitute the user's home directory.
- `%DOCUMENTS%`: substitute the OS user's Documents directory.
- `/`: translate to the native directory separator.
- Absolute paths: use as absolute destinations.
- Relative paths: interpret against the applicable base when one is provided.

Substitutions are literal and can occur anywhere in a path, not just at its
start. A path containing `~` bypasses combination with a supplied base.
There is no general environment-variable expansion in the current CLI.
In particular, `%APPDATA%` appears in the WIP examples but is not currently
supported.

### Repository source and destination selection

For module operations, defaults are:

- Source: the effective working directory.
- Destination: its parent directory.

When a named repository is found, its non-null `repositoryPath` and `linkPath`
replace the corresponding defaults after path-expression resolution.
A relative repository path is relative to the effective working directory.

An unknown repository name silently retains those defaults for module
operations and `info`; `ls <repository>` instead reports the unknown name.

A module's `target` overrides the repository destination. Resolve it relative
to the repository destination unless it is absolute or uses the home shortcut.
Explicit link targets are resolved relative to this final module destination.

## 3. YAML configuration contract

### Repository file

Location: `<configuration-root>/repository.yml`.

```yaml
repositories:
  - name: dotconfig
    repositoryPath: "~/dotconfig"
    linkPath: "~/.config"
  - name: other
    repositoryPath: "~/other-dotfiles"
    linkPath: "~"
```

Fields retain the existing names:

| Field | Meaning |
| --- | --- |
| `repositories` | Ordered repository entries. Missing-file default: empty list. |
| `name` | Name used in CLI arguments and listings. |
| `repositoryPath` | Optional module repository source path. |
| `linkPath` | Optional default destination root. |

Missing source/destination fields retain the operation's working-directory
fallbacks. The legacy `ls` command specifically requires `repositoryPath`.

### Module file

Location: `<repository>/<module>/.alma.yml`.

```yaml
platforms:
  - platform: default
    links:
      - source: lua
        target: lua
      - source: init.lua
        target: init.lua
    exclude:
      - scratch
    excludeReadme: true
  - platform: linux
    target: "~/.config/nvim"
  - platform: windows
    target: "~/AppData/Local/nvim"
    install: winget install Neovim.Neovim
  - platform: linux-alpine
    install: apk add neovim
  - platform: linux-fedora
    install: dnf install neovim
```

| Field | Meaning / default |
| --- | --- |
| `platforms` | Ordered platform-specific sections. |
| `platform` | Platform selector; `default` matches every platform. |
| `target` | Optional destination-root override. |
| `links` | Explicit mappings; omitted means no explicit mappings, not no linking. |
| `links[].source` | Source file or directory inside the module. See legacy matching caveat below. |
| `links[].target` | Destination path, relative to the final module root or absolute. |
| `exclude` | Source-relative paths to omit, including descendants. Default: none. |
| `excludeReadme` | Exclude root `README.md`, `README.txt`, and `README`; default `true`. |
| `install` | Optional shell command executed by `install`. |
| `configure` | Optional shell command executed by `configure`. |

The exclusion fields describe the existing configurable capability; they are
not reliably retained by legacy multi-section merging. See section 8.

### Platform identifiers and merging

The existing identifier includes OS and process architecture:

- `windows-<architecture>`
- `macos-<architecture>`
- `freebsd-<architecture>`
- `linux-<distribution>-<architecture>` when a distribution is available;
  otherwise `linux-<architecture>`
- `unknown` when the OS is unrecognized

Linux distribution comes from the `ID` value in `/etc/os-release`; legacy
behavior lowercases it but does not strip surrounding quotes. Architecture
names are the legacy lowercase names, such as `x64` and `arm64`, not
necessarily Go's architecture names.

A selector matches when it is exactly `default`, or the platform identifier
starts with it, case-insensitively. Prefix matching is textual, not
component-aware. Thus `linux`, `linux-fedora`, and `linux-fedora-x64` may all
match the same machine. The special `default` spelling is case-sensitive.

For matching sections, legacy behavior follows configuration declaration order:

- Later non-null `target`, `install`, and `configure` values win.
- Link mappings are combined; duplicate source keys cause failure rather than
  overriding an earlier value.
- Exclusions and `excludeReadme` are lost/reset when two sections are merged.
- A present config with no matching sections resolves to an empty configuration,
  not an unsupported/disabled module. Linking can still discover its files.

The YAML sequence explicitly preserves declaration order. Do not silently add
specificity-based sorting; that would be another behavior change.

### YAML authoring and WIP migration cautions

Existing examples are drafts, not proof that every file is valid or supported:

- Write `target: "~"`, not `target: ~`; a bare `~` is YAML null.
- Quote paths beginning with `%`, for example `"%DOCUMENTS%/PowerShell"`.
- Prefer forward slashes in paths. Plain YAML strings do not unescape doubled
  backslashes copied from JSON; double-quoted YAML strings have escape rules.
- Quote shell commands where YAML punctuation could change their meaning.
- YAML block scalars can represent multiline strings, but the existing command
  behavior rejects multiline install/configure scripts.
- `%APPDATA%` needs a separately approved expansion rule or a supported path
  replacement such as `"~/AppData/Roaming/..."`.
- The conversion script does not copy `excludeReadme`; review it manually.

Reading JSON is forbidden in the rewrite, but whether old `.alma-config.json`
files should be automatically excluded from linking remains an explicit
migration-safety decision. Many WIP modules contain both filenames.

## 4. CLI reference

### General invocation

```text
alma <command> [arguments]
```

No command prints `No command was given`. An unknown command prints
`Invalid command: <command>`. Neither automatically displays help.

For `link`, `install`, `configure`, and `readme`:

```text
alma <command> <module>
alma <command> <repository> <module>
```

For these commands and `info`, tokens beginning with `-` are removed before
positional parsing. Extra positional arguments after the first two are ignored.
There is no common option parser, `--` terminator, or universal per-command
help behavior. A missing required module prints `No module specified`.

### `help`, `--help`, `-h`

Print all command names, alphabetically:
`configure`, `diag`, `help`, `info`, `install`, `link`, `ls`, `readme`, `unlink`.
Extra arguments are ignored. These aliases apply as the first CLI token.

### `link`

```text
alma link <module>
alma link <repository> <module>
alma link [<repository>] <module> -d
alma link [<repository>] <module> --dry-run
alma link --help
```

`--help` prints link-specific usage and returns. `-h` is not link-specific help.

Workflow:

1. Resolve repository source and destination roots.
2. Report and stop if the source or module directory is missing.
3. Read the module config if present; a config is not required for linking.
4. Apply the module destination override.
5. Create the destination root if missing, **even during a legacy dry run**.
6. Discover candidate links recursively, including hidden files/directories.
7. Omit the active module config, excluded source paths, and root README files
   when README exclusion is enabled.
8. For each candidate, skip an existing destination; otherwise create a symbolic
   link to the source file or directory. Never replace or back up an existing item.
9. Record successful link destinations unless this is a dry run.

Discovery and mapping details:

- Unmapped files retain their relative location beneath the destination root.
- Unmapped directories are traversed; no link is created for the directory itself.
- Explicitly mapped directories produce a single directory symlink and are not
  traversed. Empty unmapped directories produce no links.
- Legacy file mapping lookup uses the file's **basename**, even in nested folders.
  A mapping for `config` can affect multiple nested files named `config`.
- Legacy directory mapping lookup uses a module-relative path with native
  separators. This differs from file lookup and is not portable for all keys.
- Exclusions are exact source paths and descendants, not globs. They filter
  candidates after traversal. Excluding a child cannot remove it from an already
  planned whole-directory symlink.
- With no module config, README files are not automatically excluded. A resolved
  configuration has `excludeReadme: true` by default.
- Link sources are absolute paths. Moving the repository can break them.

Existing target checks consider files and directories, including ordinary links
to existing items. Broken symlinks are not explicitly inspected as conflicts;
behavior depends on platform filesystem checks.

The old application creates only the destination root, **not intermediate
folders** for nested or remapped targets. A resulting filesystem error can
interrupt the remaining links. It catches link-creation I/O failures, reports
them, and retains/records successes made before the failure. On Windows the
message also advises Administrator privileges for symlink creation.
Other exceptions are not generally handled.

Dry run reports proposed links and existing-target conflicts, creates no
symlinks, and does not update link records. Its directory-creation/startup
side effects mean it is not truly read-only in the legacy CLI.

### `unlink`

Advertised and recognized, but currently throws an unimplemented-operation
error. There is no working unlink, restore, or cleanup functionality to port.
Implementing one requires a new behavioral specification.

### `ls`

```text
alma ls
alma ls <repository>
```

Without arguments, print configured repository names in configuration order.
With a repository name:

- Report an unknown name or missing `repositoryPath`.
- Recursively find directories with a resolvable module config.
- Print module paths relative to the repository, using `/` separators.
- Continue into nested directories even after finding a module.

There is no explicit sorting or hidden-directory exclusion. The legacy command
uses `repositoryPath` literally, unlike module commands, so home shortcuts in
that field do not work consistently. A module at the repository root is printed
as an empty relative name. Extra arguments are ignored; flags are not parsed.

### `info`

```text
alma info
alma info <repository>
alma info <repository> <module>
```

General information includes version, application-data directory, configuration
directory or preferred location, platform identifier, and configured repositories.
Missing configured source directories are annotated.

Repository information recursively discovers configured modules, sorts them by
name, and displays the explicit mapping count and whether install/configure
commands exist. The count is not the number of automatically discovered links.

Module information displays its configured target, install/configure availability,
and explicit mappings, or reports no configuration. Mapping output uses terminal
colors for existence indications; it does not verify link ownership/correctness.

Legacy caveats: empty repository results can fail instead of printing an empty
list; a configured repository-root module can fail during name generation;
module target display does not consistently apply the repository destination
base; target coloring checks the source key beneath the module target rather
than the actual mapped destination. Treat the display as informational, not a
reliable link-status command.

### `install` and `configure`

```text
alma install [<repository>] <module>
alma configure [<repository>] <module>
```

Read merged module settings. Report missing configuration or a missing respective
command. Print the selected command, then execute it without confirmation:

- Linux: `sh -c <command>`.
- Windows: `pwsh -c <command>`; PowerShell must already be available.
- Other OSes: report unsupported platform and fail.

Commands run in the effective application working directory, **not** the module
directory. They are trusted shell code and may modify the machine.

Only one line is supported, based on host-native newline splitting. Multiline
commands are reported as unsupported and not executed. Standard output is
forwarded line by line; standard error is inherited. Standard input is redirected,
so these are not normal interactive terminal sessions.

The application waits for completion but does not propagate or otherwise act on
the child exit code. Neither command automatically links files or invokes the
other command.

### `readme`

```text
alma readme [<repository>] <module>
```

Intended functionality: print the first module README in this priority order:
`README.md`, `README.txt`, `README`. Print plain lines; no Markdown rendering.
The search is case-sensitive and does not require module configuration.

Actual legacy defect: it searches for files literally named `Markdown`, `Text`,
and `NoExtension` instead. Standard README filenames normally produce the
message `No README file found. Supported formats: README.md, README.txt, README`.

### `diag`

```text
alma diag --help
alma diag list
alma diag config-home-providers
alma diag app-data-providers
alma diag special-path-resolver
```

`list` prints available diagnostic helper names alphabetically. The helpers show,
respectively, configuration-directory providers, application-data providers,
and special path substitutions with their current values.
Missing or unknown helper names produce a message. Extra helper arguments have
no functional use in the current helpers.

## 5. Output and failures

Ordinary messages and application-logged errors are written to standard output.
There is no structured/JSON output mode, quiet mode, or global verbosity option.

Most validation failures print a message and return normally. There is no
consistent nonzero exit-code contract for these failures. Unhandled failures,
including invalid configuration, unsupported shell execution, and `unlink`,
terminate abnormally. Exact runtime stack traces are not a useful compatibility
requirement.

No automatic rollback occurs after partial linking. There is no force-overwrite,
backup, repository registration command, automatic install-before-link workflow,
or working link-status/unlink command.

## 6. Persistent link history

After a non-dry-run link operation, persist successfully created destinations
associated with the absolute module-source/final-destination pair.

- Retain earlier recorded destinations and deduplicate new additions.
- Do not record skipped preexisting destinations as newly created links.
- Separate histories for different source/destination pairs.
- Records may still be written when no links were created.
- Records are not reconciled against later external filesystem changes.
- No existing command consumes this history to remove links.

The legacy on-disk encoding is not specified here: interoperability with old
application-data files has not been requested. Decide explicitly whether it is
needed rather than coupling the Go rewrite to internal storage details.

## 7. Functional validation scenarios

Use temporary fixture repositories, not the real dotfiles, for migration tests.
These scenarios describe functional expectations; legacy-defect cases need the
policy decisions in section 8 before becoming acceptance tests.

| Scenario | Expected behavior |
| --- | --- |
| No configured repositories | `ls` prints its repository heading and no names. |
| Named repository with nested modules | `ls <repo>` discovers configured descendants and uses slash-separated names. |
| Working-directory override | First valid `ALMA_WORKDIR`/`WORKDIR` controls fallback paths and shell cwd. |
| Module without a config | `link` discovers its files using working-directory defaults. |
| Default + Linux + Fedora settings | All matching sections contribute in declaration order; later scalar values win. |
| Nonmatching platform sections only | Legacy behavior still permits automatic file discovery. |
| Unmapped root file | Link beneath the selected module destination with its original name. |
| Explicit renamed file | Link at the mapped destination, not at the original name. |
| Explicit directory mapping | One directory symlink; no individual child links. |
| Nested unmapped file | Candidate retains relative nesting; legacy fails if its target parent is absent. |
| Excluded source subtree | Remove matching candidates and descendants, subject to whole-directory-link caveat. |
| README exclusion | Configured module defaults to excluding recognized root README files. |
| Existing destination | Preserve it and report a conflict/skip; do not adopt it into newly created history. |
| Dry run | Show candidates, no symlinks/history; document legacy directory side effects. |
| Partial filesystem failure | Earlier successful links remain and are recorded. |
| Install/configure absent | Informative message; no shell command runs. |
| Install/configure present | Execute selected single-line command in effective cwd with OS-specific shell. |
| JSON config only | New CLI must not read it as configuration. |
| Both config formats present | Read `.alma.yml` only; legacy-JSON payload exclusion policy needs a decision. |
| Home-only YAML target | Quoted `"~"` resolves to home; bare `~` is not a home-path string. |
| Standard README | Intended printing differs from the current filename bug. |
| `unlink` | Legacy is unsupported; no removal semantics may be inferred from its name. |

## 8. Decisions to settle before implementation

The format change is approved. The following are not silently approved additions
or fixes; they are behavior choices exposed by the source and WIP configurations.

1. **Defect compatibility:** which legacy defects should be fixed rather than
   reproduced? Recommended: document differences and fix accidental failures.
2. **Link matching:** retain basename-based file lookup, or make every `source`
   a consistent slash-separated module-relative path?
3. **Target folders and dry run:** create necessary parent directories in real
   runs, and make dry runs fully read-only, including startup?
4. **Merge rules:** preserve exclusions/README settings across sections; decide
   whether exclusions append or override, and whether duplicate sources fail
   or later mappings override earlier ones.
5. **Platform applicability:** keep automatic linking when no section matches,
   or mark such modules unavailable? Keep declaration order unless changed
   explicitly.
6. **Legacy files as payload:** recommend always excluding `.alma.yml` and old
   `.alma-config.json` from links, without reading the JSON. Decide whether an
   unconfigured folder containing only old JSON may still be linked.
7. **Path support:** add `%APPDATA%`? Normalize copied Windows separators?
   Resolve repository paths consistently in `ls` and other commands?
8. **Failures and safety:** define validation, nonzero exit codes, broken-symlink
   conflicts, unknown-repository handling, source paths escaping a module,
   and continue-versus-stop behavior after one failed link.
9. **Shell scope:** retain single-line Linux/Windows-only execution, or add
   multiline scripts, interactive input, module-local cwd, and more OSes?
10. **Deferred functionality:** implement `unlink` and/or import old link history
    only with a separate specification covering ownership and deletion safety.

## 9. Migration workflow

1. Settle the relevant open behavior choices above; record deliberate changes.
2. Use this functional inventory to implement the Go CLI and fixture-based tests.
3. Convert repository settings to `repository.yml` and modules to `.alma.yml`.
4. Review converted YAML for quoted home paths, `%` paths, backslashes, omitted
   fields, unsupported expansions, and platform-order dependencies.
5. Validate discovery and candidate destinations against fixture expectations
   and representative WIP modules before allowing writes to real targets.
6. Compare dry-run output with intended links, then apply to a disposable target.
7. Migrate real dotfiles only after conflict handling and history behavior are
   verified. Keep the source C# application and real dotfiles unchanged during
   the rewrite itself.
