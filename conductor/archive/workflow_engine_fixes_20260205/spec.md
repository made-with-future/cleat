# Specification: Workflow Engine Core Fixes

## Overview
This track addresses critical stability and correctness issues in the workflow execution engine. It focuses on error visibility during loading, execution efficiency, deterministic behavior for user interactions, and robust prevention of infinite recursion in nested workflows.

## Functional Requirements

### 1. Robust Workflow Loading
*   **Requirement:** `LoadWorkflows` must not silently swallow errors.
*   **Behavior:** If a workflow file is malformed or cannot be read (permission errors, etc.), the system must log a warning or error containing the specific failure reason.
*   **User Impact:** Users will be informed why a workflow might be missing from the list.

### 2. Efficient Task Resolution
*   **Requirement:** `WorkflowStrategy.Execute` must resolve tasks exactly once per execution.
*   **Behavior:** The list of tasks resolved in the initial validation/requirements-gathering step must be passed to the execution phase, eliminating the redundant second resolution pass.
*   **Constraint:** Ensure that any state dependent on resolution (e.g., dynamic inputs) is correctly preserved.

### 3. Deterministic Input Prompting
*   **Requirement:** User prompts for workflow inputs must appear in a consistent, deterministic order.
*   **Behavior:** When iterating over the `map[string]task.InputRequirement`, the system must sort the keys (e.g., alphabetically) before prompting the user.
*   **Scope:** This applies to both `WorkflowStrategy` and `BaseStrategy` (or the shared prompting logic).

### 4. Recursive Workflow Cycle Detection
*   **Requirement:** The system must prevent infinite recursion caused by workflows referencing themselves or forming a loop.
*   **Mechanism 1 (Cycle Detection):** Maintain a "visited stack" of workflow names during resolution. If a workflow attempts to invoke a workflow already in the current stack, execution must halt immediately with a "Cycle Detected" error.
*   **Mechanism 2 (Depth Limit):** Enforce a hard limit on nesting depth (e.g., 100 levels) as a failsafe against complex recursion scenarios.
*   **Error Reporting:** The error message must clearly indicate the cycle path (e.g., "A -> B -> A") or that the depth limit was exceeded.

## Non-Functional Requirements
*   **Performance:** Eliminating double resolution should marginally improve start-up time for large workflows.
*   **Testability:** New logic (cycle detection, prompting order) must be covered by unit tests.

## Acceptance Criteria
1.  Malformed workflow files result in a visible error log.
2.  `WorkflowStrategy` calls `ResolveTasks` only once per execution flow.
3.  Running a workflow with multiple inputs always prompts in the same order (e.g., alphabetical by key).
4.  A self-referencing workflow (e.g., A calls A) fails with a specific "Cycle Detected" error.
5.  A mutual recursion loop (A calls B, B calls A) fails with a "Cycle Detected" error.
6.  Existing valid workflows continue to function unchanged.

## Out of Scope
*   Refactoring persistence logic (moved to a separate track).
*   Addressing `Tasks()` interface method (moved to a separate track).
