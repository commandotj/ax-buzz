# RFC 0010: Single Buzz Desktop Instance

**Status:** Implemented

## Summary

Buzz desktop runs as one process per user session. Launching Buzz again exits the duplicate process and activates the existing main window.

## Problem

Opening Buzz repeatedly can start multiple desktop processes and windows, causing duplicate Homebrew operations and confusing app state.

## Goals

- Enforce one running Buzz desktop instance using Wails single-instance support.
- Restore and focus the existing main window when another launch is attempted.
- Use Buzz's stable bundle identifier as the instance identity across packaged and development launches.

## Non-goals

- Multiple windows or per-document window routing.
- Forwarding CLI arguments or deep links to the running window.
- Single-instance behavior for the standalone `buzz` CLI.

## Design

Configure `application.Options.SingleInstance` in `apps/desktop/main.go` with unique ID `com.commandotj.buzz`, matching the macOS bundle identifier. In `OnSecondInstanceLaunch`, restore and focus the already-created main `WebviewWindow`. Keep callback safe before window initialization by checking the window pointer. Use Wails' built-in lock and IPC implementation; do not add a second lock mechanism or dependency.

## Acceptance

- Starting Buzz twice leaves exactly one Buzz process and one main window.
- The second launch brings the first window to the foreground, including when it was minimized.
- Closing the first process allows Buzz to launch normally again.
- `pnpm build` packages the desktop application with single-instance behavior enabled.
