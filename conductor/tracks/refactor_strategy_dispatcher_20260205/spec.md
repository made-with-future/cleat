# Specification: Refactor GetStrategyForCommand Dispatcher

## Overview
Refactor the large `GetStrategyForCommand` function in `internal/strategy/strategy.go` from a monolithic `if-else` chain into a modular, extensible dispatcher system. This involves introducing a `CommandProvider` interface and a centralized registry to manage command-to-strategy routing.

## Functional Requirements
- **CommandProvider Interface:** Define a standard interface with `CanHandle(command string) bool` and `GetStrategy(command string, cfg *schema.Config) Strategy`.
- **Modular Implementations:** Extract the current routing logic into specific provider implementations:
    - `NpmProvider` (Install and Script execution)
    - `DockerProvider` (Service-specific commands)
    - `DjangoProvider` (Service-specific commands)
    - `GcpProvider` (App Engine commands)
    - `TerraformProvider`
    - `RegistryProvider` (For basic strategies registered via `Register`)
- **Centralized Registry:** Implement a central loop in `GetStrategyForCommand` that iterates through a prioritized list of `CommandProviders`.
- **Backward Compatibility:** All existing command strings and their corresponding strategy resolution must remain identical to the current behavior.

## Architecture
- The `strategy` package will hold the `CommandProvider` interface.
- Specific providers will be implemented (likely as private structs) within `strategy.go` or split into small files if they grow too large.
- `GetStrategyForCommand` will use a slice of these providers to resolve commands.

## Acceptance Criteria
- **Initial Coverage:** Increase test coverage of `GetStrategyForCommand` to >80% before refactoring.
- **Refactored Logic:** `GetStrategyForCommand` contains no command-specific parsing logic, only the dispatcher loop.
- **Verification:** All tests in `internal/strategy/strategy_test.go` pass with the refactored code.
- **New Tests:** A new test suite verifies that the dispatcher correctly delegates to providers in the intended order.