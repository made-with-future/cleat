# Implementation Plan: Website UI Polish and Dracula Accents

This plan outlines the steps to refine the landing page layout, fix link behaviors, and integrate subtle Dracula theme accents.

## Phase 1: Layout and Interaction Fixes
Address specific spacing issues and ensure external links function correctly.

- [x] Task: Adjust Hero spacing and H1 padding
    - [ ] Add bottom padding to the main `<h1>` in `site/index.html`.
    - [ ] Review and adjust `pt/pb` classes for Hero, Features, and Quick Start sections.
- [x] Task: Fix link behaviors and destinations
    - [ ] Update the "Install" button link to the correct URL.
    - [ ] Add `target="_blank" rel="noopener noreferrer"` to all external links (GitHub, Made With Future).
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Layout and Interaction Fixes' (Protocol in workflow.md)

## Phase 2: Dracula Accent Integration
Introduce subtle colors from the Dracula palette to enhance the visual appeal.

- [x] Task: Update Hero and Logo gradients
    - [x] Replace current white/zinc gradient with a Dracula Purple/Pink gradient.
    - [x] Apply Dracula Purple to the logo icon.
- [x] Task: Enhance Feature cards and highlights
    - [x] Change feature card numbers to Dracula Cyan.
    - [x] Refine hover states to use a subtle Cyan glow or border.
    - [x] Implement dynamic mouse-following glow effect on cards.
- [x] Task: Polish Buttons and Terminal
    - [x] Update "Install" button to Dracula Green.
    - [x] Update terminal success indicators (`==>`) to Dracula Green.
    - [x] Apply Dracula Orange/Yellow to "Pro Tips" or other callouts.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Dracula Accent Integration' (Protocol in workflow.md)

## Phase 3: Final Validation and PR
Final verification and cleanup.

- [x] Task: Final Polish
    - [x] Check responsiveness across mobile/desktop views.
    - [x] Verify accessibility (contrast) of new color accents.
- [x] Task: Finalization and Submission
    - [x] Squash commits and create PR.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Final Validation and PR' (Protocol in workflow.md)
