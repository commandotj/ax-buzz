# RFC 0006: Buzz CLI Streaming Homebrew Operations (child of 0001)

**Status:** Implemented

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Give Buzz CLI a headless Homebrew data and action interface for Buzz.app. `buzz --stream-json` emits versioned JSON Lines for cached and refreshed package catalogs, installed package state, install, uninstall, and upgrade operations. Keep the desktop process out of the terminal UI path.

## Problem

RFC 0005 establishes a machine protocol but implements only a handshake. Buzz.app cannot show available or installed Homebrew packages or invoke package actions through the CLI. The app needs a stable stream contract that exposes existing Homebrew behavior without starting Bubble Tea.

## Goals

The CLI provides machine commands to list available packages with installed state, list installed packages, refresh Homebrew catalog data on demand, and install, uninstall, or upgrade a selected package. Catalog reads use local cache when present; an explicit refresh bypasses it. Machine mode emits progress, data, and one terminal result event as JSON Lines, writes diagnostics to stderr, and never starts Bubble Tea. Buzz.app accepts and forwards all three event types. Bundle the matching CLI helper inside Buzz.app and resolve it from the app bundle. Direct terminal invocation retains the existing TUI and the `taproom` compatibility entry point.

## Non-goals

- Buzz.app React UI layout and package presentation.
- Multi-tap source policy, Buzz formula distribution, scheduled checks, and automatic updates.
- Non-Homebrew package managers and unmanaged application removal.
- Removing the existing terminal TUI.
- Moving Go packages into independent modules; RFC 0007 owns module boundaries and migration.
- Desktop UI layout and interactions; RFC 0008 owns those decisions.

## Design

Add `--stream-json` as the app-facing mode. It accepts explicit `list`, `installed`, `refresh`, `install`, `uninstall`, and `upgrade` operations. `list` reads the formula/cask cache and combines it with a fresh local installed-state scan; it reports cache age so the app can label stale data. If no cache exists, it fetches catalog data. `refresh` explicitly bypasses the catalog cache and returns refreshed data. Read operations never invoke `brew update`; updates happen only through an explicit package upgrade operation. Existing cache paths and TUI behavior remain compatible.

Each output line is one versioned JSON event: `progress`, zero or more `data` records, then exactly one terminal `result`. The terminal result carries success or failure status; process exit code remains authoritative. Standard output contains JSON only, diagnostics go to standard error, and command arguments are passed without shell interpolation. Cancellation terminates the running Homebrew process and returns a failure result. Desktop protocol validation accepts `progress`, `data`, and `result`, forwards progress/data to the frontend, and treats result as terminal. Retain `--machine handshake` as a compatibility alias while Buzz.app adopts `--stream-json`.

Resolve Homebrew from the process `PATH` first, then standard installation paths, so launching Buzz.app from Finder does not depend on a shell-configured `PATH`.

The macOS bundle contains the matching CLI executable at `Buzz.app/Contents/Resources/buzz`. The Wails packaging task copies it there before signing. Desktop resolves this path relative to its own executable and never searches the GUI process's `PATH`.

## Acceptance

`apps/cli` implements `--stream-json` operations `list`, `installed`, `refresh`, `install`, `uninstall`, and `upgrade`. `list` returns formulae and casks with stable identity, installed state, and cache age; repeated reads use cache. `refresh` bypasses cache; absent-cache reads fetch catalog data. Read operations never run `brew update`. Actions return streamed progress and one terminal result with correct process exit status. Standard output contains JSON events only; diagnostics go to stderr. Cancellation stops the active Homebrew process. Malformed arguments fail without entering the TUI. Existing interactive Buzz and `taproom` invocations remain compatible.

`apps/desktop` validates and forwards `progress`, `data`, and terminal `result` events. Packaged `Buzz.app` contains the matching CLI at `Contents/Resources/buzz`, copied before signing and resolved relative to the app executable without GUI `PATH` lookup. RFC 0007's workspace build can compile this CLI and desktop Go app; module extraction details are verified by RFC 0007.
