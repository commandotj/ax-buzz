# RFC 0017: Remove unsupported Wails iOS target

**Status:** Implemented

## Summary

Remove Buzz's unused Wails iOS target, native project files, simulator/device tasks, and iOS-only dependency helper. Buzz is shipped as a macOS desktop app; this scaffold creates an unsupported build path and expands test and maintenance scope without serving the product.

## Problem

The desktop Taskfiles expose iOS build and device commands, and `apps/desktop/build/ios` contains an Xcode project, Go entrypoints, app assets, and dependency-check code. Buzz has no iOS product requirement or release artifact. Keeping this generated target makes iOS appear supported and pulled its dependency helper into the Go coverage gate.

## Goals

Remove iOS as a Buzz build target and delete its app-owned scaffolding. Keep the macOS DMG build and other existing targets working. Keep Go coverage at 100% for every remaining product package without target-specific exclusions.

## Non-goals

Do not change Android runtime behavior, Windows/Linux targets, CLI, runtime UI, release workflow, or Wails-generated frontend bindings. Android comments may lose obsolete iOS comparisons. Do not add iOS support or replace it with another mobile target.

## Design

Delete `apps/desktop/build/ios/`. Remove the `ios` Taskfile include from `apps/desktop/Taskfile.yml`, iOS device tasks from `apps/desktop/build/Taskfile.yml`, and the commented iOS build configuration from `apps/desktop/build/config.yml`. Remove Android helper comments that describe behavior by comparison to the deleted iOS target. Remove the CI coverage exclusion for the iOS dependency helper so `make test` again covers every Go package in each remaining workspace module. Update RFC 0013 to describe this complete product-code coverage scope.

## Acceptance

Buzz Taskfiles expose no iOS build, simulator, or device task. No app-owned iOS build project, source, or assets remain under `apps/desktop`. `make test` passes with 100% statement coverage for every remaining Go package. `make vet`, frontend tests and build, and macOS DMG packaging pass. RFC 0013 accurately describes the coverage scope and `specify.mjs sync-check` passes.
