# Specification: Self-Hosting Cleat Orchestration

## Overview
This track aims to replace the project's internal use of `make` with Cleat itself. By "dogfooding" the tool, we ensure its configuration is robust and its features meet the needs of its own development lifecycle. The `Makefile` will be entirely replaced by a `cleat.yaml` configuration.

## Functional Requirements
- **Command Migration:** Map existing `make` targets to Cleat commands and workflows:
    - `build` -> `cleat build` (leveraging auto-detection or explicit configuration).
    - `test` -> `cleat go test` (or similar).
    - `fmt` -> `cleat go fmt`.
    - `vet` -> `cleat go vet`.
    - `coverage` -> `cleat workflow coverage` (running tests and generating/checking coverage).
    - `install` -> `cleat workflow install` (building and moving the binary to the correct local path).
- **Workflow Implementation:**
    - Create a `build-all` workflow that sequences the cross-compilation of Linux and Darwin binaries.
    - Create an `install` workflow that handles OS detection (Darwin vs. others) and copies the binary to either `/usr/local/bin` or `$HOME/.local/bin`.
- **Cleanup:**
    - Permanently delete the `Makefile` once the migration is verified.

## Non-Functional Requirements
- **Consistency:** The new Cleat commands should produce the same outputs and artifacts as the previous `make` targets.
- **Dogfooding:** Ensure the `cleat.yaml` reflects the latest features and best practices identified during development.

## Acceptance Criteria
- [ ] `cleat build` successfully compiles the binary.
- [ ] `cleat go test` and `cleat go fmt` function as expected.
- [ ] `cleat workflow build-all` produces all four cross-compiled binaries.
- [ ] `cleat workflow install` correctly installs the binary to the user's PATH based on their OS.
- [ ] The `Makefile` is removed from the repository.
- [ ] The project can be fully managed using only `cleat`.

## Out of Scope
- Migrating the `run` target (which simply executed the built binary).
- Migrating the `setup-hooks` target.
