# RFC 0011: Use the Buzz app icon in the app shell

**Status:** Implemented

## Summary

Use one transparent celestial-cycle mark across the macOS app icon and desktop app shell. The orbital-cycle artwork itself fills the icon canvas; no square background tile or separate shell glyph competes with it.

## Problem

The app shell currently draws a Lucide `Orbit` symbol inside a separate decorative frame. That mark differs from the Buzz icon shown by the Dock and Finder. The current icon artwork also has an opaque square fill, contrary to the requested transparent cycle mark.

## Goals

- Remove the icon's opaque background and retain full orbital cycles on a transparent canvas.
- Replace the shell's placeholder orbit symbol with the same transparent cycle artwork used by the app bundle.
- Keep the mark crisp at sidebar and compact window-rail sizes in both Light and Dark themes.
- Keep navigation and package-type icons unchanged.

## Non-goals

- Changing package, navigation, or action iconography.
- Changing layout, theme tokens, app identity, or window behavior.

## Design

Make the existing celestial-cycle SVG transparent outside its orbital artwork, regenerate the existing app icon exports, and use that same source in the shell brand lockup and macOS app bundle. Remove the Orbit glyph and icon-specific frame styling where those styles conflict with the shared mark. Keep the asset's aspect ratio and use the surrounding Buzz brand label for accessible naming. Update the existing logo source/export, `apps/desktop/frontend/src/App.tsx`, `apps/desktop/frontend/src/index.css`, and app icon bundle inputs as required by the established Wails asset pipeline.

## Acceptance

- Sidebar brand lockup and macOS app icon use identical celestial-cycle artwork.
- Artwork has transparent exterior and no grey or colored tile in Light or Dark appearance.
- Mark remains sharp at normal and compact window sizes; no layout shift or distortion.
- Navigation, package rows, theme switching, and macOS window controls retain behavior.
- Frontend typecheck and desktop build pass.
