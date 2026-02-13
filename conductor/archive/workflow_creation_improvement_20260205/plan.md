# Implementation Plan: Enhanced Workflow Creation and Persistence

This plan implements the new workflow identity system, ensuring workflows have unique machine-readable IDs and human-readable names with robust "upsert" persistence, and polishes the TUI creation flow. It includes strict quality gates for code coverage.

## Phase 1: Data Model and Core Logic
Update the core data structures and implement the slugification and validation logic.

- [x] Task: Coverage Audit and Baseline
    - [x] Audit coverage for `internal/config/schema/schema.go` and `internal/history/workflows.go`.
    - [x] If any touched existing methods are below 80%, add tests to reach the 80% baseline.
- [x] Task: Update the `Workflow` data structure
    - [x] Modify `internal/config/schema/schema.go` to add the `ID` field to the `Workflow` struct.
    - [x] Ensure correct YAML tags for serialization.
- [x] Task: Implement Name Validation and Slugification (TDD)
    - [x] Write failing tests in `internal/history/workflows_test.go` for:
        - [x] Validating that names are at least one character long.
        - [x] Generating a lowercase kebab-case slug (Option B: spaces to hyphens, remove non-alphanumeric).
        - [x] Ensuring numeric names result in string IDs.
    - [x] Implement `slugify(name string) string` helper.
    - [x] Implement `validateWorkflowName(name string) error` helper.
    - [x] **Quality Gate:** Ensure new helpers have at least 50% code coverage.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Data Model and Core Logic' (Protocol in workflow.md)

## Phase 2: Persistence Layer (ID-based Upsert)
Refactor the persistence logic to use the new `ID` field as the primary key for updates.

- [x] Task: Implement ID-based Upsert in `modifyWorkflowFile` (TDD)
    - [x] Audit coverage for `modifyWorkflowFile`. If below 80%, bring it up to 80%.
    - [x] Write failing tests in `internal/history/workflows_test.go` that attempt to save a workflow with a different name but an ID that clashes with an existing one.
    - [x] Update `modifyWorkflowFile` in `internal/history/workflows.go` to match workflows by `ID` instead of `Name`.
- [x] Task: Update Save and Load procedures
    - [x] Update `SaveWorkflowToProject` and `SaveWorkflowToUser` to generate the `ID` from the `Name` before persistence.
    - [x] Update `LoadWorkflows` to handle backward compatibility (populate `ID` from `Name` if missing in the file).
    - [x] **Quality Gate:** Ensure all modified logic in `workflows.go` maintains or exceeds 80% coverage.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Persistence Layer (ID-based Upsert)' (Protocol in workflow.md)

## Phase 3: TUI Integration
Integrate the new validation and identity logic into the user interface.

- [x] Task: Update Workflow Creation Flow
    - [x] Audit coverage for TUI event handlers related to workflow creation. If below 80%, bring up to 80%.
    - [x] Refactor the TUI event handling for workflow creation to use `validateWorkflowName`.
    - [x] Ensure error feedback is displayed in the TUI if the name is invalid.
- [x] Task: Adjust Workflow Name Modal UI
    - [x] Modify the rendering logic for the workflow name input modal in `internal/ui/rendering.go` (or applicable UI file).
    - [x] Ensure the input field appears on its own line below the prompt message.
- [x] Task: Conductor - User Manual Verification 'Phase 3: TUI Integration' (Protocol in workflow.md)
