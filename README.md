# Buzz Installer

Buzz Installer is a macOS software discovery and management app with a Go CLI and a Wails 3 desktop shell built with React and shadcn/ui.

The desktop app will browse packages from Luban's Homebrew tap and other configured taps, show installed macOS apps and CLI tools, monitor available updates, and manage supported packages. The CLI will install and inspect packages, report versions, and support automation. Automatic update monitoring and execution belong to the desktop app.

## Project status

The inherited Taproom Go TUI now lives in `apps/cli` as Buzz CLI. `apps/desktop` contains the Buzz.app shell, which launches its bundled CLI and displays protocol progress and results. Package discovery and management features remain tracked in the [RFC roadmap](.spec/ROADMAP.md). The `taproom` executable, `TAPROOM_LOG` setting, and `~/.cache/taproom` cache path remain available for compatibility.

## Current CLI

Buzz CLI requires macOS, Go 1.24.4 or later, Homebrew, and a terminal with a Nerd Font. Build both canonical and compatibility executables with:

```sh
make
./buzz
```

`./taproom` remains available and opens the same TUI. Buzz can browse Homebrew formulae and casks, inspect dependencies, and invoke Homebrew install, uninstall, upgrade, and pin commands. Machine mode currently supports the versioned app handshake; the package commands defined by RFCs 0002–0004 remain future work.

## Desktop app

Install the pinned Wails 3 CLI and run the app from the repository root:

```sh
pnpm setup:wails
pnpm install
pnpm dev
```

Build `apps/desktop/bin/Buzz.app` with `pnpm build`. Desktop development requires Go 1.25+, Node.js 22+, pnpm, and Xcode Command Line Tools.

## License and upstream

The current Go terminal interface derives from [Taproom](https://github.com/hzqtc/taproom). See [LICENSE](LICENSE) for license terms.
