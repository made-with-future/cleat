# Technology Stack: Cleat

## Core Language & Runtime
- **Go (v1.25.5+):** Chosen for its performance, static typing, and excellent support for CLI/TUI applications.

## Primary Frameworks
- **CLI Framework:** `spf13/cobra` - The industry standard for building robust and discoverable CLI applications in Go.
- **TUI Framework:** `charmbracelet/bubbletea` - A powerful, Elm-inspired framework for building interactive and beautiful terminal user interfaces.
- **TUI Components:** `charmbracelet/bubbles` and `charmbracelet/lipgloss` for ready-to-use UI components and advanced styling.

## Configuration & Data
- **YAML:** `gopkg.in/yaml.v3` for parsing and generating the declarative `cleat.yaml` configuration files, as well as managing internal history and usage statistics.
- **Auto-Detection Engine:** A custom extensible detector package (`internal/detector`) for project auto-discovery.
- **Configuration Schema:** A centralized schema package (`internal/config/schema`) to maintain a clean separation between data structures and logic.

## Utilities & Terminal Handling
- **Terminal Management:** `golang.org/x/term`, `github.com/muesli/termenv`, and `github.com/charmbracelet/x/ansi` for handling terminal resizing, raw mode, and cross-platform ANSI escapes.

## Infrastructure & Integrations (Targeted)
- **Containerization:** Docker & Docker Compose.
- **Cloud Platform:** Google Cloud Platform (GCP).
- **IaC:** Terraform.
- **Secrets:** 1Password CLI (`op`).
