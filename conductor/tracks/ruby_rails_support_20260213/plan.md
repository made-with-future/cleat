# Implementation Plan: Ruby and Ruby on Rails Support

This plan outlines the steps to add auto-detection and standardized command execution for Ruby and Ruby on Rails projects.

## Phase 1: Configuration and Detection
Update the configuration schema and implement auto-detection logic.

- [ ] Task: Update Configuration Schema
    - [ ] Add `RubyConfig` to `internal/config/schema/schema.go`.
    - [ ] Add `Ruby` field to `ModuleConfig`.
    - [ ] **Tests**: Add unit tests in `internal/config/schema/schema_test.go` for serialization and `IsEnabled` check.
- [ ] Task: Implement Ruby/Rails Detector (TDD)
    - [ ] Create fixtures in `testdata/fixtures/`:
        - [ ] `ruby-no-rails-docker`: Ruby (no Rails), root `docker-compose.yaml`, `backend/Gemfile`.
        - [ ] `rails-docker`: Rails, root `docker-compose.yaml`, `foo/Gemfile`, `foo/bin/rails`, `foo/config/application.rb`.
        - [ ] `ruby-no-rails-local`: Ruby (no Rails), no Docker, root `Gemfile`.
        - [ ] `rails-local`: Rails, no Docker, root `Gemfile`, `bin/rails`, `config/application.rb`.
    - [ ] Implement detection logic in `internal/detector/ruby.go`.
    - [ ] Add `RubyDetector` to the detection engine in `internal/detector/detector.go`.
    - [ ] **Tests**: Add unit tests in `internal/detector/detector_test.go` covering all four scenarios and environment manager detection.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Configuration and Detection' (Protocol in workflow.md)

## Phase 2: Core Task Implementation
Create the `RubyAction` and `RubyInstall` tasks.

- [ ] Task: Implement `RubyAction` Task
    - [ ] Create `internal/task/ruby.go`.
    - [ ] Implement logic for running `rails` and `bundle` commands.
    - [ ] Support Docker execution if a service is provided.
    - [ ] **Tests**: Add unit tests in `internal/task/ruby_test.go` using a mock executor to verify command construction.
- [ ] Task: Implement environment manager detection
    - [ ] Add helper to detect `rbenv`, `rvm`, or `asdf` based on version files.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Core Task Implementation' (Protocol in workflow.md)

## Phase 3: Strategy and Command Integration
Integrate Ruby tasks into the command dispatcher and standardized strategies.

- [ ] Task: Implement `RubyProvider`
    - [ ] Create `internal/strategy/ruby.go`.
    - [ ] Map "ruby migrate", "ruby console", etc., to strategies.
    - [ ] **Tests**: Add unit tests in `internal/strategy/strategy_test.go` for the `RubyProvider`.
- [ ] Task: Update `Build` and `Run` Strategies
    - [ ] Include Rails asset precompilation in `BuildStrategy`.
    - [ ] Include Rails server start in `RunStrategy`.
- [ ] Task: Add "ruby" top-level command
    - [ ] Create `internal/cmd/ruby.go` to expose specific Ruby tasks.
    - [ ] **Tests**: Add unit tests in `internal/cmd/ruby_test.go` for command-line argument parsing.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Strategy and Command Integration' (Protocol in workflow.md)

## Phase 4: Final Validation and PR
Final verification and cleanup.

- [ ] Task: Run full integration tests
    - [ ] Ensure all four Ruby fixtures are correctly handled in `integration_test.go`.
- [ ] Task: Finalization and Submission
    - [ ] Squash all track commits into a single clean commit.
    - [ ] Push the feature branch to GitHub and create a PR.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Final Validation and PR' (Protocol in workflow.md)
