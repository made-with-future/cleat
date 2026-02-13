# Configuration Reference

[Back to README](../README.md)

Cleat uses a declarative configuration file, `cleat.yaml`, to understand your project structure and orchestrate your toolchain. This page provides a comprehensive reference for all available configuration options.

## The `cleat.yaml` File

By default, Cleat looks for a `cleat.yaml` file in your project's root directory. While Cleat features intelligent auto-detection for many common patterns, the configuration file allows you to explicitly define services, modules, and custom workflows.

### Root Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `version` | integer | The version of the Cleat configuration schema. | `1` |
| `docker` | boolean | Global toggle for Docker Compose support. | `true` if `docker-compose.yaml` exists. |
| `envs` | list | List of environment names (used for Terraform, etc.). | Auto-detected from `.envs/*.env` if omitted. |
| `google_cloud_platform` | object | GCP specific configuration. See [GCP Configuration](#gcp-configuration). | |
| `terraform` | object | Terraform specific configuration. | |
| `services` | list | List of services for multi-service repositories. See [Service Configuration](#service-configuration). | |

### Service Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `name` | string | Unique name for the service. | |
| `dir` | string | Directory path relative to project root. | |
| `docker` | boolean | Whether this service uses Docker. | `true` if `docker-compose.yaml` exists in `dir`. |
| `modules` | list | List of modules (stacks) within the service. See [Module Configuration](#module-configuration). | |

### Module Configuration

A module is a specific stack within a service (e.g., Python/Django or NPM/Frontend).

| Field | Type | Description |
| :--- | :--- | :--- |
| `python` | object | Python stack configuration. See [Python Configuration](#python-configuration). |
| `npm` | object | NPM stack configuration. See [NPM Configuration](#npm-configuration). |

### Python Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `django` | boolean | Whether this is a Django project. | `true` if `manage.py` is found. |
| `django_service` | string | Docker Compose service name for Django tasks. | Service name |
| `package_manager` | string | Python package manager (`uv`, `pip`, `poetry`). | `uv` |

### NPM Configuration

| Field | Type | Description | Default / Auto-detection |
| :--- | :--- | :--- | :--- |
| `service` | string | Docker Compose service name for NPM scripts. | Service name |
| `scripts` | list | List of NPM scripts to run during build. | Auto-detected from `package.json` if omitted. |

### GCP Configuration

| Field | Type | Description |
| :--- | :--- | :--- |
| `project_name` | string | Google Cloud Project ID. |
| `account` | string | (Optional) GCP account/email to use. |

### Example

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
