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

### Unified Build Command
Cleat provides a standardized `build` command that adapts to your project's stack. By defining a `cleat.yaml` file in your project root, you can orchestrate complex build steps across Docker, Django, and NPM with a single command:

```bash
cleat build
```

### Unified Run Command
Similarly, the `run` command provides a consistent way to start your project's development environment:

```bash
cleat run
```

It automatically handles `docker compose up`, Django's `runserver`, or `npm start` based on your project's configuration. It also features built-in support for [1Password CLI](https://developer.1password.com/docs/cli/) (`op`) if a `.env/dev.env` file is detected.

### Configuration (`cleat.yaml`)

Cleat is configured via a `cleat.yaml` file in your project root. This file tells Cleat about your stack and how to orchestrate build commands.

#### Configuration Reference

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `version` | integer | The version of the Cleat configuration schema. | `1` |
| `docker` | boolean | Whether to use Docker Compose for build steps. | `true` if `docker-compose.yaml` exists. |
| `python` | object | Python build configuration. | |
| `python.django` | boolean | Whether this is a Django project. Enables `collectstatic` during build. | `false` |
| `python.django_service` | string | The name of the Docker Compose service for Django tasks. | `backend` |
| `npm` | object | NPM build configuration. | |
| `npm.service` | string | The name of the Docker Compose service for running NPM scripts. | `backend-node` (if `docker` is true) |
| `npm.scripts` | list | List of NPM scripts to run during the build process. | `["build"]` if `frontend/package.json` exists. |

#### Example

```yaml
version: 1
docker: true
python:
  django: true
  django_service: backend
npm:
  service: backend-node
  scripts:
    - build
    - tailwindcss-build
```

Cleat automatically handles whether commands should run locally or within Docker containers based on your configuration.
    
## Getting Started

### Installation

Cleat is distributed as a single binary. For now, please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to build it from source.

### Basic Usage

Once you have the `cleat` binary in your path, you can use it to orchestrate your project:

```bash
# Build your project (runs docker-compose build, npm scripts, etc.)
cleat build

# Start your development environment
cleat run

# Launch the interactive TUI
cleat
```

To check the version:

```bash
cleat version
```

### Example

You can test Cleat using the included example project:

```bash
cd examples/test-project
../../cleat build
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to set up your development environment and run tests.
