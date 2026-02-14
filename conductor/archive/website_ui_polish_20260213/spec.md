# Specification: Website UI Polish and Dracula Accents

## Overview
This track focuses on refining the visual aesthetic and usability of the Cleat landing page. It addresses specific layout issues, ensures external links behave as expected, and introduces subtle color accents from the Dracula theme to enhance the modern, high-fidelity feel of the project.

## Functional Requirements
- **Layout and Spacing Improvements:**
    - Add bottom padding to the main `<h1>` in the Hero section to prevent it from feeling cramped.
    - Perform a comprehensive spacing pass across all sections (Hero, Features, Quick Start) to ensure balanced vertical and horizontal margins.
    - Refine feature card internal padding and hover transitions for a more polished feel.
- **Link and Interaction Fixes:**
    - Update the "Install" link to use the actual installation command URL from the `README.md`.
    - Configure all external links (GitHub, Made With Future website) to open in a new browser window (`target="_blank" rel="noopener noreferrer"`).
- **Dracula Theme Integration (Subtle Accents):**
    - Integrate the Dracula color palette as subtle accents throughout the page:
        - **Purple (#bd93f9) / Pink (#ff79c6):** Use for gradients in the logo and hero headlines.
        - **Cyan (#8be9fd):** Use for feature card numbers and secondary highlights.
        - **Green (#50fa7b):** Use for the "Install" button and terminal success indicators (`==>`).
        - **Orange (#ffb86c) / Yellow (#f1fa8c):** Use for callouts or "Pro Tips."
    - Maintain the current deep black/zinc background (`bg-zinc-950`) for contrast.

## Non-Functional Requirements
- **Visual Consistency:** The Dracula accents should feel integrated and not overwhelming.
- **Accessibility:** Ensure that the new color combinations maintain sufficient contrast for readability.

## Acceptance Criteria
- [ ] The Hero `<h1>` has improved bottom padding.
- [ ] All sections have balanced vertical spacing.
- [ ] The "Install" link is corrected.
- [ ] GitHub and Made With Future links open in new windows.
- [ ] Dracula color accents are visible but subtle on key UI elements.
- [ ] The site remains responsive across all devices.

## Out of Scope
- Changing the underlying Tailwind-based implementation to a different CSS framework.
- Implementing a full dark/light mode toggle.
