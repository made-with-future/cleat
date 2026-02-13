# Implementation Plan: Switch to MIT License

This plan outlines the steps to transition the project from GPLv3 to the MIT License.

## Phase 1: Root License Update [checkpoint: 1c53948]
Update the primary license file in the project root.

- [x] Task: Replace `LICENSE` file content [970f5f4]
    - [ ] Remove existing GPLv3 text.
    - [ ] Add standard MIT License text with `Copyright (c) 2026 Josh Turmel`.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Root License Update' (Protocol in workflow.md)

## Phase 2: Project Audit and Cleanup [checkpoint: 0198e44]
Ensure no GPL references remain in the active codebase.

- [x] Task: Audit codebase for GPL references
    - [ ] Run a project-wide grep for "GPL", "GNU", and "General Public License".
    - [ ] Remove or update any lingering references.
- [x] Task: Verify functionality
    - [ ] Run `go test ./...` to ensure the project still builds and tests pass.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Project Audit and Cleanup' (Protocol in workflow.md)

## Phase 3: Finalization and Submission
Ensure quality standards are met and submit the changes.

- [x] Task: Verify code coverage (76.7%)
- [x] Task: Prepare and push Pull Request [81e9d7e]
    - [ ] Squash all track commits into a single clean commit.
    - [ ] Push the feature branch to GitHub and create a PR.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Finalization and Submission' (Protocol in workflow.md)
