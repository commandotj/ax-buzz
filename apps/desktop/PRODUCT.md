# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Inferred from the user's prior explicit product direction: a Mac owner who uses Homebrew and wants one place to inspect and manage installed and available formulae and casks.

## Product Purpose

Buzz.app shows local Homebrew software state and lets the user install, remove, and upgrade supported packages. Success means the user can see what is installed, discover what is available, and understand package actions before running them.

## Positioning

Buzz.app presents package data and actions through a desktop interface backed by the Buzz Go CLI, which remains the package-operation owner. Do not duplicate Homebrew behavior in the React frontend.

## Operating Context

The app runs locally on macOS inside a Wails 3 shell. It bundles and invokes a matching Buzz CLI. Homebrew is the package source and authority. Cached catalog data should appear immediately; the user explicitly requests a refresh.

## Capabilities and Constraints

- Scope covers Homebrew formulae and casks; do not expand to general Mac application inventory.
- Show installed and available package state; expose install, remove, and upgrade only where supported.
- Keep the existing CLI and TUI available for terminal users.
- Frontend stack is React, TypeScript, and shadcn/ui.
- Treat cache age and refresh state as visible product information.
- Automatic background updates and non-Homebrew package managers are outside this surface.

## Brand Commitments

Product name is Buzz. No visual identity, palette, or typography has been confirmed. Existing handshake screen is a placeholder and not a design authority.

## Evidence on Hand

- `README.md` documents the CLI, desktop shell, Homebrew formulae/casks, and compatibility entry points.
- `screenshot.png` shows the inherited terminal package table and package details; it is content evidence, not a desktop visual reference.
- `apps/desktop/build/appicon.png` is an existing app asset; whether it defines desktop brand direction remains open.
- No customer, usage, or performance claims are available. Do not invent them.

## Product Principles

- Show real package state before promotion or decoration.
- Load cached data first and let the user explicitly refresh it.
- Make install, remove, and upgrade actions clear and report their results.
- Preserve clear ownership: React presents; Buzz CLI performs package operations.
