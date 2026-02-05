# Specification: Refactor Configuration Logic

## Goal
Extract auto-detection logic from `internal/config` into a dedicated `internal/detector` package. Refocus `internal/config` on defining the configuration schema and handling explicit loading/parsing.

## Requirements
- Create a new package `internal/detector`.
- Move all auto-discovery logic (detecting Docker, Django, NPM, GCP, etc.) from `internal/config` to `internal/detector`.
- Define a clear interface for detectors.
- Update `internal/config` to use `internal/detector` for initial value population.
- Ensure 100% backward compatibility with existing `cleat.yaml` parsing.
- Maintain or exceed 80% test coverage for both packages.

## Architecture
- `internal/config`: Responsible for the `Config` struct, YAML unmarshaling, and validation.
- `internal/detector`: Responsible for scanning the filesystem and environment to infer project settings.