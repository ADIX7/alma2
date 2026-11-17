# Migration coverage audit (S25c)

Source of truth: [`MIGRATION.md`](../MIGRATION.md) §§7–8. Tests use disposable
fixtures; the C# source and real dotfiles are read-only. `internal/e2e` builds
`cmd/alma` and runs it with fixture-only CLI environment variables.

## §7 functional scenarios

| Scenario | Evidence |
| --- | --- |
| No configured repositories | `cli.TestListRepositoriesAndModules`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |
| Direct-child/hidden modules; root/deeper ignored | `operation.TestDiscoverModules`, `cli.TestListRepositoriesAndModules`, `e2e.TestYAMLToDryRunLinkHistoryAndRepeat` |
| Unavailable module (`ls`, `info`, `readme`, refusal) | `cli.TestListRepositoriesAndModules`, `cli.TestModuleInfoNoConfigUnavailableAndErrors`, `cli.TestReadmePriorityAndLiteralOutput`, `cli.TestLinkDryRunWorkingDirectoryAndUnavailableModule`, `cli.TestConfiguredCommandMissingAndUnavailable` |
| Working-directory override and shell cwd | `startup.TestWorkDirectoryPrecedence`, `cli.TestLinkDryRunWorkingDirectoryAndUnavailableModule`, `cli.TestInstallOnePositionUsesWorkdirModule` |
| Module without config | `cli.TestLinkUnconfiguredAndRealWrites`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |
| Default + Linux + Fedora settings | `config.TestSelectSpecificityAndOverrides`, `platform.TestLinuxReleaseFile` |
| Nonmatching sections unavailable | `config.TestSelectApplicability`, `cli.TestLinkDryRunWorkingDirectoryAndUnavailableModule` |
| Unmapped root file | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat` |
| Explicit renamed file | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkplan.TestMappedFilesDirectoriesAndTargets` |
| Nested file mapping; repeated basename isolation | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkplan.TestMappedFilesDirectoriesAndTargets` |
| Mapped directory as one symlink | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkplan.TestMappedFilesDirectoriesAndTargets` |
| Distinct sources, same destination | `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly`, `linkplan.TestPlanCollisionOnlyAfterExclusions` |
| Source symlinks skipped | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkplan.TestMappedSymlinkSourcesSkipped`, `linkplan.TestAutomaticSkipsSymlinks` |
| Nested unmapped file/parent creation | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkapply.TestApplyCreatesLinksAndPreservesConflicts` |
| Excluded subtree/whole-directory caveat | `linkplan.TestPlanMergedExclusionsAndMappedDirectoryCaveat`, `e2e.TestYAMLToDryRunLinkHistoryAndRepeat` |
| Configured README default/override | `linkplan.TestPlanConfiguredAndUnconfiguredPayloads`, `linkplan.TestPlanMergedExclusionsAndMappedDirectoryCaveat` |
| Existing destination and history skip | `cli.TestLinkUnconfiguredAndRealWrites`, `cli.TestLinkDryRunProposalsAndConflicts` |
| Fully read-only dry run | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `cli.TestLinkDryRunFailuresAreReadOnly` |
| Partial filesystem failure and persisted successes | `e2e.TestSubprocessPartialFailurePersistsOnlySuccess`, `linkapply.TestApplyContinuesOnFailureAndRejectsChangedSource` |
| Install/configure absent | `cli.TestConfiguredCommandMissingAndUnavailable` |
| Install/configure present | `cli.TestInstallConfigureExplicitHarmlessCommands`, `e2e.TestSubprocessCommandRestrictionsAndUnlink` |
| JSON-only config ignored | `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly`, `config.TestLoadModuleAndPresence` |
| Both formats; neither config as payload | `e2e.TestYAMLToDryRunLinkHistoryAndRepeat`, `linkplan.TestPlanConfiguredAndUnconfiguredPayloads` |
| Quoted home target vs bare YAML null | `e2e.TestSubprocessQuotedHomeTarget`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |
| Standard README priority | `cli.TestReadmePriorityAndLiteralOutput`, `e2e.TestYAMLToDryRunLinkHistoryAndRepeat` |
| Unsupported nondeleting `unlink` | `e2e.TestSubprocessCommandRestrictionsAndUnlink`, `cli.TestDiagnosticsErrorsAndUnlinkNeverDeletes` |

## §8 approved policies

| Policy | Evidence |
| --- | --- |
| 1: module-relative mappings, later source wins, symlink safety | `linkplan.TestMappedFilesDirectoriesAndTargets`, `linkplan.TestMappedLastSourceWinsAndYAMLMerge`, `e2e.TestYAMLToDryRunLinkHistoryAndRepeat` |
| 2: specificity/availability/exclusions/README | `config.TestSelectSpecificityAndOverrides`, `config.TestSelectEqualSpecificityAndExplicitFalse`, `linkplan.TestPlanMergedExclusionsAndMappedDirectoryCaveat`, `cli.TestConfiguredCommandMissingAndUnavailable` |
| 3: dry-run/read-only, broken links, continue failures | `cli.TestLinkDryRunProposalsAndConflicts`, `e2e.TestSubprocessPartialFailurePersistsOnlySuccess`, `linkapply.TestApplyPermissionFailureWhenEnforced` |
| 4: YAML-only payload, tokens, separators, unknown repositories | `config.TestInvalidModuleYAML`, `config.TestInvalidRepositoryYAML`, `pathresolve.TestResolveErrorsAndWindowsAppData`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |
| 5: validation/exit behavior | `cli.TestParse`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly`, `linkapply.TestApplyContinuesOnFailureAndRejectsChangedSource` |
| 6: shell restrictions and README | `shell.TestShellSelectionAndValidation`, `cli.TestConfiguredCommandRejectsMultilineAndPropagatesExit`, `cli.TestReadmePriorityAndLiteralOutput` |
| 7: `ls`/`info` display fixes | `operation.TestDiscoverModules`, `cli.TestRepositoryInfoSortedCountsAndCapabilities`, `cli.TestModuleInfoResolvedTargetsAndExistence` |
| 8: old history ignored, unlink deferred | `history.TestSavePairsDedupAndOldFilesUntouched`, `e2e.TestSubprocessPartialFailurePersistsOnlySuccess`, `cli.TestDiagnosticsErrorsAndUnlinkNeverDeletes` |
| 9: flat direct-child module boundary | `operation.TestModuleBoundaryAndSymlinks`, `cli.TestListRepositoriesAndModules`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |
| 10: reject final-plan collisions before writes | `linkplan.TestPlanCollisionOnlyAfterExclusions`, `e2e.TestSubprocessRejectedConfigurationAndPlanAreReadOnly` |

## Platform limits

Linux `sh` and symlink tests run natively here. Windows `pwsh`, roaming
`%APPDATA%`, symlink permissions, and history replacement are not yet exercised
natively; Windows identifiers and shell selection have injectable unit tests.
Native Windows verification belongs to S26. POSIX permission tests skip when
run as root or when filesystem permissions cannot be enforced. No real-dotfile
rollout or legacy history import has been performed.
