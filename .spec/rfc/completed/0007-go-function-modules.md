# RFC 0007: Buzz Go Function Modules (child of 0001)

**Status:** Implemented

**Parent:** [0001](0001-buzz-installer.md)

## Summary

Define Go module boundaries for Buzz's CLI and desktop application. Keep application projects under `apps/`; place reusable Go functional modules under `modules/`, each with its own `go.mod`; use root `go.work` only to coordinate local modules. This RFC covers package ownership and migration only.

## Problem

RFC 0005 correctly places executable projects under `apps/`, but its initial package migration puts every Go package inside `apps/cli`. The desktop then cannot import reusable Go functionality without depending on the CLI application's private `internal` packages. RFC 0006 requires the CLI and desktop to share protocol and package behavior, but its module layout was mixed into streaming operation design and mechanically mirrored existing directories.

## Goals

The CLI and desktop remain complete Go application projects in `apps/cli` and `apps/desktop`. Reusable functional packages live in independent Go modules under `modules/`. Each module owns one named capability, exposes a small public package API, and declares only the dependencies required by that capability. Root `go.work` includes both app modules and every module under `modules/`; repository root has no `go.mod` and no Go source.

## Non-goals

- CLI JSONL commands, event semantics, cache behavior, and cancellation (RFC 0006).
- Desktop screens, navigation, visual style, and interaction behavior (RFC 0008).
- Moving executable projects out of `apps/` or placing generated binaries in `modules/`.
- Creating empty modules solely to mirror old directory names.

## Design

Keep `apps/cli` as the Buzz CLI project, including `main.go`, its `go.mod`, tests, TUI entry point, and pnpm build tasks. Keep `apps/desktop` as the Wails project, including its `main.go`, `go.mod`, backend, frontend, and build configuration. Generated outputs stay app-local: `apps/cli/bin/buzz` and `apps/desktop/bin/Buzz.app`.

Create exactly these initial function modules and package locations:

| Go module | Packages moved into module | Public responsibility |
|---|---|---|
| `modules/brew` | `internal/brew` → `brew`; `internal/data` → `data`; `internal/loading` → `loading`; `internal/util.Sort`, `internal/util.SortAndUniq`, and `internal/util.FormatSize` → `util` | Homebrew catalog, inventory, package actions, package records, catalog-load progress, and package-data formatting helpers. `util` exports only behavior needed by these packages or CLI consumers. |
| `modules/gh` | `internal/gh` → `gh` | Read GitHub release metadata through installed `gh` executable. Its API accepts repository URL strings and returns a module-owned release record; it must not import `brew` or the Buzz app. |
| `modules/protocol` | `internal/protocol` → `protocol` | Encode and validate versioned JSONL progress, data, and terminal result events shared by CLI and desktop Go code. |

Keep `internal/model` and `internal/ui` in `apps/cli`; they implement Bubble Tea state and rendering, not reusable software-management capabilities. Keep `internal/util.GetEnv` in `apps/cli` because it serves CLI configuration. Move `Sort`, `SortAndUniq`, and `FormatSize` to `modules/brew/util`: Homebrew catalog construction uses sorting and package records expose formatted installed size. Update the TUI loading screen to consume `modules/brew/loading`; keep visual rendering in `apps/cli/internal/ui`. If `data.Package` stores GitHub release metadata, use the release record from `modules/gh`; keep `modules/gh` independent so dependency direction remains `modules/brew` → `modules/gh`, never the reverse. `modules/protocol` has no dependency on either app or Homebrew implementation.

Each listed module has its own `go.mod`. App and module imports use module paths. Root `go.work` lists `apps/cli`, `apps/desktop`, `modules/brew`, `modules/gh`, and `modules/protocol`. Do not add root `go.mod`, root Go source, a generic `core` module, or a `modules/cli` module. Do not create modules solely to mirror `internal` directories. Preserve CLI behavior, Taproom compatibility entry points, and existing cache/data paths while moving code. No module may import an app; avoid package and module dependency cycles.

Update app imports and build commands in the same migration. This RFC overrides RFC 0005 only on the location of reusable Go packages: app projects stay under `apps/`, reusable function modules move to `modules/`.

## Acceptance

From a clean checkout, `go work` resolves both app modules and every existing `modules/*/go.mod`; no root `go.mod` or root Go source exists. `go test ./...` succeeds from each module and app directory using the workspace. `apps/cli` builds `apps/cli/bin/buzz`, and `apps/desktop` builds without importing `apps/cli/internal`. Package actions and TUI remain reachable through the CLI. The dependency graph has no import cycle and no `modules/` package imports an app. The resulting module inventory matches documented functional responsibilities; TUI-only packages remain in `apps/cli`.
