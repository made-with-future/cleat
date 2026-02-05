# Specification: Refactor Transient State to Session Object

## Overview
Eliminate the use of package-level variables (`transientInputs`) and static setters (`SetTransientInputs`) in `internal/config`. These will be replaced by a dedicated `Session` object that encapsulates all dynamic state (inputs, configuration, and executor) and is passed explicitly through the execution call chain.

## Functional Requirements
- **Session Package:** Create a new package `internal/session` defining a `Session` struct.
- **Session Composition:** The `Session` struct will contain:
    - `Config *schema.Config`
    - `Inputs map[string]string`
    - `Exec executor.Executor`
- **Signature Refactoring:** Update the following interfaces and their implementations to accept `*session.Session`:
    - `internal/strategy/Strategy` methods (`Execute`, `ResolveTasks`, etc.)
    - `internal/task/Task` methods (`Run`, `Requirements`, `ShouldRun`)
- **Removal of Global State:** Completely remove `transientInputs` and `SetTransientInputs` from `internal/config/config.go`.
- **Initialization:** Update the CLI entry points (Cobra commands) to initialize a `Session` at the start of execution and pass it down.

## Architecture
- **Dependency Injection:** Moving to a pattern where all dependencies required for execution are explicitly provided via the `Session` object.
- **Decoupling:** `internal/config` becomes strictly about schema and parsing, while `internal/session` handles the runtime context.

## Acceptance Criteria
- **Zero Global State:** `grep` searching for `transientInputs` in `internal/config` returns no results.
- **Clean Signatures:** `Strategy.Execute` and `Task.Run` take exactly one `*session.Session` as their primary context argument.
- **Passing Tests:** The entire test suite (`go test ./...`) passes, with tests updated to use the `Session` object.
- **Maintainability:** Adding new execution-scoped data (e.g., global flags) only requires updating the `Session` struct, not all method signatures.