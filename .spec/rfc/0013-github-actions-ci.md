# RFC 0013: Buzz pull request and branch CI (child of 0012)

**Status:** Approved

**Parent:** [0012](0012-buzz-ci-cd.md)

## Summary

Add read-only GitHub Actions checks for pull requests and pushes targeting `develop` or `main`. Checks validate Go formatting, enforce 100% statement coverage across Buzz product Go packages, enforce 100% line, statement, function, and branch coverage for first-party TypeScript, vet Go code, then build Buzz.app on macOS.

## Problem

There is no hosted check on proposed changes or updates to `develop` or `main`. Local commands exist, but contributors and reviewers cannot rely on a repeatable GitHub check.

## Goals

Run formatting validation, Go tests with 100% statement coverage across Buzz product Go packages, TypeScript UI tests with 100% line, statement, function, and branch coverage, Go vet, and repository build that compiles Buzz.app. Give workflow token read-only permissions and do not expose secrets.

## Non-goals

GitHub Release creation, release asset upload, desktop signing/notarization, Homebrew changes, branch rulesets, and promotion branch policy.

## Design

Add `.github/workflows/ci.yml` with `pull_request` and `push` targeting `develop` and `main`. Require same-repository `feature/*` sources for PRs into `develop`, and `develop` sources for PRs into `main`. Use a macOS runner because Buzz.app is a macOS product and its build requires Xcode tooling. Install Go 1.25.6 from `go.work`, Node 22, and pnpm 10.11.0 from `package.json`. Install dependencies with the lockfile and run the repository's existing commands.

Extend Makefile `test` and `vet` targets to cover `apps/cli` and all modules in `go.work`; add non-mutating formatting and coverage checks so local and hosted validation use the same entry points. Go coverage must report 100% statements for every package in each remaining workspace module. Add Vitest UI tests and enforce 100% lines, statements, functions, and branches for first-party frontend source; exclude only generated Wails bindings and non-executable assets. Install JavaScript dependencies and build `@buzz/desktop-frontend` before Go tests because `apps/desktop` embeds `frontend/dist`. Then run frontend coverage, Go coverage/vet, install Wails, and build the Buzz.app DMG. Keep workflow `permissions` read-only. Do not configure branch protection here: workflow checks run automatically, while making them merge-blocking requires repository rules and stable check names after first hosted run.

## Acceptance

Pull requests and pushes to `develop` or `main` start CI. PR source rules enforce `feature/*` → `develop` → `main`. Formatting check, Go tests, Go vet, and 100% Go statement coverage cover every package across all five workspace modules. Frontend tests pass with 100% line, statement, function, and branch coverage. Buzz.app DMG build succeeds on macOS. Any failed test or coverage threshold fails the workflow. Workflow uses read-only permissions and no secrets. Merge blocking is not claimed until repository rules are separately configured.
