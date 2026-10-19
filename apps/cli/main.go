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
	flagShowVersion  = pflag.BoolP("version", "v", false, "Show app version")
	flagShowHelp     = pflag.BoolP("help", "h", false, "Show help message")
	loadCatalog      = brew.LoadPackages
	performAction    = brew.RunPackageAction
	processExit      = os.Exit
	startProgram     = runProgram
	newSignalContext = signal.NotifyContext
)

//go:embed .version
var version string

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	processExit(1)
}

func runProgram(p *tea.Program) error {
	_, err := p.Run()
	return err
}

func run(args []string, output, diagnostics io.Writer) error {
	if len(args) > 0 && args[0] == "--stream-json" {
		log.SetOutput(diagnostics)
		return runStreamJSON(output, args[1:])
	}
	if len(args) > 0 && args[0] == "--machine" {
		return runMachine(output, args[1:])
	}

	flagSet := pflag.NewFlagSet("buzz", pflag.ContinueOnError)
	flagSet.SetOutput(diagnostics)
	flagSet.Usage = func() { flagSet.PrintDefaults() }
	pflag.CommandLine.VisitAll(flagSet.AddFlag)
	if err := flagSet.Parse(args); err != nil && err != pflag.ErrHelp {
		return err
	}

	if *flagShowVersion {
		_, err := fmt.Fprint(output, strings.Fields(version)[0])
		return err
	}

	if *flagShowHelp {
		flagSet.Usage()
		return nil
	}

	ui.InitTheme()

	logfile := util.GetEnv("BUZZ_LOG", util.GetEnv("TAPROOM_LOG", "/tmp/taproom.log"))
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer f.Close()
	// Send log output to the file
	log.SetOutput(f)

	// The WithAltScreen() option provides a full-screen TUI experience.
	p := tea.NewProgram(model.InitialModel(), tea.WithAltScreen())
	if err := startProgram(p); err != nil {
		return fmt.Errorf("Alas, there's been an error: %w", err)
	}
	return nil
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
		"version": strings.Fields(version)[0],
	})
}

type streamOperation struct {
	name      string
	packageID string
	isCask    bool
}

func runStreamJSON(output io.Writer, args []string) error {
	writer := protocol.NewWriter(output)
	operation, err := parseStreamOperation(args)
	if err != nil {
		return finishStream(writer, err)
	}
	ctx, stop := newSignalContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	progress := func(message string) {
		_ = writer.Progress(map[string]string{"message": message})
	}
	forceRefresh := operation.name == "refresh"
	snapshot, err := loadCatalog(ctx, forceRefresh, progress)
	if err != nil {
		return finishStream(writer, err)
	}
	if operation.name == "install" || operation.name == "uninstall" || operation.name == "upgrade" {
		pkg := findPackage(snapshot.Packages, operation.packageID, operation.isCask)
		if pkg == nil {
			return finishStream(writer, fmt.Errorf("package %q was not found", operation.packageID))
		}
		if !pkg.InstallSupported {
			return finishStream(writer, fmt.Errorf("package %q is not supported by Buzz", pkg.Name))
		}
		if operation.name == "upgrade" && !pkg.IsInstalled {
			return finishStream(writer, fmt.Errorf("package %q is not installed", pkg.Name))
		}
		progress(fmt.Sprintf("Running brew %s for %s", operation.name, pkg.Name))
		if err := performAction(ctx, operation.name, pkg.Name, pkg.IsCask, progress); err != nil {
			return finishStream(writer, err)
		}
		return writer.Result(map[string]any{"success": true, "operation": operation.name, "package": pkg.Name})
	}

	packages := snapshot.Packages
	if operation.name == "installed" {
		packages = installedPackages(packages)
	}
	for start := 0; start < len(packages); start += 100 {
		if err := ctx.Err(); err != nil {
			return finishStream(writer, err)
		}
		end := min(start+100, len(packages))
		if err := writer.Data(map[string]any{"packages": packages[start:end]}); err != nil {
			return fmt.Errorf("write package data: %w", err)
		}
	}
	if err := ctx.Err(); err != nil {
		return finishStream(writer, err)
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

func finishStream(writer *protocol.Writer, err error) error {
	_ = writer.Result(map[string]any{"success": false, "error": err.Error()})
	return err
}
