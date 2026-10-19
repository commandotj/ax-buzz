# RFC 0016: Buzz feature, develop, and main branch flow

**Status:** Approved

## Summary

Define Buzz's branch flow: feature branches merge into `develop`, then `develop` promotes to `main` through a pull request. GitHub Actions must reject other sources for production pull requests and run full validation at both integration gates.

## Problem

The repository has only `main`, no `develop` branch, and no rulesets. RFC 0013 currently checks only pull requests targeting `main`; it does not define feature integration, promotion source enforcement, or protected branch rules. A base-branch filter alone cannot stop a feature branch from opening a pull request into `main`.

## Goals

- Establish `feature/*` → `develop` → `main` as the only supported pull request path.
- Run full CI for pull requests targeting `develop` and `main`, with distinct source-guard check names for each target branch.
- Require production pull requests to come from `develop` in this repository; reject fork branches named `develop` and all other sources.
- Protect `develop` and `main` with pull requests, required checks, and force-push/deletion restrictions.
- Keep release tags anchored to commits already merged into `main`.

## Non-goals

Creating feature branches for individual product changes, changing Buzz versioning rules, changing DMG signing or notarization (RFC 0015), and publishing CLI artifacts.

## Design

Update the CI workflow so pull requests targeting `develop` or `main` run formatting, Go test/vet, and desktop build verification. Keep the full-validation check named `Verify` on both base branches. Give source guards target-specific names (`Development source (develop)` and `Promotion source (main)`) so GitHub does not merge their check contexts when one commit is used by pull requests to both branches.

For pull requests targeting `develop`, add a required `Development source (develop)` check that passes for same-repository `feature/*` branches. RFC 0019 adds one narrow exception for the same-repository `release-please--branches--develop` branch: its diff may change only release metadata, and its synchronized version must exceed the PR base. For pull requests targeting `main`, add a required `Promotion source (main)` check that passes only when the head branch is exactly `develop` and the head repository is `commandotj/ax-buzz`. Run full CI only after the matching source guard passes. Direct pushes to `main` still run CI, but do not bypass the repository's pull request rules.

Bootstrap remote `develop` from the current `main` tip, then add repository rulesets for `develop` and `main`. Both rulesets require pull requests, require `Verify` plus their matching source-guard check, reject force pushes and deletion, and have no bypass actors. Do not invent check contexts: run the workflows first, inspect GitHub's reported check names, then configure the exact names and reread effective rules.

The release workflow remains a separate trusted tag workflow from RFC 0015. It must verify the tagged commit is reachable from `main`; pull request workflows receive no release permissions or signing secrets.

## Acceptance

Same-repository `feature/*` pull requests to `develop` run full CI. A same-repository `develop` promotion pull request to `main` passes the source guard and full CI. A non-feature source or fork pull request to `develop`, a feature-branch pull request to `main`, and a fork branch named `develop` all fail the source guard. GitHub rulesets require `Verify` and the exact target-specific source check on each branch, reject force-push/deletion, and have no bypass actors. Release tags pointing outside `main` cannot publish. GitHub rules and workflow runs are reread and reported separately from local validation.
