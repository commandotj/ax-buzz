# RFC 0001: Buzz Installer (Umbrella)

**Status:** Draft

**Type:** Umbrella

## Summary

Define Buzz Installer as a macOS software discovery and management product with a CLI and a Wails 3 desktop app. The CLI installs packages from Luban's Homebrew tap and other configured taps, reports versions, and supports automation. The desktop app discovers packages across those sources, manages installed software, monitors updates, and can update software.

## Problem

Package discovery, installation, update policy, and installed-software inventory share package identity and source metadata. They remain separate concerns, but need one product boundary so CLI and desktop behavior agree.

## Goals

The Buzz CLI and Wails 3 app have documented responsibilities; package source identity is consistent across them; each child RFC is approved and implemented before this umbrella closes.

## Non-goals

Non-Homebrew package ecosystems, remote account services, and changes to Homebrew itself. New concerns require a new RFC or umbrella.

## Children

| RFC | Concern |
|-----|---------|
| [0009](0009-sci-fi-greek-visual-system.md) | Buzz Desktop Sci-Fi Greek Visual System |
| [0008](0008-desktop-software-ui.md) | Buzz Desktop Software Management UI |
| [0007](completed/0007-go-function-modules.md) | Buzz Go Function Modules |
| [0006](completed/0006-streaming-homebrew-cli.md) | Buzz CLI Streaming Homebrew Operations |
| [0007](completed/0007-go-function-modules.md) | Buzz Go Function Modules |
| [0008](0008-desktop-software-ui.md) | Buzz Desktop Software Management UI |
| [0005](0005-buzz-cli-desktop-shell.md) | Buzz CLI and Desktop Shell Integration |
| [0004](0004-buzz-installed-software-management.md) | Buzz installed software management |
| [0003](0003-buzz-update-management.md) | Buzz update monitoring and package version reporting |
| [0002](0002-multi-source-install.md) | Multi-source Homebrew packages and Buzz CLI installation |

## Acceptance

Close when all listed children are Approved or Implemented and shared CLI/app responsibilities are consistent. Do not reopen for later features.
