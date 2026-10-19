# RFC 0005: Buzz CLI and Desktop Shell Integration (child of 0001)

**Status:** Implemented

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Turn the inherited Taproom Go TUI into Buzz CLI and add the Buzz.app desktop shell. Place executable products under `apps/cli` and `apps/desktop`; build the desktop app with Wails 3, React, and shadcn/ui. Buzz.app invokes a version-matched Buzz CLI process through a documented machine-readable protocol. Preserve existing Taproom entry points and user data paths during the rebrand.

## Problem

The repository currently contains a root-level Go TUI named Taproom. It has no non-interactive command contract for a desktop app, and no desktop frontend. The desktop app needs one supported path to perform package operations without duplicating Homebrew behavior or parsing terminal UI output. The current RFC set defines package features, but not this executable boundary or product layout.

## Goals

The Go CLI is the canonical owner of package commands and machine-readable results; its existing interactive TUI remains available. `Buzz.app` is a Wails 3 app with a React and shadcn/ui frontend, and runs its bundled, version-matched CLI executable with explicit arguments. The repository has `apps/cli` and `apps/desktop` product directories, native Go module metadata, and root pnpm/Turborepo tasks for cross-app development and builds. Existing Taproom commands, build entry points, log configuration, and cache data remain usable through compatibility behavior.

## Non-goals

Package-source discovery and install selection remain RFC 0002 scope. Version reporting, scheduling, and update policy remain RFC 0003 scope. Installed-software inventory and supported management actions remain RFC 0004 scope. Product branding assets, signing/notarization, release automation, and non-Homebrew ecosystems are out of scope.

## Design

Move the existing Go executable and its internal packages under `apps/cli`, update the Go module path and internal imports to the repository's Buzz CLI identity, and retain the root Makefile entry points as compatibility wrappers. Build the canonical executable as `buzz`; retain `taproom` as a compatibility executable for existing invocations. Keep the existing Taproom cache location. Support `BUZZ_LOG` as the preferred log setting while continuing to honor `TAPROOM_LOG` when the new setting is absent.

Create `apps/desktop` as the Wails 3 application. Use React with TypeScript and shadcn/ui in its frontend. The Wails Go backend owns subprocess lifecycle and relays CLI progress and results to the frontend; it does not implement package operations or invoke Homebrew directly. Bundle the matching `buzz` executable inside `Buzz.app` and invoke it by explicit bundle path, never by searching the GUI process's `PATH`.

Define a versioned JSON Lines protocol for app-driven CLI commands. Each event carries protocol version, event type, and data; progress events precede one terminal result event. Machine mode writes protocol events to stdout and diagnostics to stderr. Process exit status indicates success or failure. The app passes an argument vector without shell interpolation, supports cancellation, and reports malformed output, process startup failure, cancellation, and nonzero exit as operation errors. RFCs 0002–0004 define command-specific operations and must follow this shared protocol.

Add a root pnpm workspace and Turborepo tasks to coordinate frontend and Go product commands. Go remains managed by its native module and toolchain; package-level scripts only expose Go CLI operations to the root task graph. Preserve existing root `make` build/run entry points while adding canonical Buzz tasks. Do not add a separate shared runtime library in this RFC.

## Acceptance

Build the CLI and Buzz.app from a clean checkout using documented root commands. CLI acceptance confirms `buzz` starts the existing TUI, machine mode emits valid versioned JSON Lines with distinct progress and terminal result events, and failures have nonzero exit status with diagnostics on stderr. Compatibility acceptance confirms existing `make` entry points, `taproom` invocation, `TAPROOM_LOG`, and existing cache contents continue to work. App acceptance confirms Buzz.app invokes its bundled CLI independent of shell `PATH`, streams progress/results into the UI, cancels a running operation, and displays startup, protocol, cancellation, and command failures. Frontend acceptance confirms React and shadcn/ui render inside the Wails 3 shell. Root pnpm/Turborepo tasks invoke the native Go and Wails/frontend commands without duplicating their build logic.
