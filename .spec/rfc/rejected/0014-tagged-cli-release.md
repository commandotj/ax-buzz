# RFC 0014: Tag-based Buzz CLI GitHub releases (child of 0012)

**Status:** Superseded

**Parent:** [0012](../0012-buzz-ci-cd.md)

Superseded by [RFC 0015](../0015-buzz-app-dmg-release.md): Buzz CLI is bundled in Buzz.app and is not released separately.

## Summary

Publish Buzz CLI GitHub Releases from version tags. Validate the tag against `apps/cli/.version`, reuse `make release` to build existing macOS ARM64 and Intel archives plus checksums, then attach those artifacts to a GitHub Release using a trusted, least-privilege workflow.

## Problem

The repository has no published GitHub Releases. `make release` already defines the archive contract, but `make gh-release` pushes a tag and opens an editor, so it cannot run unattended in GitHub Actions.

## Goals

On `v*` tag pushes, require the tag to match `apps/cli/.version`, build the existing CLI assets, create a GitHub Release, upload all archives and checksum file, and fail if any required asset is absent. Restrict `contents: write` to the release job.

## Non-goals

Buzz.app distribution, code signing/notarization, Homebrew tap updates, automatic version bumps or tag pushes, and changing the Makefile archive contract.

## Design

Add `.github/workflows/release-cli.yml`, triggered only by `v*` tags. Check out the exact tag commit, compare `github.ref_name` with `apps/cli/.version`, and require the tagged commit to be reachable from `main`. Run `make release`; this target builds `buzz` and legacy-compatible `taproom` archives for macOS `arm64` and `amd64` and writes `release/checksum.txt`.

Before publishing, verify the expected four `.tar.gz` archives and `checksum.txt` exist and are non-empty. Create the GitHub Release from the triggering tag and upload those files. Grant `contents: write` only to the publishing job; do not pass this permission to CI or pull request workflows. Serialize release runs to avoid concurrent updates to the same release. Do not call `make gh-release`: it creates and pushes tags and requires an interactive editor.

## Acceptance

Pushing a tag equal to `apps/cli/.version` whose commit is reachable from `main` builds and publishes the four expected archives and checksum file on that tag's GitHub Release. A mismatched tag, off-main commit, failed build, or missing asset fails before publishing. The workflow creates a draft first and publishes only after all assets upload. Workflow run has write permission only in the release job. Confirm asset names and published checksums against local build output. Buzz.app release remains explicitly unsupported by this workflow.
