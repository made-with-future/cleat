# Specification: Project-wide Test Coverage Audit and Improvement

## Overview
This track focuses on increasing the robustness and reliability of Cleat by significantly improving test coverage. The primary goal is to reach an overall project coverage of at least 75%, with a consistent bar of 75% for individual modules and a minimum of 50% for every function. This will involve refactoring complex, untestable logic into smaller, testable units and implementing comprehensive mocks for external side effects.

## Functional Requirements
- **Coverage Targets:**
    - Overall Project: >= 75%
    - Individual Modules: >= 75%
    - Individual Functions: >= 50%
- **Refactoring for Testability:** Identify and refactor complex methods (especially in `internal/ui` and `internal/cmd`) to separate logic from side effects.
- **Priority Areas:**
    - **TUI Event Logic (`internal/ui`):** Comprehensive testing of keyboard interactions and state machine transitions.
    - **Command Routing (`internal/cmd`):** Verification of argument parsing and strategy dispatching.
    - **Auto-detection (`internal/detector`):** Hardening of stack and environment discovery edge cases.

## Non-Functional Requirements
- **Maintainability:** Refactored code must adhere to existing project patterns and improve overall code quality.
- **Performance:** Test suite execution time should remain within reasonable limits.

## Acceptance Criteria
- `go test -cover ./...` shows >= 75% overall coverage.
- No individual module in `internal/` has less than 75% coverage.
- All newly created or refactored functions have >= 50% coverage.
- All existing tests pass.

## Out of Scope
- Integration tests requiring real cloud credentials or live Docker environments (mocks will be used).
- Full end-to-end TUI automation (unit testing of the model and view logic is preferred).
