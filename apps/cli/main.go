package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/commandotj/ax-buzz/apps/cli/internal/model"
	"github.com/commandotj/ax-buzz/apps/cli/internal/ui"
	"github.com/commandotj/ax-buzz/apps/cli/internal/util"
	"github.com/commandotj/ax-buzz/modules/brew/brew"
	"github.com/commandotj/ax-buzz/modules/brew/data"
	"github.com/commandotj/ax-buzz/modules/protocol/protocol"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"
)

var (
	flagShowVersion = pflag.BoolP("version", "v", false, "Show app version")
	flagShowHelp    = pflag.BoolP("help", "h", false, "Show help message")
)

//go:embed .version
var version string

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--stream-json" {
		log.SetOutput(os.Stderr)
		if err := runStreamJSON(os.Stdout, os.Stderr, os.Args[2:]); err != nil {
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--machine" {
		if err := runMachine(os.Stdout, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	pflag.Parse()

	if *flagShowVersion {
		fmt.Print(version)
		os.Exit(0)
	}

	if *flagShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	ui.InitTheme()

	logfile := util.GetEnv("BUZZ_LOG", util.GetEnv("TAPROOM_LOG", "/tmp/taproom.log"))
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}
	defer f.Close()
	// Send log output to the file
	log.SetOutput(f)

	// The WithAltScreen() option provides a full-screen TUI experience.
	p := tea.NewProgram(model.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func runMachine(output io.Writer, args []string) error {
	if len(args) != 1 || args[0] != "handshake" {
		return fmt.Errorf("unsupported machine command; expected handshake")
	}

	writer := protocol.NewWriter(output)
	if err := writer.Progress(map[string]string{"message": "Buzz CLI ready"}); err != nil {
		return err
	}
	return writer.Result(map[string]string{
		"product": "Buzz CLI",
		"version": strings.TrimSpace(version),
	})
}

type streamOperation struct {
	name      string
	packageID string
	isCask    bool
}

func runStreamJSON(output io.Writer, diagnostics io.Writer, args []string) error {
	writer := protocol.NewWriter(output)
	operation, err := parseStreamOperation(args)
	if err != nil {
		return finishStream(writer, diagnostics, err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	progress := func(message string) {
		_ = writer.Progress(map[string]string{"message": message})
	}
	forceRefresh := operation.name == "refresh"
	snapshot, err := brew.LoadPackages(ctx, forceRefresh, progress)
	if err != nil {
		return finishStream(writer, diagnostics, err)
	}
	if operation.name == "install" || operation.name == "uninstall" || operation.name == "upgrade" {
		pkg := findPackage(snapshot.Packages, operation.packageID, operation.isCask)
		if pkg == nil {
			return finishStream(writer, diagnostics, fmt.Errorf("package %q was not found", operation.packageID))
		}
		if !pkg.InstallSupported {
			return finishStream(writer, diagnostics, fmt.Errorf("package %q is not supported by Buzz", pkg.Name))
		}
		if operation.name == "upgrade" && !pkg.IsInstalled {
			return finishStream(writer, diagnostics, fmt.Errorf("package %q is not installed", pkg.Name))
		}
		progress(fmt.Sprintf("Running brew %s for %s", operation.name, pkg.Name))
		if err := brew.RunPackageAction(ctx, operation.name, pkg.Name, pkg.IsCask, progress); err != nil {
			return finishStream(writer, diagnostics, err)
		}
		return writer.Result(map[string]any{"success": true, "operation": operation.name, "package": pkg.Name})
	}

	packages := snapshot.Packages
	if operation.name == "installed" {
		packages = installedPackages(packages)
	}
	for start := 0; start < len(packages); start += 100 {
		if err := ctx.Err(); err != nil {
			return finishStream(writer, diagnostics, err)
		}
		end := min(start+100, len(packages))
		if err := writer.Data(map[string]any{"packages": packages[start:end]}); err != nil {
			return fmt.Errorf("write package data: %w", err)
		}
	}
	if err := ctx.Err(); err != nil {
		return finishStream(writer, diagnostics, err)
	}
	return writer.Result(map[string]any{
		"success":         true,
		"operation":       operation.name,
		"count":           len(packages),
		"cacheAgeSeconds": int64(snapshot.CacheAge.Seconds()),
		"cacheStale":      snapshot.Stale,
	})
}

func parseStreamOperation(args []string) (streamOperation, error) {
	if len(args) == 0 {
		return streamOperation{}, fmt.Errorf("expected list, installed, refresh, install, uninstall, or upgrade")
	}
	operation := streamOperation{name: args[0]}
	for _, arg := range args[1:] {
		if arg == "--cask" {
			operation.isCask = true
			continue
		}
		if operation.packageID != "" {
			return streamOperation{}, fmt.Errorf("unexpected argument %q", arg)
		}
		operation.packageID = arg
	}
	switch operation.name {
	case "list", "installed", "refresh":
		if operation.packageID != "" || operation.isCask {
			return streamOperation{}, fmt.Errorf("%s does not accept package arguments", operation.name)
		}
	case "install", "uninstall", "upgrade":
		if operation.packageID == "" {
			return streamOperation{}, fmt.Errorf("%s requires a package name", operation.name)
		}
	default:
		return streamOperation{}, fmt.Errorf("unsupported operation %q", operation.name)
	}
	return operation, nil
}

func findPackage(packages []*data.Package, name string, isCask bool) *data.Package {
	for _, pkg := range packages {
		if pkg.Name == name && pkg.IsCask == isCask {
			return pkg
		}
	}
	return nil
}

func installedPackages(packages []*data.Package) []*data.Package {
	installed := make([]*data.Package, 0)
	for _, pkg := range packages {
		if pkg.IsInstalled {
			installed = append(installed, pkg)
		}
	}
	return installed
}

func finishStream(writer *protocol.Writer, diagnostics io.Writer, err error) error {
	_, _ = fmt.Fprintln(diagnostics, err)
	_ = writer.Result(map[string]any{"success": false, "error": err.Error()})
	return err
}
