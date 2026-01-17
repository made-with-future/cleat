# Contributing to Cleat

Thank you for your interest in contributing to Cleat!

## Development

### Prerequisites

- [Go](https://go.dev/doc/install) 1.22 or later
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
