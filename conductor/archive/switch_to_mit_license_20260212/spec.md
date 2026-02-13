# Specification: Transition to MIT License

## Overview
This track involves transitioning the project's licensing from the GNU General Public License version 3 (GPLv3) to the MIT License. The goal is to increase the project's permissiveness, making it more accessible for a wider range of users and contributors.

## Functional Requirements
- **Update License File:**
    - Replace the content of the `LICENSE` file in the project root with the standard MIT License text.
    - Set the copyright notice to: `Copyright (c) 2026 Josh Turmel`.
- **Update Project Metadata:**
    - Verify and update any license references in project metadata files (e.g., `go.mod` comments or other configuration files if they contain license identifiers).
- **Audit for GPL:**
    - Ensure no GPL-specific boilerplate or references remain in the project's active codebase.

## Non-Functional Requirements
- **Clarity:** Ensure the new license is clearly visible and correctly attributed.
- **Consistency:** All relevant metadata should reflect the new license.

## Acceptance Criteria
- [ ] The `LICENSE` file in the project root contains the MIT License text.
- [ ] The `LICENSE` file contains the line `Copyright (c) 2026 Josh Turmel`.
- [ ] A grep for "GPL" and "GNU General Public License" returns no results in source files.
- [ ] The project continues to build and pass all tests (including coverage >70%) after the changes.

## Out of Scope
- Changing the licensing of third-party dependencies.
- Detailed legal audit of the codebase.
