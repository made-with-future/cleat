# Implementation Plan: Comprehensive Error Context and Logging Audit

Improve error handling across Cleat by introducing structured JSON logging (using `rs/zerolog`) and surfacing critical errors in the TUI.

## Phase 1: Logging Foundation [checkpoint: 1d8e004]
- [x] Task: Create \`internal/logger\` package (fd578fc)
    - [x] Write tests for the JSON logger (verify output format and file writing)
    - [x] Implement \`internal/logger\` using \`rs/zerolog\` for structured JSON logging
    - [x] Support log levels: DEBUG, INFO, WARN, ERROR
    - [x] Ensure logs are written to \`~/.cleat/cleat.log\` (handle directory creation)
- [x] Task: Conductor - User Manual Verification 'Phase 1: Logging Foundation' (Protocol in workflow.md) (1d8e004)

## Phase 2: UI Error Notification Support
- [ ] Task: Update `internal/ui` to support error notifications
    - [ ] Write tests for the UI error display logic
    - [ ] Add an `Error` field or a notification system to the UI model
    - [ ] Implement a visually distinct error display in the TUI (e.g., a red status bar)
- [ ] Task: Conductor - User Manual Verification 'Phase 2: UI Error Notification Support' (Protocol in workflow.md)

## Phase 3: System-wide Audit and Refactoring
- [ ] Task: Audit and Refactor `internal/history` and `internal/config`
    - [ ] Identify ignored errors and missing context in these packages
    - [ ] Add tests for error cases that were previously ignored
    - [ ] Wrap errors with context and integrate the new logger
- [ ] Task: Audit and Refactor remaining `internal/` packages
    - [ ] Scan `executor`, `strategy`, `task`, and `cmd` for ignored errors
    - [ ] Update logic to either log (non-fatal) or return/notify (fatal) errors
- [ ] Task: Final System Verification
    - [ ] Run the entire test suite
    - [ ] Manually verify that errors are correctly logged to the file and surfaced in the TUI
- [ ] Task: Conductor - User Manual Verification 'Phase 3: System-wide Audit and Refactoring' (Protocol in workflow.md)
