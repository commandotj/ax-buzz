# RFC 0009: Buzz Desktop Ancient Greek Luxury Visual System (child of 0001)

**Status:** Approved

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Give Buzz.app a colorful luxury identity that fuses ancient Greek architecture and celestial engineering with a modern science-fiction instrument. Greek column rhythm, meander geometry, painted pottery color, and Antikythera-like orbital precision ground the visual language. Define two complete appearances: a sunlit marble Light theme and a separate deep-space Dark theme. `System` switches between those complete themes with macOS appearance. Preserve the Homebrew manager's behavior, layout, virtualization, and accessibility. Replace the README screenshot only after the real app implements this direction.

## Problem

The current Night City treatment and prior mockups reduced Greek Sci-Fi to generic dark panels, a Greek-key strip, and a column mark. The Light and Dark appearances were also mixed into one screen, so neither read as a complete theme. Buzz needs a coherent Greek Sci-Fi design language rooted in architecture, painted pottery, and celestial engineering, with two visibly distinct, internally consistent themes.

## Goals

- Make Greek Sci-Fi unmistakable through classical proportion, architectural rhythm, meander geometry, pottery color, and celestial-mechanism precision.
- Design Light and Dark as separate visual systems, not a shared surface with a few colors inverted.
- Light theme: sunlit marble, Aegean and lapis blue, terracotta, olive, and restrained bronze; surfaces remain light throughout.
- Dark theme: deep Aegean and midnight indigo, luminous cyan/lapis signals, terracotta or violet accents, and warm ivory text; surfaces remain dark throughout.
- Make luxury read through material, proportion, typography, and precise construction. Reserve luminous effects for small sci-fi signals; avoid gold overload, decorative shine, or generic neon.
- Preserve `System`, `Light`, and `Dark` selection and saved preference. `System` must switch the entire appearance between the separate Light and Dark systems.
- On macOS, remove the visible system title bar, retain native Traffic Light controls, and provide a draggable branded top rail.
- Keep package names, sources, versions, status, and actions immediately scannable.
- Update the README desktop screenshot from the implemented Buzz.app after implementation.

## Non-goals

- New Homebrew operations, package sources, inventory rules, or CLI protocol changes.
- More appearance modes or a theme marketplace.
- Reproducing Cyberpunk 2077, Greek-themed games, or other franchise UI, marks, or artwork.
- Pixel art, 8-bit graphics, retro teletext, or neon-noir styling.
- A museum, temple illustration, promotional hero, or decorative content that displaces package data.

## Design

### Visual direction: Greek celestial instrument

Treat Buzz as a precision instrument from a future Greek observatory. Build hierarchy from the real forms of Greek architecture—column rhythm, entablature, carved edges—and the Antikythera mechanism's measured orbital geometry. Use a purposeful meander as a structural rule, not a repeated decorative border. Make package rows and versions the main content; Greek and celestial forms belong in the shell, navigation, and selected-package detail.

**Light is a complete sunlit theme.** Use luminous marble and pale stone surfaces across the full app, with Aegean/lapis blue for structure, terracotta and olive for selected semantic accents, and bronze for fine architectural rules. Keep controls and inspector light too; do not insert dark panels into this theme.

**Dark is a complete celestial theme.** Use layered midnight Aegean, indigo, and lapis surfaces throughout, with luminous cyan and blue for active signals, terracotta or ion-violet for secondary meaning, and warm ivory text. Keep every panel dark; do not insert marble-white cards into this theme. Use small, precise light emissions to suggest a celestial instrument, never broad neon fields.

The two appearances share shape, type hierarchy, spacing, iconography, and semantic state meaning. Their surface colors, contrast, and light behavior remain distinct. Luxury comes from precise proportions and material craft: carved marble and painted ceramic in Light; deep enamel, bronze, and controlled luminous indicators in Dark. Avoid gold gradients, glass blur, excessive ornament, pseudo-antique texture, and copied game or franchise UI.

Use an expressive, legible display face for major headings; use a clear sans serif for controls and descriptions; reserve monospace for versions, identifiers, and source data. Preserve the existing navigation, package-list, and inspector composition, compact rows, tabular values, familiar macOS behavior, and package-data-first hierarchy.

On macOS, hide the visible system title-bar title and use a branded, draggable top rail that reserves space for the native Traffic Light controls. Keep the native controls and their standard close, minimize, and zoom behavior; do not draw fake replacements.

This RFC replaces RFC 0008's visual direction and palette only. RFC 0008's management behavior, layout, accessibility, virtualization, and scrolling requirements remain in force. Implementation updates `DESIGN.md`, `.impeccable/design.json`, desktop theme tokens and components, and the README screenshot. It does not change Homebrew behavior or the CLI boundary.

### Appearance behavior

Keep the existing Appearance choices: `System`, `Light`, and `Dark`. `System` follows macOS changes while Buzz is open and switches the whole interface between the complete Light and Dark themes. Explicit Light and Dark override System and persist across relaunch. All app surfaces, package states, dialogs, focus indicators, and scrollbars use semantic tokens in both appearances.

## Acceptance

- Installed, Discover, and Updates retain current behavior, package data, actions, navigation, and virtualized scrolling.
- Greek Sci-Fi identity is visible through architectural proportion, celestial-mechanism precision, and coherent craft details; a Greek icon or palette swap alone does not pass.
- Light preview shows one complete sunlit marble theme; Dark preview shows one complete deep-space theme. No mixed light/dark panels within either appearance.
- Light and Dark each have distinct, colorful palettes, readable text, visible focus, and distinguishable install, update, warning, selection, and destructive states.
- Luxury reads through material, spacing, type, and finish without excessive gold, ornament, gradient, or glow.
- System follows macOS appearance changes; explicit Light and Dark override System and survive relaunch.
- macOS shows no visible system title-bar title; native Traffic Light controls remain visible and functional, and the custom top rail drags the window.
- Main window and package list remain usable at RFC 0008's 1160×760 and 900×650 sizes, with no window-level scrollbar.
- README screenshot is recaptured from the implemented Buzz.app and shows genuine application UI; no mockup is presented as a product screenshot.
- Existing RFC 0008 management and accessibility acceptance criteria continue to pass.
