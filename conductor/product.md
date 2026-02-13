# Initial Concept

A unified orchestration interface for diverse engineering stacks, providing a declarative CLI and TUI to standardize development and operational tasks.

# Product Definition: Cleat

## Vision
Cleat aims to be the definitive "structural binding layer" for modern engineering teams. By providing a unified, declarative interface for disparate tools (Terraform, Docker, GCP, Django, NPM), Cleat eliminates operational drift and context-switching fatigue. It leans heavily into **convention over configuration**, prioritizing auto-detection of common patterns to minimize manual setup while providing a robust override mechanism for specific needs.

## Target Users
- **DevOps Engineers:** Managing complex, multi-service environments with diverse toolchains who need to standardize operational workflows.
- **Frontend & Backend Developers:** Seeking a friction-less, consistent local development experience across different projects and stacks.

## Core Goals
1. **Toolchain Standardization:** Wrap project-specific commands into a consistent `build` and `run` interface.
2. **Convention Over Configuration:** Prioritize auto-detection of tools, environments, and services to provide an "it just works" experience out of the box.
3. **Rich Visibility:** Provide a high-fidelity TUI (Terminal User Interface) for real-time monitoring of logs and service status.
4. **Operational Consistency:** Use declarative configuration (`cleat.yaml`) to ensure environment parity and reproducible executions when auto-detection is insufficient or needs overrides.

## Key Features
- **Intelligent Auto-Detection:** Automatically identifies project stacks (e.g., Docker, Django, NPM) and environment configurations without requiring an initial `cleat.yaml`.
- **Unified Build & Run:** Standardized entry points that adapt to the detected or defined project context.
- **Interactive TUI:** Real-time log streaming and service management powered by Bubble Tea.
- **Declarative Overrides:** A standardized `cleat.yaml` schema that allows users to supplement or override auto-detected settings.
- **Ecosystem Integration:** First-class support for Docker Compose, GCP, Terraform, and secret providers like 1Password.

