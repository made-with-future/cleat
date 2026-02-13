# Cleat

*The unified orchestration interface for divergent engineering stacks.*

Cleat is a declarative CLI and TUI designed to standardize development and operational tasks across distributed projects. It acts as a structural binding layer, wrapping your project's underlying toolchain—whether that’s Terraform, Django, Google Cloud SDK, or Docker—into a single, consistent entry point.

## The Problem: Operational Drift

In a microservices or multi-project environment, operational capabilities inevitably diverge.

* **Project A** uses `make deploy` to push to Kubernetes.
* **Project B** uses a custom shell script to run Terraform.
* **Project C** requires a complex sequence of `gcloud` commands just to initialize the local environment.

This entropy forces developers to memorize project-specific idiosyncrasies, leading to context-switching fatigue and execution errors.

## Features

### Standardized Commands
Cleat provides standardized commands that adapt to your project's stack. By defining a `cleat.yaml` file in your project root, Cleat will auto-detect your tools (Docker, Go, Django, NPM, Terraform, GCP) and provide a consistent interface to interact with them via the TUI or CLI.

### Workflows
For more complex sequences of tasks, Cleat supports custom workflows. Define them in `cleat.yaml` to orchestrate multiple steps into a single named command.

## Installation

To install the latest version of Cleat:

```bash
curl -fsSL https://raw.githubusercontent.com/made-with-future/cleat/main/install.sh | sh
```

To install a specific version (e.g., v1.0.0):

```bash
curl -fsSL https://raw.githubusercontent.com/made-with-future/cleat/main/install.sh | sh -s -- v1.0.0
```

For source builds, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Basic Usage

Once you have the `cleat` binary in your path, you can use it to orchestrate your project. Launch the interactive TUI to see all available commands for your project:

```bash
cleat
```

You can also run specific commands directly:

```bash
# Example: run a service-specific command
cleat docker up

# Example: run a named workflow
cleat workflow ci
```

To check the version:

```bash
cleat version
```

### Example

You can test Cleat using the included example project:

```bash
cd examples/test-project
../../cleat
```

## Configuration (`cleat.yaml`)

Cleat uses a `cleat.yaml` file in your project root to orchestrate your toolchain. While many projects work out of the box thanks to auto-detection, the configuration file allows you to define services, modules (like Django or NPM), and custom workflows.

For a full reference of all available options, see our [Configuration Documentation](docs/configuration.md).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to set up your development environment and run tests.
