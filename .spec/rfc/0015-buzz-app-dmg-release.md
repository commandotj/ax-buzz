# RFC 0015: Buzz.app DMG GitHub Release (child of 0012)

**Status:** Approved

**Parent:** [0012](0012-buzz-ci-cd.md)

## Summary

Release Buzz as a macOS `.dmg` containing `Buzz.app` and its bundled Buzz CLI. Use Wails 3's native DMG packager and publish app DMGs only; do not publish standalone CLI archives or binaries.

## Problem

Current CI builds the `.app` but produces no installable disk image. The previous release plan targeted standalone CLI tarballs, which are the wrong product because the CLI ships inside Buzz.app. The repository's pinned Wails beta supports DMG creation, but the project's custom Darwin Taskfile has no DMG task.

## Goals

Add a reproducible Wails task for creating a DMG from the packaged `.app`. For version tags, build and validate separate `arm64` and `amd64` DMGs, then publish those DMGs as the only GitHub Release assets. Keep CLI build and test targets for the bundled component, without standalone CLI release artifacts.

## Non-goals

Standalone CLI releases, Windows/Linux distribution, Homebrew tap updates, automatic version bumps, and application auto-updates.

## Design

Extend `apps/desktop/build/darwin/Taskfile.yml` with Wails-native DMG creation tasks, following the task model shipped in pinned Wails `v3.0.0-beta.30`. Build each `.app` natively on its matching macOS runner so the bundled `Buzz.app/Contents/Resources/buzz` CLI matches the app architecture. Call Wails `tool package --format dmg` with the existing `icons.icns` volume icon. Each DMG contains `Buzz.app` and an Applications shortcut.

Remove manual standalone CLI archive and GitHub release targets from the root `Makefile`. Keep local CLI build, test, vet, and format targets because the CLI remains a bundled application component.

Replace the CLI-only release workflow with a tag-triggered desktop release. Build architecture-specific DMGs on matching `macos-15` (`arm64`) and `macos-15-intel` (`amd64`) runners. Validate the embedded `CFBundleShortVersionString` against the tag, verify each DMG and its contents, and publish only the DMGs. Keep release write permission in the publishing job.

The trusted tag workflow creates an annotated version tag after main CI succeeds. Configure repository-local Git `user.name` and `user.email` as `github-actions[bot]` before creating that tag; the release token authenticates the push but does not provide Git committer identity.

Local app packaging uses ad-hoc signing. The tag release workflow uses the protected `release` environment for Apple signing and notarization secrets. A public macOS release requires Developer ID signing and notarization. Map required secrets into Wails signing and Apple notarization setup, validate the Developer ID identity and notary credentials, sign and notarize the app, create and staple the DMG, then publish it. Fail closed when credentials or verification are missing.

## Acceptance

Wails task builds the app with matching bundled CLI and creates a valid DMG for `arm64` and `amd64`. Mounted images contain `Buzz.app` and Applications shortcut; app bundle version matches the pushed tag. Gatekeeper accepts the signed, notarized app and DMGs. GitHub Release contains only the validated DMG assets. Missing signing/notarization credentials or failed verification blocks publishing. The root `Makefile` exposes no standalone CLI archive or release target.

After successful main CI, the tag workflow creates and pushes an annotated version tag on the verified main commit without relying on runner-global Git identity; that tag starts the signed and notarized DMG release workflow.
