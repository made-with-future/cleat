# Implementation Plan: Refactor GetStrategyForCommand Dispatcher

Refactor the strategy routing logic to use a modular dispatcher pattern, ensuring high test coverage and extensibility.

## Phase 1: Baseline and Coverage
- [ ] Task: Increase test coverage for `GetStrategyForCommand` to >80%
    - [ ] Analyze current gaps in `strategy_test.go`
    - [ ] Write additional unit tests for edge cases (GCP, Terraform, missing services)
    - [ ] Verify coverage reaches >80% for `GetStrategyForCommand`
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Baseline and Coverage' (Protocol in workflow.md)

## Phase 2: Dispatcher Foundation
- [ ] Task: Define `CommandProvider` interface
    - [ ] Add `CommandProvider` interface to `internal/strategy/strategy.go`
- [ ] Task: Implement `RegistryProvider`
    - [ ] Write tests for the `RegistryProvider`
    - [ ] Implement `RegistryProvider` to handle strategies registered via `Register()`
- [ ] Task: Implement Dispatcher Loop
    - [ ] Refactor `GetStrategyForCommand` to use a slice of `CommandProviders`
    - [ ] Verify existing tests still pass with only the `RegistryProvider` integrated
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Dispatcher Foundation' (Protocol in workflow.md)

## Phase 3: Module Extraction
- [ ] Task: Extract NPM Strategy Logic
    - [ ] Write tests for `NpmProvider`
    - [ ] Move NPM routing logic to `NpmProvider` and register it
- [ ] Task: Extract Docker Strategy Logic
    - [ ] Write tests for `DockerProvider`
    - [ ] Move Docker routing logic to `DockerProvider` and register it
- [ ] Task: Extract Django Strategy Logic
    - [ ] Write tests for `DjangoProvider`
    - [ ] Move Django routing logic to `DjangoProvider` and register it
- [ ] Task: Extract GCP Strategy Logic
    - [ ] Write tests for `GcpProvider`
    - [ ] Move GCP routing logic to `GcpProvider` and register it
- [ ] Task: Extract Terraform Strategy Logic
    - [ ] Write tests for `TerraformProvider`
    - [ ] Move Terraform routing logic to `TerraformProvider` and register it
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Module Extraction' (Protocol in workflow.md)