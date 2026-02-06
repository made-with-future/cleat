# Implementation Plan: Workflow Engine Core Fixes

## Overview
This plan implements critical fixes to the workflow engine to ensure robust error handling, efficient execution, deterministic user interaction, and protection against infinite recursion. Focus is on automated verification.

## Phase 1: Workflow Loading & Error Handling
Address the silent swallowing of errors during workflow loading.

- [x] Task: Write failing tests for workflow loading errors [checkpoint: 1218751]
    - [x] Create a test case with a malformed workflow file (e.g., bad YAML syntax).
    - [x] Assert that `LoadWorkflows` returns an error or logs a warning.
- [x] Task: Implement error propagation in `LoadWorkflows` [checkpoint: 1219046]
    - [x] Modify `LoadWorkflows` in `internal/history/workflows.go` to capture and log/return errors from `yaml.Unmarshal`.
    - [x] Ensure the application doesn't crash, but informs the user.
- [x] Task: Verify fix with automated tests
    - [x] Run the test suite to confirm the loading error is correctly reported.

## Phase 2: Execution Efficiency
Eliminate the redundant double task resolution in `WorkflowStrategy`.

- [x] Task: Write failing tests for double resolution (Mock verification) [checkpoint: 1219459]
    - [x] Create a mock `ResolveCommandTasks` or instrument the strategy to count resolution calls.
    - [x] Assert that `ResolveTasks` is called exactly once per command in the workflow.
- [x] Task: Refactor `WorkflowStrategy.Execute` [checkpoint: 1219779]
    - [x] Modify `Execute` to use the result of the initial `ResolveTasks` call.
    - [x] Remove the second loop that calls `GetStrategyForCommand` and `ResolveTasks` again.
    - [x] Ensure that execution logic (running the tasks) iterates over the *resolved* tasks.
- [x] Task: Verify fix with automated tests
    - [x] Run the instrumented test to confirm single resolution pass.

## Phase 3: Deterministic UX
Ensure input prompts always appear in a stable order.

- [x] Task: Write failing tests for input prompt ordering [checkpoint: 1220384]
    - [x] Create a strategy with multiple inputs (e.g., "z-input", "a-input", "m-input").
    - [x] Mock the executor's `Prompt` method to record the order of calls.
    - [x] Run the test multiple times to confirm current non-determinism or fail if not sorted.
- [x] Task: Implement sorted input prompting [checkpoint: 1220599]
    - [x] Modify the input collection logic in `WorkflowStrategy` (and `BaseStrategy` if applicable/shared) to extract keys, sort them, and iterate the sorted slice for prompting.
- [x] Task: Verify fix with automated tests
    - [x] Assert that prompts always occur in alphabetical order (a, m, z).

## Phase 4: Recursion Safety
Implement cycle detection and depth limits for nested workflows.

- [x] Task: Write failing tests for recursion cycles [checkpoint: 1223600]
    - [x] Create a self-referencing workflow (A -> workflow:A).
    - [x] Create a mutual recursion loop (A -> workflow:B, B -> workflow:A).
    - [x] Assert that these fail with a specific "Cycle Detected" error (and do not stack overflow).
- [x] Task: Implement Cycle Detection & Depth Limit [checkpoint: 1223827]
    - [x] Update `WorkflowProvider` and `WorkflowStrategy` to accept a context/visited-map.
    - [x] Implement the "visited" check before resolving/executing a sub-workflow.
    - [x] Implement a hard depth counter.
- [x] Task: Verify fix with automated tests
    - [x] Run the recursion tests and confirm they pass with the correct error message.