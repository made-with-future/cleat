# Specification: Automated GitHub Releases and Installation Script

## Overview
This track implements an automated release pipeline using GitHub Actions. It triggers on version tags, cross-compiles the `cleat` binary for multiple platforms, packages them as compressed archives, and publishes them as a GitHub Release. Additionally, it updates the `README.md` with a smart installation script that handles different OS paths and decompression.

## Functional Requirements
- **Release Automation:**
    - Create a GitHub Actions workflow in `.github/workflows/release.yml`.
    - Trigger the workflow on pushes to tags matching `refs/tags/v*`.
    - Extract the version number from the git tag.
- **Cross-Compilation:**
    - Build binaries for `linux/amd64`, `darwin/arm64`, and `darwin/amd64` using `go build`.
    - Use the naming convention: `cleat_{version}_{os}_{arch}`.
- **Packaging & Distribution:**
    - Compress each binary into a `.tar.gz` archive.
    - Use the GitHub CLI (`gh`) to create the release and upload the archives.
    - Auto-populate release notes with download links for the artifacts.
- **Installation Experience:**
    - Add a `curl`-based installation command to `README.md`.
    - The script must detect the OS:
        - **Darwin (macOS):** Install to `/usr/local/bin` (may require `sudo`).
        - **Other (Linux):** Install to `$HOME/.local/bin`.
    - The script must handle downloading the appropriate `.tar.gz`, decompressing it, and ensuring the final binary is named `cleat` with execution permissions.
    - Provide URLs for both the "latest" version and the specific tagged version.

## Non-Functional Requirements
- **Efficiency:** Use GitHub Actions' built-in capabilities to minimize external dependencies.
- **Reliability:** Ensure the installation script handles missing directories (like `~/.local/bin`) gracefully.
- **Maintainability:** Use standard Go cross-compilation patterns.

## Acceptance Criteria
- [ ] Pushing a tag (e.g., `v1.0.0`) triggers the `release.yml` workflow.
- [ ] Three `.tar.gz` archives are successfully uploaded to a new GitHub Release.
- [ ] The `README.md` contains a working `curl` command that installs `cleat` correctly on both Linux and macOS.
- [ ] The installed binary is functional and correctly named `cleat`.

## Out of Scope
- Support for Windows (`windows/amd64`) in this initial release track.
- Support for package managers like Homebrew or APT.
