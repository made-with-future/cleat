# Implementation Plan: Automated GitHub Releases and Installation Script

This plan implements a robust CI/CD pipeline for releasing `cleat` binaries and provides a user-friendly installation script.

## Phase 1: GitHub Actions Workflow Scaffolding [checkpoint: 28d91ef]
Set up the basic workflow structure and trigger mechanism.

- [x] Task: Create `.github/workflows/release.yml` [edae63e]
    - [ ] Define the workflow name and trigger on `push` to `refs/tags/v*`.
    - [ ] Set up permissions for `contents: write` (required for releases).
- [x] Task: Implement version extraction [edae63e]
    - [ ] Add a step to extract the version from `${{ github.ref_name }}`.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: GitHub Actions Workflow Scaffolding' (Protocol in workflow.md)

## Phase 2: Cross-Compilation and Packaging [checkpoint: 734f925]
Configure the build matrix and packaging logic.

- [x] Task: Implement the build matrix [996d9b4]
    - [ ] Define a matrix for `os` (linux, darwin) and `arch` (amd64, arm64).
    - [ ] Exclude `linux/arm64` if not required by spec (spec says linux/amd64, darwin/arm64, darwin/amd64).
- [x] Task: Implement Go build and compression [996d9b4]
    - [ ] Use `go build` with `GOOS` and `GOARCH` environment variables.
    - [ ] Implement `tar -czf` to create archives named `cleat_${version}_${os}_${arch}.tar.gz`.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Cross-Compilation and Packaging' (Protocol in workflow.md)

## Phase 3: Release Management [checkpoint: cab9ee2]
Integrate GitHub CLI for automated release creation and artifact uploading.

- [x] Task: Implement release creation [9773cbe]
    - [ ] Use `gh release create ${{ github.ref_name }} --generate-notes` to initialize the release.
- [x] Task: Implement artifact upload [9773cbe]
    - [ ] Use `gh release upload ${{ github.ref_name }} ./dist/*.tar.gz` to attach all archives.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Release Management' (Protocol in workflow.md)

## Phase 4: Installation Script and Documentation [checkpoint: 61c6246]
Create the installation script and update the project README.

- [x] Task: Develop the installation script logic [125049a]
    - [ ] Draft a shell script that detects OS, determines the target directory (`/usr/local/bin` for Darwin, `~/.local/bin` for Linux), downloads the latest release, decompresses it, and moves the binary to the target.
- [x] Task: Update `README.md` [1f58a9e]
    - [ ] Add the "Installation" section with the `curl` command.
    - [ ] Ensure URLs for both "latest" and version-specific downloads are documented.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Installation Script and Documentation' (Protocol in workflow.md)
