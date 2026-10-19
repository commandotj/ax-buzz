# RFC 0019: Release Please version pull request automation

**Status:** Approved

## Summary

Use Release Please to turn Conventional Commits merged into `develop` into one version and changelog PR targeting `develop`. Version the bundled Buzz CLI and Buzz.app together. Do not create production tags or releases on `develop`; after the version PR is promoted, build and publish signed DMGs from the exact `main` commit, then create the matching `vX.Y.Z` tag and GitHub Release.

## Problem

The repository has no remote version tags or published GitHub Release, despite source metadata already being `0.6.2`. Today a maintainer must edit version files, create a tag, and start the DMG workflow manually. The existing tag-triggered workflow also has no automated step that turns a merged `develop` version into a production release.

## Goals

- Use Conventional Commits to select patch/minor/major version bumps; no bump for docs-only or hidden commit types.
- Seed Release Please from the existing `0.6.2` source version and the existing `v0.6.2` commit, without claiming a release already exists.
- Generate one version PR on `develop`; update the CLI embedded version, Wails config version, Info.plist version fields, manifest, and changelog together.
- Permit Release Please's exact `release-please--branches--develop` PR branch through the develop source guard only when it comes from this repository and changes release metadata only.
- Keep tags and GitHub Releases off `develop`.
- Open the `develop` → `main` PR only after the merged Release Please PR marks a pending version.
- On merge to `main`, build, sign, notarize, and verify arm64 and amd64 DMGs from that exact commit; then create the matching tag and GitHub Release with only those DMGs and checksums.
- Keep release publishing idempotent and make partial asset publication recoverable without changing the tagged source commit.

## Non-goals

Standalone CLI releases, non-macOS app distribution, automatic PR merging, and Homebrew tap changes (RFC 0020).

## Design

Add manifest-mode Release Please configuration and a workflow on `push` to `develop`. Pin `googleapis/release-please-action` to a reviewed release SHA. Use tag format `vX.Y.Z`, one manifest package at repository root, `include-v-in-tag: true`, `include-component-in-tag: false`, `skip-github-release: true`, and `target-branch: develop`. Set the manifest baseline to the checked-in source version `0.6.2`; set `bootstrap-sha` to `f43ac21dea20a3db6871aabeecaca2b3de225aba`, the commit referenced by the local `v0.6.2` tag, so older Taproom history is excluded. This is an unpublished source baseline, not a claim that v0.6.2 is released. Remove `bootstrap-sha` after the first Release Please PR is merged.

Mark `apps/cli/.version` with Release Please's generic updater marker and keep its first token as `vX.Y.Z` so existing Go version output remains compatible. Update the Go embed consumer to read the first token. Use Release Please's generic updater marker in `apps/desktop/build/config.yml` to change only `info.version` while preserving the Wails YAML formatting and comments; use the XML updater to change both `CFBundleShortVersionString` and `CFBundleVersion` in `apps/desktop/build/darwin/Info.plist`. Use a fine-grained PAT secret `RELEASE_BOT_TOKEN`; it needs Contents, Issues, and Pull requests write access because Release Please pushes its release branch and opens a PR. Never pass it to pull-request or feature-controlled jobs. Publish generated PR links and state in the Actions summary.

Configure merged version PRs with Release Please's `autorelease: pending` label. Promotion creation is RFC 0021 scope. After successful `main` Verify, a trusted workflow creates `vX.Y.Z` on that exact main SHA using `RELEASE_BOT_TOKEN`; the tag event starts RFC 0015's signed DMG workflow and satisfies the existing `release` environment's `v*` deployment policy. Only after both DMGs pass signing, notarization, and content validation does the workflow create/publish the GitHub Release. If release work fails after tag creation, manual recovery dispatches the release workflow against that existing tag and never moves it. Retag the merged Release Please PR as `autorelease: tagged` only after the release assets are verified, so Release Please can proceed to the next version.

The normal `Development source (develop)` check accepts same-repository `feature/*` branches plus Release Please's fixed `release-please--branches--develop` branch. For the Release Please branch, require its diff to contain only `.release-please-manifest.json`, `apps/cli/.version`, `apps/desktop/build/config.yml`, `apps/desktop/build/darwin/Info.plist`, and `CHANGELOG.md`; require synchronized versions and a version greater than the PR base. Reject forks and every other branch.

## Acceptance

- With pinned Release Please CLI/action, dry-run from bootstrap commit and prove version PR changes all configured version fields exactly once while preserving the CLI's `v` prefix.
- On `develop`, an eligible conventional commit opens/updates one version PR; a docs-only change creates none.
- The version PR targets `develop`, carries `autorelease: pending`, and has normal CI checks.
- The Release Please source guard rejects forked release branches, arbitrary source branches, non-release file changes, and a version that is not ahead of develop.
- Merging it creates no tag or release; promotion automation is separately verified by RFC 0021.
- A successful `main` Verify creates a `vX.Y.Z` tag on that exact commit only when the synchronized version is newer than the latest tag.
- A tag release validates all version fields, builds both supported macOS architectures, and publishes only signed/notarized DMGs plus checksums.
- Duplicate version/tag/release state cannot overwrite a published version; failed or partial publication can resume from the same source commit.
- After successful release verification, the merged version PR is relabeled `autorelease: tagged` and a later eligible `develop` commit can create the next version PR.
