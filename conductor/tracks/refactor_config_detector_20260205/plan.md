# Implementation Plan: Refactor Configuration Logic

This plan outlines the refactoring of configuration and auto-detection logic.

## Phase 1: Foundation and Extraction
- [x] Task: Create \`internal/detector\` package structure and interface (c10d6f8)
    - [x] Write tests for the base detector interface
    - [x] Implement the base detector interface
- [x] Task: Extract Docker auto-detection logic (6601538)
    - [x] Write tests for Docker detector
    - [x] Move Docker detection logic from \`internal/config\` to \`internal/detector\`
- [x] Task: Extract Django and NPM auto-detection logic (a2dcc96)
    - [x] Write tests for Django and NPM detectors
    - [x] Move Django/NPM detection logic to \`internal/detector\`
- [ ] Task: Extract GCP and Terraform auto-detection logic
    - [ ] Write tests for GCP and Terraform detectors
    - [ ] Move GCP/Terraform detection logic to `internal/detector`
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Foundation and Extraction' (Protocol in workflow.md)

## Phase 2: Config Integration and Cleanup
- [ ] Task: Update `internal/config` to use `internal/detector`
    - [ ] Write tests for the integrated configuration loading
    - [ ] Refactor `internal/config` to call detectors
- [ ] Task: Remove redundant detection logic from `internal/config`
    - [ ] Clean up `internal/config` and ensure all tests pass
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Config Integration and Cleanup' (Protocol in workflow.md)