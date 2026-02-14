# Implementation Plan: Website UI Polish and Dracula Accents

This plan outlines the steps to refine the landing page layout, fix link behaviors, and integrate subtle Dracula theme accents.

## Phase 1: Layout and Interaction Fixes
Address specific spacing issues and ensure external links function correctly.

- [ ] Task: Adjust Hero spacing and H1 padding
    - [ ] Add bottom padding to the main `<h1>` in `site/index.html`.
    - [ ] Review and adjust `pt/pb` classes for Hero, Features, and Quick Start sections.
- [ ] Task: Fix link behaviors and destinations
    - [ ] Update the "Install" button link to the correct URL.
    - [ ] Add `target="_blank" rel="noopener noreferrer"` to all external links (GitHub, Made With Future).
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Layout and Interaction Fixes' (Protocol in workflow.md)

## Phase 2: Dracula Accent Integration
Introduce subtle colors from the Dracula palette to enhance the visual appeal.

- [ ] Task: Update Hero and Logo gradients
    - [ ] Replace current white/zinc gradient with a Dracula Purple/Pink gradient.
    - [ ] Apply Dracula Purple to the logo icon.
- [ ] Task: Enhance Feature cards and highlights
    - [ ] Change feature card numbers to Dracula Cyan.
    - [ ] Refine hover states to use a subtle Cyan glow or border.
- [ ] Task: Polish Buttons and Terminal
    - [ ] Update "Install" button to Dracula Green.
    - [ ] Update terminal success indicators (`==>`) to Dracula Green.
    - [ ] Apply Dracula Orange/Yellow to "Pro Tips" or other callouts.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Dracula Accent Integration' (Protocol in workflow.md)

## Phase 3: Final Validation and PR
Final verification and cleanup.

- [ ] Task: Final Polish
    - [ ] Check responsiveness across mobile/desktop views.
    - [ ] Verify accessibility (contrast) of new color accents.
- [ ] Task: Finalization and Submission
    - [ ] Squash commits and create PR.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Final Validation and PR' (Protocol in workflow.md)
