# Specification: Enhanced Workflow Creation and Persistence

## Overview
This feature enhances the workflow creation process by introducing a unique, machine-readable `ID` for each workflow while preserving a human-readable `Name`. It also implements "upsert" (update or insert) logic to allow users to easily overwrite existing workflows by name.

## Functional Requirements

### 1. Data Structure Update
- The `config.Workflow` structure (and the corresponding YAML schema) shall be updated to include an `ID` field.
- The `ID` field will store the machine-readable slug.
- The `Name` field will continue to store the human-readable string provided by the user.

### 2. Workflow Identity and Validation
- **Name Validation:** The human-readable `Name` must be at least one character long. Empty strings are not permitted.
- **ID Generation (Slugification):**
    - The `ID` is auto-generated from the `Name`.
    - **Logic:**
        1. Convert to all lowercase.
        2. Replace all spaces with hyphens (`-`).
        3. Remove all characters that are not alphanumeric (`a-z`, `0-9`) or hyphens.
- **Type Safety:** The `ID` must always be treated and stored as a string, even if the resulting slug consists only of numbers.

### 3. Persistence and Upsert Logic
- When saving a workflow (either to the project-local `cleat.workflows.yaml` or the user-level configuration), the system shall use the generated `ID` as the unique key.
- **Overwrite Behavior:** If a workflow with the same `ID` already exists in the target storage location, it shall be replaced with the new definition silently (without prompting for confirmation).
- If no matching `ID` is found, the new workflow is appended to the list.

### 4. TUI Polish
- In the TUI modal that prompts for a workflow name, the input field shall be displayed on a new line below the prompt message.

## Non-Functional Requirements
- **YAML Compatibility:** The `ID` field must be serialized to and deserialized from YAML.
- **Backward Compatibility:** Existing workflows that lack an `ID` should be handled gracefully (e.g., by generating an `ID` upon first load or treat `name` as ID source).
- **Code Coverage:** 
    - Existing methods touched by this track must reach at least 80% coverage.
    - New methods introduced by this track must reach at least 50% coverage.

## Acceptance Criteria
- [ ] Users can create a workflow with a human-readable name.
- [ ] The system correctly generates a lowercase kebab-case slug for the `ID`.
- [ ] Saving a workflow with a name that slugifies to an existing `ID` overwrites the old workflow.
- [ ] Workflow names are validated to ensure they are not empty.
- [ ] Workflow IDs are always stored as strings in the YAML files.
- [ ] The TUI name input is on its own line.

## Out of Scope
- Manual editing of the `ID` field through the UI (it is always auto-generated from the Name).
- Global uniqueness across different storage levels (e.g., a project workflow can still have the same ID as a user workflow; the upsert logic applies per-file).
