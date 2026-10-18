package main

import (
	_ "embed"
	"fmt"
	"github.com/commandotj/ax-buzz/apps/cli/internal/model"
	"github.com/commandotj/ax-buzz/apps/cli/internal/ui"
	"github.com/commandotj/ax-buzz/apps/cli/internal/util"
	"github.com/commandotj/ax-buzz/modules/protocol/protocol"
	"io"
	"log"
	"os"
	"strings"

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
