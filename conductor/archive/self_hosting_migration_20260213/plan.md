# Implementation Plan: Self-Hosting Cleat Orchestration

This plan outlines the migration from `make` to `cleat` for the project's internal development tasks.

## Phase 1: Cleat Configuration and Basic Commands [checkpoint: 916571f]
Configure `cleat.yaml` to handle the core Go build and linting tasks.

- [x] Task: Initialize and configure `cleat.yaml` [7514f7f]
- [x] Task: Verify basic commands [7514f7f]
- [x] Task: Conductor - User Manual Verification 'Phase 1: Cleat Configuration and Basic Commands' (Protocol in workflow.md)

## Phase 2: Complex Workflows Implementation
Implement the `build-all`, `coverage`, and `install` workflows in `cleat.yaml`.

- [x] Task: Implement `build-all` workflow [a8c30a3]
- [x] Task: Implement `coverage` workflow [a8c30a3]
- [x] Task: Implement `install` workflow [a8c30a3]
- [x] Task: Conductor - User Manual Verification 'Phase 2: Complex Workflows Implementation' (Protocol in workflow.md)

## Phase 3: Final Transition and Cleanup
Decommission the `Makefile` and finalize the migration.

- [x] Task: Verify full orchestration [a8c30a3]
- [x] Task: Remove `Makefile` [a8c30a3]
- [x] Task: Update documentation [a8c30a3]
    - [x] Ensure `README.md` and `CONTRIBUTING.md` reflect the switch to `cleat` for development tasks.
- [x] Task: Fix CI workflow coverage dependency [da57bd1]
- [x] Task: Conductor - User Manual Verification 'Phase 3: Final Transition and Cleanup' (Protocol in workflow.md)

## Phase 4: Finalization and Submission
Clean up history and submit the changes.

- [x] Task: Squash commits and push [d0823c4]
- [x] Task: Push the feature branch to GitHub and create a PR. [d0823c4]
- [x] Task: Conductor - User Manual Verification 'Phase 4: Finalization and Submission' (Protocol in workflow.md)
