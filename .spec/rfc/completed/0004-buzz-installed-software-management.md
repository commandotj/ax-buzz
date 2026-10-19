# RFC 0004: Homebrew formula and cask management (child of 0001)

**Status:** Implemented

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Define how Buzz's Wails 3 app displays and manages Homebrew formulae and casks.

## Problem

Users need one accurate view of installed and available Homebrew formulae and casks, with supported package actions and source metadata. Buzz must not imply ownership of software installed outside Homebrew.

## Goals

The app lists installed and available Homebrew formulae and casks, shows package type, tap, installed and available versions, and exposes only supported Homebrew actions. Install, upgrade, and uninstall results update the displayed Homebrew inventory and report failures.

## Non-goals

Non-Homebrew package managers, standalone macOS applications, multi-tap catalog policy (RFC 0002), and update scheduling or automatic update policy (RFC 0003).

## Design

Represent Homebrew formulae and casks with explicit type, tap, installed state, installed and available versions, and supported actions. Read installed state and package metadata through the existing Homebrew module and CLI protocol. Perform install, upgrade, and uninstall through Homebrew operations. Do not scan, display, or modify standalone applications or software owned by other package managers.

## Acceptance

Acceptance confirms installed and available formulae/casks show correct type, tap, versions, and supported actions; Homebrew install, upgrade, and uninstall update inventory; and failed Homebrew actions surface errors. Formula/cask classification and receipt parsing use fixture-backed tests. No acceptance criterion covers generic macOS application inventory.
