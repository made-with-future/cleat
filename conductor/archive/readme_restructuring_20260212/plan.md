# Implementation Plan: README and Documentation Restructuring

This plan outlines the steps to restructure the project documentation, simplifying the README and creating a dedicated configuration guide.

## Phase 1: Documentation Scaffolding [checkpoint: 98480bc]
Create the structure for the new documentation page.

- [x] Task: Create `docs` directory and `configuration.md` [65d90a0]
    - [ ] Create the `docs/` directory in the project root.
    - [ ] Initialize `docs/configuration.md` with a proper title and introductory paragraph.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Documentation Scaffolding' (Protocol in workflow.md)

## Phase 2: Content Migration [checkpoint: 451f119]
Move detailed technical information from the README to the configuration page.

- [x] Task: Migrate Configuration Reference [46f02fd]
- [x] Task: Refine `docs/configuration.md` [46f02fd]
    - [ ] Ensure all headers are consistent and the document is easy to navigate.
    - [ ] Add a "Back to README" link at the top or bottom.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Content Migration' (Protocol in workflow.md)

## Phase 3: README Refactoring [checkpoint: 67a8140]
Restructure and simplify the main README file.

- [x] Task: Reposition Installation section [46f02fd]
- [x] Task: Simplify Configuration section in README [1294879]
    - [ ] Remove the tables and complex examples.
    - [ ] Replace them with a high-level summary of `cleat.yaml`.
    - [ ] Add a prominent link to `docs/configuration.md` for "Full Configuration Reference".
- [ ] Task: Conductor - User Manual Verification 'Phase 3: README Refactoring' (Protocol in workflow.md)

## Phase 4: Final Validation [checkpoint: fbcae94]
Ensure all links and formatting are correct.

- [x] Task: Verify links and formatting [1294879]
    - [ ] Check that the new link to `docs/configuration.md` works in GitHub preview.
    - [ ] Ensure no broken relative links were introduced during migration.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Final Validation' (Protocol in workflow.md)

## Phase 5: Finalization and Submission [checkpoint: e5383aa]
Clean up history and submit the changes.

- [x] Task: Squash commits and push [b8ba77d]
- [x] Task: Create Pull Request [b8ba77d]
    - [ ] Use `gh pr create` to submit the restructuring for review.
- [ ] Task: Conductor - User Manual Verification 'Phase 5: Finalization and Submission' (Protocol in workflow.md)
