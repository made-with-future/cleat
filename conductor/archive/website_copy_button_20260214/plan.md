# Implementation Plan: Website Terminal Copy Button

This plan outlines the steps to add a copy-to-clipboard button to the terminal component on the project website.

## Phase 1: Structure & Styling [checkpoint: b7d9559]
Add the HTML and CSS for the copy button to `site/index.html`.

- [x] Task: Update `site/index.html` structure 76f676b
    - [x] Wrap the terminal body (`p-6 mono ...`) in a relative container to allow absolute positioning of the button.
    - [x] Insert the `<button>` element with the clipboard SVG icon.
- [x] Task: Add CSS for the pulse animation 8bf5391
    - [x] Add a `@keyframes pulse` to the `<style>` block in `site/index.html`.
    - [x] Define a `.copied` class that triggers the background color change and animation.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Structure & Styling' (Protocol in workflow.md)

## Phase 2: Interactivity [checkpoint: 4d4871a]
Implement the JavaScript logic for copying text and providing feedback.

- [x] Task: Implement `copyToClipboard` function d85954c
    - [x] Add a script block (or update existing one) to handle the click event.
    - [x] Use `navigator.clipboard.writeText()` to copy the install command.
    - [x] Toggle the `.copied` class and use `setTimeout` to revert it after 2 seconds.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Interactivity' (Protocol in workflow.md)

## Phase 3: Final Validation
Verify the functionality across different screen sizes and ensure it meets the design spec.

- [x] Task: Visual Polish & Responsiveness 92987c7
    - [x] Ensure the button doesn't overlap text on small screens.
    - [x] Verify the Dracula theme colors are accurate.
- [x] Task: Finalization and Submission
    - [x] Perform a final manual check of the copy functionality.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Final Validation' (Protocol in workflow.md)
