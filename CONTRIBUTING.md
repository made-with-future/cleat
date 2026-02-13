# Contributing to Cleat

Thank you for your interest in contributing to Cleat!

## Development

### Prerequisites

- [Go](https://go.dev/doc/install) - version specified in `go.mod`
- [Cleat](README.md#installation) - Recommended for orchestration, but you can also use Go directly.

### Building from Source

To build the `cleat` binary for your local platform:

```bash
cleat build
```

This will create a `cleat` executable in the project root.

### Running in Development

To launch the interactive TUI:

```bash
./cleat
```
*Press **'q'** or **'Ctrl+C'** to exit the TUI.*

### Running Tests

Execute the full test suite:

```bash
cleat go test
```

### Code Coverage

Check test coverage and verify it meets the 70% threshold:

```bash
cleat workflow coverage
```

The CI pipeline enforces a minimum coverage of 70%.

### Code Quality

Run formatting and linting checks:

```bash
cleat go fmt
cleat go vet
```

### Cross-Platform Builds

To build binaries for Linux and macOS (amd64 and arm64):

```bash
cleat workflow build-all
```

## Dependency Updates (Dependabot)

Cleat uses GitHub Dependabot to automate dependency updates for Go modules and GitHub Actions.

### Handling Dependabot PRs

1. **Verify CI:** Wait for the CI pipeline to complete. If the automated tests fail, investigate the dependency's changelog for breaking changes.
2. **Local Testing:** If the update is critical (e.g., a major Go version or a core library), it is recommended to checkout the Dependabot branch and run `cleat go test` locally.
3. **Merging:** PRs can be merged once CI passes. Ensure that the dependency update doesn't introduce regressions in the TUI or command execution logic.
