# Product Guidelines: Cleat

## Tone & Voice
- **Professional & Efficient:** User-facing messages and CLI output should be direct and concise. Avoid unnecessary jargon or "chatty" feedback. Focus on providing immediate value and actionable information.
- **Modern & Minimalist:** Use clean layouts in both CLI and TUI. Prioritize essential information and use whitespace (or terminal equivalents) to create a sense of order and focus.
- **Reliable & Authoritative:** Communication should inspire confidence. Errors should be clearly explained with actionable remediation steps.

## Visual Identity (TUI)
- **Material-Inspired Hierarchy:** Use a structured layout for the TUI, with clear separation between service lists, status indicators, and log streams.
- **Functional Color Palette:**
    - **Success:** Green for active services and successful task completion.
    - **Warning:** Yellow for pending tasks or minor issues.
    - **Error:** Red for failed services or critical errors.
    - **Info:** Blue or subtle gray for logs and general information.
- **Readability First:** Use high-contrast text and clear glyphs (where supported) to ensure the TUI is easy to read across different terminal themes.

## Documentation Standards
- **Standardized Formats:** Use Markdown for all documentation.
- **Code-First Examples:** Documentation should prioritize clear, copy-pasteable examples of `cleat.yaml` configurations and CLI commands.
- **Implicit Over Explicit:** Leverage the "convention over configuration" philosophy in documentation by showing the simplest auto-detected use case first.
