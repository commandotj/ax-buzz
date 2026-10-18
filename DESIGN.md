---
name: Buzz
description: Greek Sci-Fi Homebrew library with distinct sunlit and celestial appearances.
colors:
  light-marble: "#f1ece2"
  light-surface: "#fbf9f3"
  light-aegean: "#315d70"
  light-terracotta: "#a64d37"
  light-olive: "#617547"
  light-bronze: "#9b8354"
  dark-aegean: "#132940"
  dark-enamel: "#172f49"
  dark-indigo: "#222d4b"
  dark-cyan: "#73c9e0"
  dark-terracotta: "#e27c61"
  warm-ivory: "#f1eadd"
typography:
  display:
    fontFamily: "Georgia, Times New Roman, serif"
    fontSize: "32px"
    fontWeight: 500
    lineHeight: 1.12
    letterSpacing: "-0.025em"
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: 1.5
  data:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "10px"
    fontWeight: 400
    lineHeight: 1.4
rounded:
  control: "5px"
  field: "5px"
  panel: "7px"
  dialog: "14px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  xl: "24px"
  2xl: "32px"
---

# Design System: Buzz

## Overview

**Creative North Star: Greek Celestial Instrument.** Buzz is a Homebrew library imagined as a precision instrument from a future Greek observatory. Architectural rhythm, painted ceramic color, and measured orbital geometry give the shell its identity. Package data stays first. Light and Dark are complete, separate appearances; System follows macOS.

## Color and material

Light layers sunlit marble and pale ceramic with Aegean blue structure, terracotta selection, olive actions, and fine bronze rules. Dark layers midnight Aegean, indigo enamel, and lapis with cyan signals, terracotta states, and warm ivory data. Every surface, dialog, scrollbar, focus ring, and state uses semantic tokens. Accent colors communicate selection, action, update, success, or destructive state; color never carries status alone.

Greek character comes from architectural proportion and one precise meander/entablature detail. Celestial character comes from orbital geometry in the brand and selected-package inspector. Keep ornament small and purposeful.

## Typography and layout

Georgia display headings provide an inscriptional contrast to the restrained system sans interface. Versions, counts, identifiers, and sources use tabular monospace. Preserve the navigation rail, content header, virtualized package list, selected-package inspector, dense rows, and native scrolling. Only package list and inspector scroll; the window does not.

## Appearance and components

The Appearance control offers System, Light, and Dark. System tracks macOS changes while open; explicit choices persist across relaunch. Selected navigation uses terracotta; install and upgrade use olive in Light and luminous Aegean blue in Dark; update, success, and removal have distinct semantic treatments. Focus remains visible in both themes.

Avoid broad glow, glass blur, gold gradients, repetitive borders, oversized ornament, and game-inspired interface motifs.
