# Initial Concept

The unified orchestration interface for diverse engineering stacks, providing a declarative CLI and TUI to standardize development and operational tasks.

# Product Definition: Cleat

## Vision
Cleat aims to be the definitive "structural binding layer" for modern engineering teams. By providing a unified, declarative interface for diverse stacks, Cleat eliminates the "Operational Tax" and context-switching fatigue. It leans heavily into **convention over configuration**, prioritizing auto-detection of common patterns to provide a standardized, "it just works" experience across heterogeneous toolchains.

## Target Users
- **DevOps Engineers:** Managing complex, multi-service environments with diverse toolchains who need to standardize operational workflows.
- **Frontend & Backend Developers:** Seeking a friction-less, consistent local development experience across different projects and stacks.

## The Problem: Operational Tax
As engineering organizations grow, the mental overhead required to switch between different project stacks—the "Operational Tax"—skyrockets.

1. **Fragmentation:** One project uses `make`, another uses `npm`, and a third uses custom scripts. This inconsistency forces developers to memorize project-specific idiosyncrasies.
2. **Context Switching:** Developers spend more time remembering *how* to run a service than actually *improving* it, leading to fatigue and reduced productivity.
3. **Human Error:** Manual sequences of `docker`, `gcloud`, and `terraform` commands are brittle and prone to mistakes.

## Core Goals
1. **Toolchain Standardization:** Wrap project-specific commands into a consistent `build` and `run` interface that adapts to the detected stack.
2. **Convention Over Configuration:** Prioritize auto-detection of tools, environments, and services to minimize manual setup and "plumbing."
3. **Rich Visibility:** Provide a high-fidelity TUI (Terminal User Interface) for real-time monitoring of logs and service management.
4. **Operational Consistency:** Use declarative configuration (`cleat.yaml`) to ensure environment parity and reproducible executions when auto-detection needs supplements or overrides.

## Key Features
- **Intelligent Auto-Detection:** Automatically identifies project stacks (Docker, Go, Django, NPM, Terraform, GCP, Ruby) with zero manual configuration for most standard layouts.
- **Unified Build & Run:** Standardized entry points that provide a single, consistent way to build and run any project, regardless of its underlying technology.
- **Interactive TUI:** Real-time log streaming and interactive service management powered by Bubble Tea for a superior developer experience.
- **Declarative Overrides:** A standardized `cleat.yaml` schema that allows users to supplement or override auto-detected settings for advanced orchestration.
- **Ecosystem Integration:** First-class support for Docker Compose, GCP, Terraform, and secret providers like 1Password.
