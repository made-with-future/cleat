# Implementation Plan: Self-Hosting Cleat Orchestration

This plan outlines the migration from `make` to `cleat` for the project's internal development tasks.

## Phase 1: Cleat Configuration and Basic Commands
Configure `cleat.yaml` to handle the core Go build and linting tasks.

- [ ] Task: Initialize and configure `cleat.yaml`
    - [ ] Create or update `cleat.yaml` in the project root.
    - [ ] Define the `go` module and ensure auto-detection works for basic `build`, `test`, `fmt`, and `vet`.
- [ ] Task: Verify basic commands
    - [ ] Run `cleat build` and verify binary creation.
    - [ ] Run `cleat go test` and `cleat go fmt`.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Cleat Configuration and Basic Commands' (Protocol in workflow.md)

## Phase 2: Complex Workflows Implementation
Implement the `build-all`, `coverage`, and `install` workflows in `cleat.yaml`.

- [ ] Task: Implement `build-all` workflow
    - [ ] Define shell tasks for cross-compiling to Linux/AMD64, Linux/ARM64, Darwin/AMD64, and Darwin/ARM64.
    - [ ] Sequence them into a `build-all` workflow.
- [ ] Task: Implement `coverage` workflow
    - [ ] Define a workflow that runs tests with coverage, generates the report, and checks against the threshold.
- [ ] Task: Implement `install` workflow
    - [ ] Create a shell script or multi-step workflow that detects the OS and installs the binary to the correct directory.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Complex Workflows Implementation' (Protocol in workflow.md)

## Phase 3: Final Transition and Cleanup
Decommission the `Makefile` and finalize the migration.

- [ ] Task: Verify full orchestration
    - [ ] Run all migrated workflows and commands to ensure parity with `make`.
- [ ] Task: Remove `Makefile`
    - [ ] Delete the `Makefile` from the repository.
- [ ] Task: Update documentation
    - [ ] Ensure `README.md` and `CONTRIBUTING.md` reflect the switch to `cleat` for development tasks.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Final Transition and Cleanup' (Protocol in workflow.md)

## Phase 4: Finalization and Submission
Clean up history and submit the changes.

- [ ] Task: Squash commits and push
    - [ ] Squash all track commits into a single clean commit.
    - [ ] Push the feature branch to GitHub and create a PR.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Finalization and Submission' (Protocol in workflow.md)
