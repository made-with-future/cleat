# Specification: Restore Workflow Command Dispatching

## Overview
A regression was introduced during the strategy dispatcher refactor where workflows defined in `cleat.yaml` are no longer recognized. Commands prefixed with `workflow:` result in an `ERROR: unknown command`. This track will restore the `WorkflowProvider` to the dispatcher system to enable workflow execution.

## Functional Requirements
- **WorkflowProvider Implementation:** Create a new `WorkflowProvider` in `internal/strategy/strategy.go` (or a dedicated file).
- **Command Resolution:**
    - Detect commands starting with `workflow:`.
    - Look up the corresponding workflow name in the `Session.Config.Workflows` slice.
    - If found, return a strategy that executes the sequence of commands defined in the workflow.
- **Dynamic Input Support:** Ensure that the resulting strategy correctly handles requirement collection and input passing for the constituent tasks, maintaining the behavior prior to the refactor.
- **TUI Integration:** Ensure the TUI can successfully trigger these resolved strategies without error.

## Acceptance Criteria
- Running `cleat workflow:{name}` (or triggering it via the TUI) successfully resolves the strategy.
- The workflow executes its constituent commands in the defined sequence.
- If any command in the sequence fails, the workflow execution stops (Fail Fast).
- All tests in `internal/strategy/strategy_test.go` pass, including new test cases for workflow resolution.
