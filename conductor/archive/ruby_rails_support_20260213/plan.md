# Implementation Plan: Ruby and Ruby on Rails Support

This plan outlines the steps to add auto-detection and standardized command execution for Ruby and Ruby on Rails projects.

## Phase 1: Configuration and Detection [checkpoint: 0c47cd6]
Update the configuration schema and implement auto-detection logic.

- [x] Task: Update Configuration Schema
    - [x] Add `RubyConfig` to `internal/config/schema/schema.go`.
    - [x] Add `Ruby` field to `ModuleConfig`.
    - [x] **Tests**: Add unit tests in `internal/config/schema/schema_test.go` for serialization and `IsEnabled` check.
- [x] Task: Implement Ruby/Rails Detector (TDD)
    - [x] Create fixtures in `testdata/fixtures/`:
        - [x] `ruby-no-rails-docker`: Ruby (no Rails), root `docker-compose.yaml`, `backend/Gemfile`.
        - [x] `rails-docker`: Rails, root `docker-compose.yaml`, `foo/Gemfile`, `foo/bin/rails`, `foo/config/application.rb`.
        - [x] `ruby-no-rails-local`: Ruby (no Rails), no Docker, root `Gemfile`.
        - [x] `rails-local`: Rails, no Docker, root `Gemfile`, `bin/rails`, `config/application.rb`.
    - [x] Implement detection logic in `internal/detector/ruby.go`.
    - [x] Add `RubyDetector` to the detection engine in `internal/detector/detector.go`.
    - [x] **Tests**: Add unit tests in `internal/detector/detector_test.go` covering all four scenarios and environment manager detection.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Configuration and Detection' (Protocol in workflow.md)

## Phase 2: Core Task Implementation [checkpoint: ad1abf2]
Create the `RubyAction` and `RubyInstall` tasks.

- [x] Task: Implement `RubyAction` Task
    - [x] Create `internal/task/ruby.go`.
    - [x] Implement logic for running `rails` and `bundle` commands.
    - [x] Support Docker execution if a service is provided.
    - [x] **Tests**: Add unit tests in `internal/task/ruby_test.go` using a mock executor to verify command construction.
- [x] Task: Implement environment manager detection
    - [x] Add helper to detect `rbenv`, `rvm`, or `asdf` based on version files.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Core Task Implementation' (Protocol in workflow.md)

## Phase 3: Strategy and Command Integration [checkpoint: 3a3c595]
Integrate Ruby tasks into the command dispatcher and standardized strategies.

- [x] Task: Implement `RubyProvider`
    - [x] Create `internal/strategy/ruby.go`.
    - [x] Map "ruby migrate", "ruby console", etc., to strategies.
    - [x] **Tests**: Add unit tests in `internal/strategy/strategy_test.go` for the `RubyProvider`.
- [x] Task: Update `Build` and `Run` Strategies
    - [x] Include Rails asset precompilation in `BuildStrategy`.
    - [x] Include Rails server start in `RunStrategy`.
- [x] Task: Add "ruby" top-level command
    - [x] Create `internal/cmd/ruby.go` to expose specific Ruby tasks.
    - [x] **Tests**: Add unit tests in `internal/cmd/ruby_test.go` for command-line argument parsing.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Strategy and Command Integration' (Protocol in workflow.md)

## Phase 4: Final Validation and PR [checkpoint: 81ca545]
Final verification and cleanup.

- [x] Task: Run full integration tests
    - [x] Ensure all four Ruby fixtures are correctly handled in `integration_test.go`.
- [x] Task: Finalization and Submission
    - [x] Squash all track commits into a single clean commit.
    - [x] Push the feature branch to GitHub and create a PR.
- [x] Task: Conductor - User Manual Verification 'Phase 4: Final Validation and PR' (Protocol in workflow.md)
