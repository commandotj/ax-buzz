# RFC 0008: Buzz Desktop Software Management UI (child of 0001)

**Status:** Approved

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Define Buzz.app's software management screens and interactions using the existing Wails 3, React, TypeScript, and shadcn/ui stack. The UI lets users inspect installed and available Homebrew software, then install, remove, and upgrade supported packages. This RFC defines presentation and user interaction; backend command behavior remains owned by RFCs 0004 and 0006.

## Problem

The current React screen only checks whether Buzz CLI responds to a handshake. It has no installed inventory, available package catalog, cache status, manual refresh, or package actions. RFC 0004 defines inventory capabilities and RFC 0006 defines the CLI stream, but neither specifies how the desktop presents those capabilities. The current decorative overview leaves the primary job unreachable.

## Goals

Replace the handshake landing page with one software management workspace. Provide distinct Installed and Discover views, with an Updates filter/view for installed packages that have upgrades. Show package name, installed/available version, source/type, and only actions supported by the item. Load cached catalog data at startup, display its age, and let the user explicitly refresh it. Keep the selected view, search, and list position stable during data updates and row actions.

## Non-goals

- New Homebrew operations, JSONL event semantics, and cache storage rules (RFC 0006).
- Inventory classification and whether an item supports removal or upgrade (RFC 0004).
- Multi-tap discovery policy and source identity (RFC 0002).
- Preferences, automatic update scheduling, notifications, or background update execution (RFC 0003).
- Branding asset creation or a general design-system rewrite.

## Design

### Visual direction: Night City Software Index

Use the neon-noir atmosphere of Cyberpunk 2077's Night City as visual reference, while keeping Buzz a legible, data-first macOS utility. Use deep charcoal surfaces, sharp cyan and signal-yellow accents, restrained magenta for selected high-salience states, compact technical typography, crisp rules, and small diagnostic labels. Keep glow and gradients limited to focus and active-state edges. Do not reproduce game screens, logos, fictional brands, or decorative city art. Package data remains the visual focus; no promotional hero.

Use system sans-serif for interface text and system monospace for versions, package identifiers, and source labels. Align versions and counts with tabular numerals. Status colors also need text or icon labels; color alone never communicates state. Define semantic theme tokens, not component-level literals. Light appearance keeps a pale neutral canvas with the same cyan/yellow accents; dark appearance uses a near-black canvas, graphite surfaces, electric cyan accent, and signal-yellow warning. Verify text, controls, focus, and status contrast in both appearances, adjusting token values while preserving these roles.

### Window and navigation

Build the main window as a persistent left navigation rail and a content pane. Title the window `Buzz`. Navigation has `Installed`, `Discover`, and `Updates`; selecting a navigation item changes content without scrolling the window to the top as a side effect of pointer hover. The content header contains the view title, package count, search field, and a `Refresh` action. Use shadcn/ui primitives already installed where suitable; keep package rows information-dense and avoid a large promotional hero. The window and document never scroll; only the package list and selected-package inspector may scroll.

At the normal desktop window size, show a package list beside a selected-package inspector. The inspector shows the selected package's description, installed and available versions, source/type, and supported actions. Keep list scrolling in the content pane. At narrower supported widths, let the inspector collapse into a dismissible detail panel so the list remains usable; navigation selection must not alter scroll position. Initial window size is 1160×760; verify layout at 900×650 and the normal size.

`Installed` lists locally installed Homebrew formulae and casks. Each row shows name, installed version, package type/source, and available actions. `Remove` requires an explicit confirmation naming the package. `Updates` lists installed packages with a newer version and provides per-package upgrade plus an explicit `Upgrade all` action. `Discover` lists available formulae and casks, indicates installed state, and offers `Install` for uninstalled supported items. Installed items in Discover are informational and do not expose Remove. Use TanStack Virtual's React adapter (`@tanstack/react-virtual`) to virtualize the package list: mount only visible rows plus a small overscan, while retaining the full filtered catalog and native list scrolling. Unsupported/unmanaged items remain visible only where inventory data provides them and have no destructive action.

On launch, render cached catalog and inventory data immediately when available. Show last refresh time and a stale indicator when the CLI reports stale cache. Refresh is user-triggered; it shows inline progress while retaining the current list and selection. Empty cache shows a clear first-load state and starts the catalog fetch. Loading, empty, command failure, cancellation, and action success each have explicit visible states. A successful install/remove/upgrade refreshes the affected inventory/catalog state while preserving the active view, search query, scroll offset, and selected package when that package remains in the list.

Rows expose actions on keyboard focus as well as pointer interaction. Confirmation dialogs and buttons have accessible names; progress is announced without stealing focus. The content pane owns list scrolling; navigation hover never changes scroll position. Use the generated Wails bindings to invoke backend operations and consume the existing JSONL event stream; do not duplicate Homebrew logic in React.

### Appearance and theme

Provide an Appearance control in app settings with `System`, `Light`, and `Dark` choices. `System` follows the current macOS appearance and reacts when macOS changes between light and dark while Buzz is open. `Light` and `Dark` override the system appearance. Persist the user's choice across relaunch; missing or invalid stored values resolve to `System`.

The Night City theme is the initial and only shipped visual theme in this RFC. Its semantic tokens must have light and dark values, and every surface, text, border, focus, selection, and status treatment must consume those tokens. “Theme support” here means this tokenized theme works in all three appearance modes; it does not add a theme marketplace or multiple selectable visual skins.

## Acceptance

With fixture-backed CLI events, the app renders Installed, Discover, and Updates with correct package counts, versions, source/type, installed state, and action availability. Cache data appears before refresh; cache age/stale state and explicit refresh are visible. Refresh progress, empty catalog, CLI failure, cancellation, and successful action outcomes render correctly. Remove requires confirmation; install/remove/upgrade invoke the corresponding backend operation. After an action, current view, search, scroll position, and selection remain stable when the selected package still exists. Keyboard-only navigation can reach package actions and dialogs, labels are accessible, and pointer hover over navigation does not jump the list to the top.

At 1160×760 and 900×650, the index list and navigation remain usable; the selected-package inspector adapts to the narrower layout. The document and main window have no outer scrollbar; only package list and inspector content scroll. Large Discover catalogs use `@tanstack/react-virtual`; rendered package rows remain bounded to visible rows plus overscan while scrolling the complete filtered result set. Light and Dark render the Night City semantic theme with readable text, visible focus, and distinguishable statuses. System follows macOS appearance changes during a session. Explicit Light/Dark choices override macOS appearance and survive relaunch; unset or invalid preferences select System. Verify the main window at both specified sizes.
