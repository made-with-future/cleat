# Implementation Plan: Workflow Engine Core Fixes

## Overview
This plan implements critical fixes to the workflow engine to ensure robust error handling, efficient execution, deterministic user interaction, and protection against infinite recursion. Focus is on automated verification.

## Phase 1: Workflow Loading & Error Handling
Address the silent swallowing of errors during workflow loading.

- [ ] Task: Write failing tests for workflow loading errors
    - [ ] Create a test case with a malformed workflow file (e.g., bad YAML syntax).
    - [ ] Assert that `LoadWorkflows` returns an error or logs a warning.
- [ ] Task: Implement error propagation in `LoadWorkflows`
    - [ ] Modify `LoadWorkflows` in `internal/history/workflows.go` to capture and log/return errors from `yaml.Unmarshal`.
    - [ ] Ensure the application doesn't crash, but informs the user.
- [ ] Task: Verify fix with automated tests
    - [ ] Run the test suite to confirm the loading error is correctly reported.

## Phase 2: Execution Efficiency
Eliminate the redundant double task resolution in `WorkflowStrategy`.

- [ ] Task: Write failing tests for double resolution (Mock verification)
    - [ ] Create a mock `ResolveCommandTasks` or instrument the strategy to count resolution calls.
    - [ ] Assert that `ResolveTasks` is called exactly once per command in the workflow.
- [ ] Task: Refactor `WorkflowStrategy.Execute`
    - [ ] Modify `Execute` to use the result of the initial `ResolveTasks` call.
    - [ ] Remove the second loop that calls `GetStrategyForCommand` and `ResolveTasks` again.
    - [ ] Ensure that execution logic (running the tasks) iterates over the *resolved* tasks.
- [ ] Task: Verify fix with automated tests
    - [ ] Run the instrumented test to confirm single resolution pass.

## Phase 3: Deterministic UX
Ensure input prompts always appear in a stable order.

- [ ] Task: Write failing tests for input prompt ordering
    - [ ] Create a strategy with multiple inputs (e.g., "z-input", "a-input", "m-input").
    - [ ] Mock the executor's `Prompt` method to record the order of calls.
    - [ ] Run the test multiple times to confirm current non-determinism or fail if not sorted.
- [ ] Task: Implement sorted input prompting
    - [ ] Modify the input collection logic in `WorkflowStrategy` (and `BaseStrategy` if applicable/shared) to extract keys, sort them, and iterate the sorted slice for prompting.
- [ ] Task: Verify fix with automated tests
    - [ ] Assert that prompts always occur in alphabetical order (a, m, z).

## Phase 4: Recursion Safety
Implement cycle detection and depth limits for nested workflows.

- [ ] Task: Write failing tests for recursion cycles
    - [ ] Create a self-referencing workflow (A -> workflow:A).
    - [ ] Create a mutual recursion loop (A -> workflow:B, B -> workflow:A).
    - [ ] Assert that these fail with a specific "Cycle Detected" error (and do not stack overflow).
- [ ] Task: Implement Cycle Detection & Depth Limit
    - [ ] Update `WorkflowProvider` and `WorkflowStrategy` to accept a context/visited-map.
    - [ ] Implement the "visited" check before resolving/executing a sub-workflow.
    - [ ] Implement a hard depth counter.
- [ ] Task: Verify fix with automated tests
    - [ ] Run the recursion tests and confirm they pass with the correct error message.