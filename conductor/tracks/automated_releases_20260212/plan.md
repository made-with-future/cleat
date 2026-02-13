# Implementation Plan: Automated GitHub Releases and Installation Script

This plan implements a robust CI/CD pipeline for releasing `cleat` binaries and provides a user-friendly installation script.

## Phase 1: GitHub Actions Workflow Scaffolding
Set up the basic workflow structure and trigger mechanism.

- [ ] Task: Create `.github/workflows/release.yml`
    - [ ] Define the workflow name and trigger on `push` to `refs/tags/v*`.
    - [ ] Set up permissions for `contents: write` (required for releases).
- [ ] Task: Implement version extraction
    - [ ] Add a step to extract the version from `${{ github.ref_name }}`.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: GitHub Actions Workflow Scaffolding' (Protocol in workflow.md)

## Phase 2: Cross-Compilation and Packaging
Configure the build matrix and packaging logic.

- [ ] Task: Implement the build matrix
    - [ ] Define a matrix for `os` (linux, darwin) and `arch` (amd64, arm64).
    - [ ] Exclude `linux/arm64` if not required by spec (spec says linux/amd64, darwin/arm64, darwin/amd64).
- [ ] Task: Implement Go build and compression
    - [ ] Use `go build` with `GOOS` and `GOARCH` environment variables.
    - [ ] Implement `tar -czf` to create archives named `cleat_${version}_${os}_${arch}.tar.gz`.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Cross-Compilation and Packaging' (Protocol in workflow.md)

## Phase 3: Release Management
Integrate GitHub CLI for automated release creation and artifact uploading.

- [ ] Task: Implement release creation
    - [ ] Use `gh release create ${{ github.ref_name }} --generate-notes` to initialize the release.
- [ ] Task: Implement artifact upload
    - [ ] Use `gh release upload ${{ github.ref_name }} ./dist/*.tar.gz` to attach all archives.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Release Management' (Protocol in workflow.md)

## Phase 4: Installation Script and Documentation
Create the installation script and update the project README.

- [ ] Task: Develop the installation script logic
    - [ ] Draft a shell script that detects OS, determines the target directory (`/usr/local/bin` for Darwin, `~/.local/bin` for Linux), downloads the latest release, decompresses it, and moves the binary to the target.
- [ ] Task: Update `README.md`
    - [ ] Add the "Installation" section with the `curl` command.
    - [ ] Ensure URLs for both "latest" and version-specific downloads are documented.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Installation Script and Documentation' (Protocol in workflow.md)
