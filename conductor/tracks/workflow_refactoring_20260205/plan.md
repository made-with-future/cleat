# Implementation Plan: Workflow Code Refactoring and Cleanup

## Overview
This plan addresses technical debt in the workflow implementation, focusing on deduplication and interface clarity.

## Phase 1: Persistence Refactoring
Deduplicate the file I/O logic for workflow persistence.

- [ ] Task: Create persistence helper
    - [ ] Create a private function `modifyWorkflowFile(path string, op func([]config.Workflow) ([]config.Workflow, error)) error` in `internal/history/workflows.go`.
    - [ ] Implement the read/unmarshal/modify/marshal/write cycle within this helper.
- [ ] Task: Refactor `SaveWorkflowToProject`
    - [ ] Update to use `modifyWorkflowFile`.
- [ ] Task: Refactor `SaveWorkflowToUser`
    - [ ] Update to use `modifyWorkflowFile`.
- [ ] Task: Refactor `DeleteWorkflow`
    - [ ] Update to use `modifyWorkflowFile`.
- [ ] Task: Verify refactoring
    - [ ] Run existing history/workflow tests to ensure no regression.

## Phase 2: Project ID Hashing
Remove duplicated hashing logic.

- [ ] Task: Refactor `GetUserWorkflowFilePath`
    - [ ] Import `github.com/madewithfuture/cleat/internal/config`.
    - [ ] Replace manual SHA-256 logic with `config.GetProjectID()` (ensure `GetProjectID` is exported and accessible).
- [ ] Task: Verify refactoring
    - [ ] Run tests dependent on user workflow paths.

## Phase 3: Interface Integrity
Clarify the `Strategy.Tasks()` contract.

- [ ] Task: Update `Strategy` interface documentation
    - [ ] Add comments to `Strategy.Tasks()` in `internal/strategy/strategy.go`.
- [ ] Task: Update `WorkflowStrategy.Tasks()`
    - [ ] Add comments or a placeholder return if deemed necessary to avoid nil-pointer exceptions in careless callers.
- [ ] Task: Verify
    - [ ] Ensure the project builds and lints.
