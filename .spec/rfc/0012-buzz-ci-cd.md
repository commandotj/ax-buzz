# RFC 0012: Buzz GitHub CI/CD (Umbrella)

**Status:** Draft

**Type:** Umbrella

## Summary

Add standard GitHub Actions automation for Buzz: validate pull requests and `main` updates, then publish the macOS Buzz.app DMG when a version tag is pushed. The CLI is bundled inside Buzz.app and is not a separate release product. Keep CI and publishing in separate workflows so untrusted code gets read-only permissions and only trusted tag releases can write GitHub Releases.

## Problem

Repository had no GitHub Actions workflows. Local Make targets define Go checks, and Wails builds Buzz.app, but there was no hosted verification or installable app release path. A CLI archive release would publish the wrong product because the CLI is bundled in Buzz.app.

## Goals

Complete when CI and Buzz.app DMG release children are implemented and their acceptance criteria pass. Preserve `main` as the integration branch and version tags as the app release trigger.

## Non-goals

Provisioning Apple Developer credentials, auto-update metadata, Homebrew tap updates, branch rulesets, automatic version changes, and additional branch promotion models. The DMG child defines the signing and notarization gate for releases.

## Children

| RFC | Concern |
|-----|---------|
| [0015](0015-buzz-app-dmg-release.md) | Buzz.app DMG GitHub Release |
| [0013](0013-github-actions-ci.md) | Buzz pull request and main CI |

## Acceptance

Close when RFC 0013 CI and RFC 0015 Buzz.app DMG release are implemented and verified. Later distribution channels or branch policy work require separate RFCs.
