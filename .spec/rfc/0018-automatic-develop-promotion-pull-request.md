# RFC 0018: Automatically open feature to develop pull requests

**Status:** Approved

## Summary

Run full CI for pushes to same-repository `feature/*` branches. When it passes, automatically open or reuse a pull request from that branch to `develop`. The merge remains protected by the existing required checks and human review rules.

## Problem

The current workflow validates feature branches only after a pull request exists. A pushed feature branch gets no CI result and requires a maintainer to create the integration PR manually, so the requested feature-to-develop handoff is not automatic.

## Goals

- Run the existing full `Verify` checks on pushes to same-repository `feature/*` branches.
- After successful verification, open or reuse exactly one PR from that branch to `develop`.
- Ensure the generated PR starts normal `pull_request` validation and uses the existing `Development source (develop)` guard.
- Keep merges manual and subject to `develop` branch rules.
- Never expose a write token to feature branch code.

## Non-goals

Creating feature branches, merging PRs automatically, promoting `develop` to `main`, generating release versions, or changing protected branch rules.

## Design

Extend `.github/workflows/ci.yml` to validate `push` events on `feature/*`. Keep the verification job read-only. Add a separate privileged workflow listening to successful `workflow_run` completions for the named CI workflow. This listener must run from the default branch, check the completed run's event, repository, branch prefix, and conclusion, and use run metadata only; it must not check out or execute feature branch content or download its artifacts.

GitHub only runs `workflow_run` listeners present on the repository's default branch. The first integration PR carrying this workflow must therefore be opened once by a maintainer; automation applies to later feature pushes after the workflow reaches `main`.

Use a fine-grained PAT stored as `FEATURE_PR_TOKEN` with only `Pull requests: write` on `commandotj/ax-buzz`. The trusted workflow-run listener finds an existing open PR by exact head branch and base `develop`; it creates one only when absent. Pass the token only to the GitHub CLI call. A PAT is required because a PR created with `GITHUB_TOKEN` does not reliably start the normal PR workflow. See [GitHub workflow token event behavior](https://docs.github.com/en/actions/concepts/security/github_token).

## Acceptance

- A feature push runs full `Verify`; unrelated branch pushes do not.
- Failed or cancelled verification creates no PR.
- Successful same-repository feature verification opens one `feature/*` → `develop` PR; later runs reuse it.
- A fork run, non-feature branch, wrong workflow, wrong event, or foreign repository cannot create a PR.
- The token is absent from the CI job and never reaches feature-controlled code.
- The created PR runs `Development source (develop)` and `Verify`; no workflow merges it.
