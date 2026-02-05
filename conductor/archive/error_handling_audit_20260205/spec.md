# Specification: Comprehensive Error Context and Logging Audit

## Overview
Cleat currently has several instances where errors are either silently ignored (e.g., `_ = history.Load()`) or lack sufficient context for debugging. This track aims to perform a system-wide audit of the `internal/` packages to improve error handling by introducing a structured logging system and surfacing critical errors to the user via the TUI.

## Functional Requirements
- **System-wide Error Audit:** Search and identify all instances of ignored errors (`_ = `) and missing error context across all `internal/` packages.
- **Structured Logging System:**
    - Implement a logging utility, such as rs/zerolog, that writes to a persistent log file (e.g., `~/.cleat/cleat.log`).
    - Use **Structured JSON** format for log entries.
    - Log levels: `DEBUG`, `INFO`, `WARN`, `ERROR`.
- **Hybrid Error Handling:**
    - **Non-Fatal Errors/Warnings:** Log to the persistent file without interrupting the user.
    - **"Deal-Breaker" Errors:** Surface these to the user via the TUI (e.g., status bar or notifications) in addition to logging.
- **Improved Context:** Wrap errors with meaningful context using `fmt.Errorf("context: %w", err)` to aid in traceability.

## Architecture
- **Logging Package:** Create a new `internal/logger` package to handle the persistent JSON logging.
- **UI Integration:** Update the `internal/ui` model to support displaying critical error notifications.

## Acceptance Criteria
- No instances of `_ =` for functions that return meaningful errors in the `internal/` package.
- A `~/.cleat/cleat.log` file is created and populated with structured JSON logs when errors or warnings occur.
- Critical errors (like failing to load a required config) are clearly visible in the TUI.
- All tests pass, including new tests for the logging system.
