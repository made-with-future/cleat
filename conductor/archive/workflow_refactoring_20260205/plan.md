# Implementation Plan: Workflow Code Refactoring and Cleanup

## Overview
This plan addresses technical debt in the workflow implementation, focusing on deduplication and interface clarity.

## Phase 1: Persistence Refactoring
Deduplicate the file I/O logic for workflow persistence.

- [x] Task: Create persistence helper [checkpoint: 1227443]
    - [x] Create a private function `modifyWorkflowFile(path string, op func([]config.Workflow) ([]config.Workflow, bool, error)) error` in `internal/history/workflows.go`.
    - [x] Implement the read/unmarshal/modify/marshal/write cycle within this helper.
- [x] Task: Refactor `SaveWorkflowToProject`
    - [x] Update to use `modifyWorkflowFile`.
- [x] Task: Refactor `SaveWorkflowToUser`
    - [x] Update to use `modifyWorkflowFile`.
- [x] Task: Refactor `DeleteWorkflow`
    - [x] Update to use `modifyWorkflowFile`.
- [x] Task: Verify refactoring
    - [x] Run existing history/workflow tests to ensure no regression.

## Phase 2: Project ID Hashing
Remove duplicated hashing logic.

- [x] Task: Refactor `GetUserWorkflowFilePath` [checkpoint: 1227689]
    - [x] Import `github.com/madewithfuture/cleat/internal/config`.
    - [x] Replace manual SHA-256 logic with `config.GetProjectID()` (ensure `GetProjectID` is exported and accessible).
- [x] Task: Verify refactoring
    - [x] Run tests dependent on user workflow paths.

## Phase 3: Interface Integrity
Clarify the `Strategy.Tasks()` contract.

- [x] Task: Update `Strategy` interface documentation [checkpoint: 1227955]
    - [x] Add comments to `Strategy.Tasks()` in `internal/strategy/strategy.go`.
- [x] Task: Update `WorkflowStrategy.Tasks()`
    - [x] Add comments or a placeholder return if deemed necessary to avoid nil-pointer exceptions in careless callers.
- [x] Task: Verify
    - [x] Ensure the project builds and lints.
