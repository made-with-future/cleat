# Contributing to Cleat

Thank you for your interest in contributing to Cleat!

## Development

### Prerequisites

- [Go](https://go.dev/doc/install) - version specified in `go.mod`
- [Make](https://www.gnu.org/software/make/)

### Building from Source

To build the `cleat` binary for your local platform:

```bash
make build
```

This will create a `cleat` executable in the project root.

### Running in Development

To launch the TUI:

```bash
make run
```
*Press **'q'** or **'Ctrl+C'** to exit the TUI.*

### Running Tests

Execute the full test suite using the `Makefile`:

```bash
make test
```

### Code Coverage

Check test coverage and verify it meets the 70% threshold:

```bash
make coverage
```

Generate an HTML coverage report for detailed analysis:

```bash
make coverage-html
# Opens coverage.html in your browser
```

The CI pipeline enforces a minimum coverage of 70%. The `make coverage` command will fail if coverage drops below this threshold.

### Code Quality

Run formatting and linting checks:

```bash
make fmt
make vet
```

### Cross-Platform Builds

To build binaries for Linux and macOS (amd64 and arm64):

```bash
make build-all
```
