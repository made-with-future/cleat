# Specification: Website Terminal Copy Button

## Overview
This track adds a "Copy" UI control to the terminal area on the Cleat project website (`site/index.html`). This allows users to easily copy the installation command to their clipboard, improving the onboarding experience.

## Functional Requirements
- **Copy to Clipboard:** Clicking the button must copy the string `curl -fsSL https://made-with-future.github.io/cleat/install.sh | sh` to the user's system clipboard.
- **Visual Feedback:** Upon a successful copy, the button must provide a subtle color pulse (using the Dracula Green palette: `#50fa7b`) to indicate success.
- **Persistence:** The feedback should be temporary, reverting the button to its original state after a short duration (e.g., 2 seconds).

## UI/UX Design
- **Placement:** The button will be positioned as a floating element in the top-right corner of the terminal body area (the dark `#282a36` section).
- **Styling:**
    - **Shape:** Small and rounded.
    - **Colors:** Background using `var(--dracula-selection)` (#44475a) and foreground/icon using `var(--dracula-fg)` (#f8f8f2).
    - **Icon:** A standard clipboard icon (SVG).
- **Behavior:** The button should be clearly visible but not distracting from the command text itself.

## Technical Constraints
- **Zero Dependencies:** The implementation should use vanilla JavaScript (`navigator.clipboard.api`) and existing CSS frameworks (Tailwind CSS) already present in `index.html`.
- **Cross-Browser Compatibility:** Ensure the copy functionality works in all modern browsers supported by the site.

## Acceptance Criteria
- [ ] A "Copy" button is visible in the top-right corner of the terminal body on `site/index.html`.
- [ ] Clicking the button copies the correct install command to the clipboard.
- [ ] The button pulses green briefly after a successful copy.
- [ ] The implementation follows the existing Dracula theme and Tailwind CSS patterns.

## Out of Scope
- Adding copy buttons to other code blocks or sections.
- Implementing complex tooltip systems.
