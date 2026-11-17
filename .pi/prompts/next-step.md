---
description: Complete the next ready Alma migration step, test it, and commit it
---
Continue this repository's Alma Go migration. Complete at most one migration
step in this session; do not start subsequent steps automatically.

1. Establish context from repository files, not prior conversation. Read
   applicable `AGENTS.md` instructions, `MIGRATION.md`, and all of `PLAN.md`,
   including dependencies, acceptance checks, and handoff/blocker notes.
2. Inspect version-control status and relevant existing changes. Use `jj` if
   supported, otherwise Git. Preserve unrelated work; never reset, discard,
   or commit it as part of this task.
3. Prefer resuming a documented partial step if its dependencies are satisfied.
   Otherwise select the first unchecked step in plan order whose dependencies
   are complete. Verify readiness against the actual code and recorded decisions,
   not checkboxes alone. Announce the selected ID, scope, and verification plan.
4. If that step requires unresolved user decisions (especially S01), ask focused
   questions and record only approved choices. Do not invent approvals, silently
   change legacy behavior, or skip a policy gate. If blocked, record the blocker
   and next action without marking the step complete, then stop. If all steps
   are complete, report that and stop without inventing more work.
5. Implement the selected step and focused tests within its stated scope. Follow
   the YAML-only contract. Use disposable fixtures, controlled environments,
   and harmless/fake shell commands. Keep the original C# repository and real
   dotfiles read-only. Do not perform a real rollout, destructive unlink, or
   unapproved history import. If the work will not fit one session, split the
   remaining scope into explicit plan tasks and leave unfinished work unchecked.
6. Run focused tests and the full available suite; once the Go module exists,
   run `go test ./...`. Run any other required acceptance checks. For a purely
   documentation/decision step with no executable tests, run applicable document
   checks and explicitly state that no executable tests exist. Never present
   skipped, failed, or unrun required tests as passing.
7. On failure or incomplete work, leave the step unchecked, update `PLAN.md`
   with files changed, actual test results, blockers, and the precise next action,
   and stop without making a completion commit.
8. Only after the step's acceptance checks and required tests pass, mark its
   checkbox complete and update handoff notes. Keep `MIGRATION.md` consistent
   with approved decisions. Review the final diff and commit only the completed
   step's changes plus its progress/documentation updates. Do not include
   unrelated preexisting changes. If ownership of a change is ambiguous, ask
   rather than staging it. Follow applicable commit-message/trailer rules.
9. Finish with the completed or blocked step ID, concise change summary, actual
   checks/results, commit identifier if created, and next ready step/blocker.
   Do not begin that next step in this session.
