# RFC 0013: Buzz pull request and main CI (child of 0012)

**Status:** Approved

**Parent:** [0012](0012-buzz-ci-cd.md)

## Summary

Add read-only GitHub Actions checks for pull requests targeting `main` and pushes to `main`. Checks validate Go formatting, test and vet every Go workspace module, then build Buzz.app on macOS using the repository's existing package-manager commands.

## Problem

There is no hosted check on proposed changes or updates to `main`. Local commands exist, but contributors and reviewers cannot rely on a repeatable GitHub check.

## Goals

Run formatting validation, Go tests and vetting across every Go workspace module, plus the repository build that compiles the CLI and Buzz.app. Give workflow token read-only permissions and do not expose secrets.

## Non-goals

GitHub Release creation, release asset upload, desktop signing/notarization, Homebrew changes, branch rulesets, promotion branch policy, and adding new test/lint frameworks.

## Design

Add `.github/workflows/ci.yml` with `pull_request` targeting `main` and `push` targeting `main`. Use a macOS runner because Buzz.app is a macOS product and its build requires Xcode tooling. Install Go 1.25.6 from `go.work`, Node 22, and pnpm 10.11.0 from `package.json`. Install dependencies with the lockfile and run the repository's existing commands.

Extend Makefile `test` and `vet` targets to cover `apps/cli` and all modules in `go.work`; add a non-mutating formatting check so local and hosted validation use the same entry points. Install JavaScript dependencies and build `@buzz/desktop-frontend` before Go tests because `apps/desktop` embeds `frontend/dist`. Then run Go test/vet, install Wails, and build the Buzz.app DMG. Keep workflow `permissions` read-only. Do not configure branch protection here: workflow checks run automatically, while making them merge-blocking requires repository rules and stable check names after first hosted run.

## Acceptance

Pull requests to `main` and pushes to `main` start the same CI checks. Formatting check, Go tests and vet cover all five workspace modules; workspace build succeeds on macOS. Any failed check fails the workflow. Workflow uses read-only permissions and no secrets. Merge blocking is not claimed until repository rules are separately configured.
