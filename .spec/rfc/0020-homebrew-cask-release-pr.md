# RFC 0020: Automated Buzz.app Homebrew Cask pull request

**Status:** Approved

## Summary

After a Buzz.app release has been published with both verified DMGs and checksums, automatically open or update a PR to `develop` in `systembugtj/luban-homebrew` for its `buzz` Cask. The Cask installs `Buzz.app`; it does not install or publish a standalone CLI.

## Problem

The earlier Buzz Cask seed used placeholder asset URLs and hashes and was removed. Buzz has no published Homebrew install path. The release pipeline now needs to submit a real Cask change only after public release assets and their checksums exist.

## Goals

- Use the signed, notarized arm64 and amd64 DMGs from the exact successful Buzz release.
- Generate verified `url` and `sha256` entries from published release assets; no placeholder values.
- Open or update one release-specific PR to the existing tap `develop` branch.
- Create the Cask when absent and update it on later releases.
- Keep Buzz CLI bundled inside `Buzz.app`; do not create a Formula or CLI artifact.

## Non-goals

Publishing directly to either tap branch, updating `main`, standalone CLI packaging, changing Buzz release signing, or installing Buzz through the app's own UI.

## Design

Add a post-publish job to `.github/workflows/release-app.yml` that depends on the DMG release and asset verification job. Fetch release assets from GitHub, verify the asset names and checksums against the just-published release, then use a fine-grained token from secret `HOMEBREW_CASK_TOKEN` to create a release-specific branch and PR in `systembugtj/luban-homebrew` with base `develop`. The token is limited to the tap repository and Contents plus Pull requests write; it is not exposed to Buzz pull-request jobs.

Generate `Casks/buzz.rb` with the actual release version, both architecture URLs and SHA-256 values, app name `Buzz`, bundle identifier, and supported macOS metadata from the existing tap conventions. If an open PR for this Buzz version exists, update that PR branch; otherwise create it. Never push to tap `develop` or silently fall back to tap `main`. A tap API or token failure marks only the Cask job failed and reports the published release link and actionable error; it cannot roll back the release.

## Acceptance

- The job does not run until the release job has succeeded and both DMGs are publicly downloadable.
- Cask version, URLs, and hashes match the release exactly; no placeholder value reaches the PR.
- The PR targets `systembugtj/luban-homebrew:develop`, has one release-specific branch, and is idempotently updated.
- The Cask points to Buzz.app and has no standalone CLI formula.
- Missing asset, checksum mismatch, wrong tap target, or missing token fails closed without changing either tap branch.
