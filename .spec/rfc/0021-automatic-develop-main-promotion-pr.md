# RFC 0021: Automatically open promotion PR after release version merge

**Status:** Approved

## Summary

When a Release Please version PR is merged into `develop`, automatically open or update one `develop` → `main` promotion PR only when the merged version is ahead of `main`. Keep review and merge manual.

## Problem

The required branch policy permits only `develop` as the source for production PRs, but currently a maintainer must manually create that PR after the version PR merges. Opening a promotion on every `develop` push would be wrong: feature merges without a version bump are not a production release decision.

## Goals

- Trigger only when a merged pull request to `develop` carries Release Please label `autorelease: pending`.
- Verify latest `develop` version files are synchronized and the app version exceeds the version on `main`.
- Stop if another open Release Please version PR on `develop` is pending.
- Open or reuse exactly one same-repository PR with head `develop` and base `main`.
- Keep existing `Promotion source (main)` and `Verify` checks and manual merge rules effective.

## Non-goals

Creating feature-to-develop PRs (RFC 0018), generating version PRs (RFC 0019), merging PRs, creating tags, signing or publishing DMGs, and editing branch rulesets.

## Design

Add a privileged promotion workflow listening to `pull_request: closed` for base `develop`. Gate the job on `merged == true` and the exact `autorelease: pending` label. It checks out trusted `develop` with full history, checks the branch version against `main`, confirms there is no open version PR with that label, and queries for an existing open PR with head `develop` and base `main` before creating one. Use `RELEASE_BOT_TOKEN` only in this trusted merged-PR workflow so the generated PR starts normal required CI; do not use it in feature PR jobs. The token needs Pull requests write. The PR body records the release version and points to the version PR. No PR is opened when versions are equal or checks cannot prove the version is ahead.

## Acceptance

- A merged pending Release Please PR with a greater synchronized version opens one `develop` → `main` PR.
- Repeated event delivery reuses the existing PR; no duplicate appears.
- Feature PR merges, docs-only changes, unrelated closed PRs, unmerged PRs, and equal/invalid versions create no promotion PR.
- The resulting PR passes `Promotion source (main)` and `Verify` without manual Actions approval.
- No merge or production release is automated by this workflow.
