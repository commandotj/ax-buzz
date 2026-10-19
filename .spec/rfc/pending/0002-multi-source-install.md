# RFC 0002: Multi-source Homebrew packages and Buzz CLI installation (child of 0001)

**Status:** Pending

**Parent:** [0001](../completed/0001-buzz-installer.md)

## Summary

Add multi-tap package discovery and installation to Buzz CLI, make the Luban tap the default source, and publish Buzz CLI through that tap so users can install Buzz with Homebrew.

## Problem

The current project is a Homebrew TUI and has no source-selection contract or Buzz distribution tap. Users need to install Buzz and packages from Luban's catalog while retaining access to other taps.

## Goals

The Buzz CLI can list/search packages from the Luban tap and other configured Homebrew taps, install a package from an explicit source, and report source in results. The Luban tap contains a maintainable formula for Buzz CLI. Homebrew remains the installer and package database.

## Non-goals

Desktop UI and auto-update policy (RFC 0003), installed-software inventory UI (RFC 0004), and non-Homebrew sources.

## Design

Represent each package as tap-qualified identity plus formula/cask type; never collapse same-named packages from different taps. Delegate tap setup, package installation, and removal to the installed `brew` executable. Luban's tap is the default discovery source; other installed taps are searchable. Require an explicit qualified name when a package name is ambiguous. Ship a formula for Buzz CLI in the Luban Homebrew tap. Preserve existing CLI commands and behavior where source is unambiguous.

## Acceptance

Tests cover source-qualified identity, duplicate names, default Luban source, and command argument construction. A local tap fixture proves search and install selection without changing a user's real Homebrew state. Document tap setup and Homebrew installation of Buzz CLI. Existing single-source commands continue to work.
