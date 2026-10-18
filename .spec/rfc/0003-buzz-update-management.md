# RFC 0003: Buzz update monitoring and package version reporting (child of 0001)

**Status:** Draft

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Define version reporting in Buzz CLI and update monitoring/automatic updates in the Wails 3 desktop app. Keep automatic update behavior app-owned; CLI reports installed and available versions without silently updating packages.

## Problem

Users need update status and a controlled update path. CLI automation must remain predictable, while the desktop app can monitor and apply updates under user-configured policy.

## Goals

Buzz CLI reports installed version, available version, and update status. Wails 3 app periodically checks installed managed packages and Buzz itself, shows pending updates, and supports user-enabled automatic updates with visible progress and failure results.

## Non-goals

Source discovery and tap installation (RFC 0002), installed software inventory and uninstall operations (RFC 0004), silent CLI updates, and forced updates without user policy.

## Design

Use Homebrew as version and update authority for Homebrew-managed software. Keep version comparison and update-result data shared between CLI and app. Buzz CLI exposes a read-only version/status command. Wails 3 app owns scheduling, preferences, notifications, and update execution; it must not run updates until enabled by the user. Check and update Buzz app/CLI through their declared Homebrew artifacts. Report unsupported or unmanaged software without claiming it can update it.

## Acceptance

CLI output distinguishes installed and available versions and exits predictably when update data is unavailable. App tests cover disabled-by-default automatic updates, scheduled checks, user-triggered updates, and failed updates. Manual acceptance confirms visible progress and that CLI status checks never mutate installed packages.
