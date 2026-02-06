# Implementation Plan: Enhanced Workflow Creation and Persistence

This plan implements the new workflow identity system, ensuring workflows have unique machine-readable IDs and human-readable names with robust "upsert" persistence, and polishes the TUI creation flow. It includes strict quality gates for code coverage.

## Phase 1: Data Model and Core Logic
Update the core data structures and implement the slugification and validation logic.

- [ ] Task: Coverage Audit and Baseline
    - [ ] Audit coverage for `internal/config/schema/schema.go` and `internal/history/workflows.go`.
    - [ ] If any touched existing methods are below 80%, add tests to reach the 80% baseline.
- [ ] Task: Update the `Workflow` data structure
    - [ ] Modify `internal/config/schema/schema.go` to add the `ID` field to the `Workflow` struct.
    - [ ] Ensure correct YAML tags for serialization.
- [ ] Task: Implement Name Validation and Slugification (TDD)
    - [ ] Write failing tests in `internal/history/workflows_test.go` for:
        - [ ] Validating that names are at least one character long.
        - [ ] Generating a lowercase kebab-case slug (Option B: spaces to hyphens, remove non-alphanumeric).
        - [ ] Ensuring numeric names result in string IDs.
    - [ ] Implement `slugify(name string) string` helper.
    - [ ] Implement `validateWorkflowName(name string) error` helper.
    - [ ] **Quality Gate:** Ensure new helpers have at least 50% code coverage.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Data Model and Core Logic' (Protocol in workflow.md)

## Phase 2: Persistence Layer (ID-based Upsert)
Refactor the persistence logic to use the new `ID` field as the primary key for updates.

- [ ] Task: Implement ID-based Upsert in `modifyWorkflowFile` (TDD)
    - [ ] Audit coverage for `modifyWorkflowFile`. If below 80%, bring it up to 80%.
    - [ ] Write failing tests in `internal/history/workflows_test.go` that attempt to save a workflow with a different name but an ID that clashes with an existing one.
    - [ ] Update `modifyWorkflowFile` in `internal/history/workflows.go` to match workflows by `ID` instead of `Name`.
- [ ] Task: Update Save and Load procedures
    - [ ] Update `SaveWorkflowToProject` and `SaveWorkflowToUser` to generate the `ID` from the `Name` before persistence.
    - [ ] Update `LoadWorkflows` to handle backward compatibility (populate `ID` from `Name` if missing in the file).
    - [ ] **Quality Gate:** Ensure all modified logic in `workflows.go` maintains or exceeds 80% coverage.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Persistence Layer (ID-based Upsert)' (Protocol in workflow.md)

## Phase 3: TUI Integration
Integrate the new validation and identity logic into the user interface.

- [ ] Task: Update Workflow Creation Flow
    - [ ] Audit coverage for TUI event handlers related to workflow creation. If below 80%, bring up to 80%.
    - [ ] Refactor the TUI event handling for workflow creation to use `validateWorkflowName`.
    - [ ] Ensure error feedback is displayed in the TUI if the name is invalid.
- [ ] Task: Adjust Workflow Name Modal UI
    - [ ] Modify the rendering logic for the workflow name input modal in `internal/ui/rendering.go` (or applicable UI file).
    - [ ] Ensure the input field appears on its own line below the prompt message.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: TUI Integration' (Protocol in workflow.md)
