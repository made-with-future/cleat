# Specification: README and Documentation Restructuring

## Overview
This track aims to simplify and improve the project's primary documentation by restructuring the `README.md` and moving detailed configuration technicalities to a dedicated page. The goal is to make the project more approachable for new users while maintaining comprehensive reference material.

## Functional Requirements
- **README Restructuring:**
    - Move the "Installation" section higher in the `README.md`, specifically after the "Features" section but before "Basic Usage".
    - Simplify the `README.md` by removing detailed configuration tables and complex examples.
    - Provide a high-level summary of the `cleat.yaml` configuration's purpose in the `README.md`.
    - Add a clear link in the `README.md` to the new dedicated configuration documentation.
- **Dedicated Configuration Page:**
    - Create a new file at `docs/configuration.md`.
    - Move the full "Configuration Reference" (tables, field descriptions, etc.) from `README.md` to this new file.
    - Move complex `cleat.yaml` examples to this new file.
    - Ensure `docs/configuration.md` is well-formatted and easy to navigate.

## Non-Functional Requirements
- **Clarity:** The documentation should be easier to read and scan.
- **Discoverability:** The transition from `README.md` to `docs/configuration.md` should be seamless for users seeking more detail.

## Acceptance Criteria
- [ ] `README.md` is significantly shorter and more focused on introduction and quick start.
- [ ] "Installation" section is positioned correctly after "Features".
- [ ] `docs/configuration.md` contains the full technical configuration reference.
- [ ] A link to the configuration documentation exists in the `README.md`.
- [ ] All internal and external links in the documents remain functional.

## Out of Scope
- Rewriting the content of the configuration reference (this is a move/refactor of existing text).
- Creating other documentation pages (e.g., `CONTRIBUTING.md` refactor).
