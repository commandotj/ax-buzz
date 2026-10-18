# Buzz.app

Wails 3 desktop shell. React and shadcn/ui frontend lives in `frontend/`; Go backend runs the bundled Buzz CLI and streams its versioned JSON Lines events to the window.

## Development

Install Wails CLI and dependencies from repository root:

```sh
pnpm setup:wails
pnpm install
pnpm dev
```

Build the macOS app bundle from repository root with `pnpm build`. The build requires Go 1.25+, Node.js 22+, pnpm, Xcode Command Line Tools, and Homebrew for the CLI's existing operations.
