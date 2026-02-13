# Cleat

*The unified orchestration interface for diverse engineering stacks.*

Cleat is a declarative CLI and TUI designed to standardize development and operational tasks across distributed projects. It acts as a structural binding layer, wrapping your project's underlying toolchain—whether that’s Terraform, Go, Django, Google Cloud SDK, or Docker—into a single, consistent entry point.

## The Problem: Operational Tax

As engineering organizations grow, the "Operational Tax"—the mental overhead required to switch between different project stacks—skyrockets. 

* **The Fragmentation:** One project uses `make`, another uses `npm`, and a third uses a bash script that only the original author understands.
* **The Context Switch:** Developers spend more time remembering *how* to run a service than actually *improving* it.
* **The "Works on My Machine" Trap:** Manual sequences of `docker`, `gcloud`, and `terraform` commands are brittle and prone to human error.

Cleat eliminates this friction by providing a unified, declarative entry point for every project. It auto-detects your stack and provides a standardized interface for common tasks, allowing you to focus on the code, not the plumbing.

## Features

### Standardized Commands
Cleat provides standardized commands that adapt to your project's stack. Run common operations like `build` or `run` without needing to remember the underlying toolchain specificities.

```bash
# Cleat knows if it should run 'npm run build', 'go build', or 'docker compose build'
cleat build
```

### Workflows
For more complex sequences of tasks, Cleat supports custom workflows. Define them in `cleat.yaml` to orchestrate multiple steps into a single named command.

```yaml
# Example workflow definition in cleat.yaml
workflows:
  - name: deploy-prod
    commands:
      - build
      - terraform apply:production
      - gcp app-engine deploy
```

### Intelligent Auto-Detection
Cleat automatically identifies your project's stack—Docker, Go, Django, NPM, Terraform, GCP, and Ruby—providing an "it just works" experience with zero manual configuration for most standard layouts.

```text
==> Auto-detected project context:
    - Docker: found docker-compose.yaml
    - Django: found manage.py (service: backend)
    - NPM: found package.json (service: frontend)
```

## Quick Start

1. **Install:** `curl -fsSL https://cleat.sh | sh`
2. **Launch:** Run `cleat` in any project root.
3. **Automate:** Explore auto-detected tasks or define a `cleat.yaml`.

---

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

Launch the interactive TUI to see all available commands for your project:

```bash
cleat
```

You can also run specific commands directly:

```bash
# Run a service-specific command
cleat docker up

# Run a named workflow
cleat workflow ci
```

### Pro Tip
Add `cleat` to your project's `README.md` as the recommended way to get started. It eliminates the need for long setup documents.

## Configuration (`cleat.yaml`)

Cleat uses a `cleat.yaml` file in your project root to orchestrate your toolchain. While many projects work out of the box thanks to auto-detection, the configuration file allows you to define services, modules (like Django or NPM), and custom workflows.

For a full reference of all available options, see our [Configuration Documentation](docs/configuration.md).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to set up your development environment and run tests.
