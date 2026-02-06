# Specification: Workflow Code Refactoring and Cleanup

## Overview
This track focuses on improving the code quality, maintainability, and design integrity of the workflow implementation. It addresses technical debt related to code duplication in persistence logic and interface inconsistencies in the strategy pattern.

## Functional Requirements

### 1. Interface Integrity (WorkflowStrategy.Tasks)
*   **Issue:** `WorkflowStrategy.Tasks()` currently returns `nil`, effectively being a dead method, while the `Strategy` interface implies it should return tasks. Callers relying on it get nothing.
*   **Requirement:** The ambiguity must be resolved.
*   **Approach:**
    *   Since `WorkflowStrategy` tasks depend on the `Session` (resolved dynamically), `Tasks()` without arguments cannot be correct.
    *   **Action:** Add comprehensive GoDoc comments to `Strategy.Tasks()` explaining that dynamic strategies may return nil.
    *   **Action:** Modify `WorkflowStrategy.Tasks()` to return a specific "PlaceholderTask" that explains (if introspected) that tasks must be resolved via `ResolveTasks(session)`, OR simply ensure all internal callers use `ResolveTasks`.

### 2. Deduplicate Persistence Logic
*   **Issue:** `SaveWorkflowToProject`, `SaveWorkflowToUser`, and `DeleteWorkflow` share identical boilerplate for reading, unmarshaling, modifying, marshaling, and writing YAML files.
*   **Requirement:** Refactor this common logic into a shared helper function.
*   **Helper Signature (Example):** `modifyWorkflowFile(path string, modificationFunc func([]config.Workflow) []config.Workflow) error`
*   **Impact:** Reduces code duplication by ~40 lines and centralizes error handling for file I/O.

### 3. Deduplicate Project ID Hashing
*   **Issue:** `GetUserWorkflowFilePath` manually re-implements the SHA-256 hashing logic for project IDs.
*   **Requirement:** Replace the manual implementation with a call to the existing `config.GetProjectID()` (or move the logic to a shared utility if `config` is circular, though `config` should be accessible).

## Non-Functional Requirements
*   **Code Quality:** Resulting code should be cleaner (DRY) and easier to test.
*   **No Behavior Change:** These refactorings must NOT change the external behavior of the CLI. Existing tests must pass.

## Acceptance Criteria
1.  `WorkflowStrategy.Tasks()` is documented or refactored to be safe.
2.  `SaveWorkflowToProject`, `SaveWorkflowToUser`, and `DeleteWorkflow` use a shared helper function.
3.  `GetUserWorkflowFilePath` uses `config.GetProjectID()`.
4.  All existing workflow tests pass.
