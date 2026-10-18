# RFC 0004: Buzz installed software management (child of 0001)

**Status:** Draft

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Give Buzz's Wails 3 app an inventory of installed macOS apps and CLI packages and actions to manage software supported by its installation source.

## Problem

Buzz Installer needs to show available packages and manage installed software. Homebrew-only inventory misses apps and CLI tools installed through other supported macOS mechanisms.

## Goals

The app distinguishes installed apps and CLI tools, shows version and installation source when known, and supports install, update, and uninstall for Homebrew-managed items. It labels items whose source or management actions are unknown instead of guessing.

## Non-goals

Multi-tap catalog behavior (RFC 0002), update scheduling and automatic update policy (RFC 0003), and destructive removal of unmanaged software.

## Design

Build the desktop shell with Wails 3. Keep inventory records explicit about item type, source, installed version, and supported actions. Use Homebrew metadata for Homebrew-managed formulae and casks. Treat other macOS applications as inventory-only unless a reliable install source and supported operation are known. Require user action for uninstall; never remove unmanaged files or infer ownership from a display name.

## Acceptance

Fixture tests cover app/CLI classification, known and unknown source metadata, and supported-action decisions. App acceptance verifies Homebrew install/uninstall results refresh inventory, unknown apps remain visible without destructive actions, and failed actions surface errors.
