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

It automatically handles `docker compose up`, Django's `runserver`, or `npm start` based on your project's configuration. It also features built-in support for [1Password CLI](https://developer.1password.com/docs/cli/) (`op`) if a `.envs/dev.env` file is detected.

### Configuration (`cleat.yaml`)

Cleat is configured via a `cleat.yaml` file in your project root. This file tells Cleat about your stack and how to orchestrate build commands.

#### Configuration Reference

##### Root Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `version` | integer | The version of the Cleat configuration schema. | `1` |
| `docker` | boolean | Global toggle for Docker Compose support. | `true` if `docker-compose.yaml` exists. |
| `envs` | list | List of environment names (used for Terraform, etc.). | Auto-detected from `.envs/*.env` if omitted. |
| `google_cloud_platform` | object | GCP specific configuration. See [GCP Configuration](#gcp-configuration). | |
| `terraform` | object | Terraform specific configuration. | |
| `services` | list | List of services for multi-service repositories. See [Service Configuration](#service-configuration). | |

##### Service Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `name` | string | Unique name for the service. | |
| `dir` | string | Directory path relative to project root. | |
| `docker` | boolean | Whether this service uses Docker. | `true` if `docker-compose.yaml` exists in `dir`. |
| `modules` | list | List of modules (stacks) within the service. See [Module Configuration](#module-configuration). | |

##### Module Configuration

A module is a specific stack within a service (e.g., Python/Django or NPM/Frontend).

| Field | Type | Description |
| :--- | :--- | :--- |
| `python` | object | Python stack configuration. See [Python Configuration](#python-configuration). |
| `npm` | object | NPM stack configuration. See [NPM Configuration](#npm-configuration). |

##### Python Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `django` | boolean | Whether this is a Django project. | `true` if `manage.py` is found. |
| `django_service` | string | Docker Compose service name for Django tasks. | `backend` |
| `package_manager` | string | Python package manager (`uv`, `pip`, `poetry`). | `uv` |

##### NPM Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `service` | string | Docker Compose service name for NPM scripts. | `backend-node` (if docker enabled) |
| `scripts` | list | List of NPM scripts to run during build. | Auto-detected from `package.json` if omitted. |

##### GCP Configuration

| Field | Type | Description |
| :--- | :--- | :--- |
| `project_name` | string | Google Cloud Project ID. |
| `account` | string | (Optional) GCP account/email to use. |

#### Example

```yaml
version: 1
docker: true
services:
  - name: my-app
    dir: .
    modules:
      - python:
          django: true
          package_manager: uv
      - npm:
          scripts:
            - build
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
