package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/commandotj/ax-buzz/modules/protocol/protocol"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type cliEvent struct {
	ProtocolVersion int             `json:"protocolVersion"`
	Event           string          `json:"event"`
	Data            json.RawMessage `json:"data,omitempty"`
	Error           string          `json:"error,omitempty"`
}

type CLIService struct {
	app       *application.App
	emitEvent func(cliEvent)
	mu        sync.Mutex
	cancel    context.CancelFunc
}

var executablePath = os.Executable
var newCLICommand = exec.CommandContext
var emitWailsEvent = (*application.EventManager).Emit

func NewCLIService() *CLIService {
	return &CLIService{}
}

func (s *CLIService) SetApp(app *application.App) {
	s.app = app
}

func (s *CLIService) StartHandshake() error {
	return s.start([]string{"--machine", "handshake"})
}

func (s *CLIService) StartOperation(operation, packageName string, isCask bool) error {
	args := []string{"--stream-json", operation}
	if packageName != "" {
		args = append(args, packageName)
	}
	if isCask {
		args = append(args, "--cask")
	}
	return s.start(args)
}

func (s *CLIService) start(args []string) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return fmt.Errorf("Buzz CLI operation already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	go s.runCLI(ctx, args)
	return nil
}

func (s *CLIService) CancelOperation() bool {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func (s *CLIService) CancelHandshake() bool {
	return s.CancelOperation()
}

func (s *CLIService) ServiceShutdown() error {
	s.CancelHandshake()
	return nil
}

func (s *CLIService) runCLI(ctx context.Context, args []string) {
	defer func() {
		s.mu.Lock()
		s.cancel = nil
		s.mu.Unlock()
	}()

	cliPath, err := resolveCLIPath()
	if err != nil {
		s.emitError(err.Error())
		return
	}
	cmd := newCLICommand(ctx, cliPath, args...)
	cmd.Cancel = func() error { return interruptCLIProcess(cmd) }
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.emitError(fmt.Sprintf("start Buzz CLI output: %v", err))
		return
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		s.emitError(fmt.Sprintf("start Buzz CLI: %v", err))
		return
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	resultSeen := false
	var resultEvent *cliEvent
	var protocolErr error
	for scanner.Scan() {
		var event cliEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			protocolErr = fmt.Errorf("Buzz CLI returned invalid JSON: %w", err)
			_ = cmd.Process.Kill()
			break
		}
		if err := protocol.ValidateEvent(event.ProtocolVersion, event.Event, len(event.Data) > 0, resultSeen); err != nil {
			protocolErr = err
			_ = cmd.Process.Kill()
			break
		}
		if event.Event == "result" {
			resultSeen = true
			resultEvent = &event
			continue
		}
		s.emit(event)
	}
	if err := scanner.Err(); err != nil && protocolErr == nil {
		protocolErr = fmt.Errorf("read Buzz CLI output: %w", err)
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		if resultEvent != nil {
			s.emit(*resultEvent)
		} else {
			s.emitError("Buzz CLI operation cancelled")
		}
		return
	}
	if protocolErr != nil {
		s.emitError(protocolErr.Error())
		return
	}
	if waitErr != nil {
		if resultEvent != nil {
			s.emit(*resultEvent)
			return
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = waitErr.Error()
		}
		s.emitError(message)
		return
	}
	if resultEvent == nil {
		s.emitError("Buzz CLI exited without a result event")
		return
	}
	s.emit(*resultEvent)
}

func resolveCLIPath() (string, error) {
	executable, err := executablePath()
	if err != nil {
		return "", fmt.Errorf("locate Buzz.app executable: %w", err)
	}
	return findCLIPath(os.Getenv("BUZZ_CLI_PATH"), executable)
}

func interruptCLIProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(os.Interrupt)
}

func findCLIPath(override, executable string) (string, error) {
	if override != "" {
		return filepath.Abs(override)
	}
	executableDir := filepath.Dir(executable)
	candidates := []string{
		filepath.Join(executableDir, "..", "Resources", "buzz"),
		filepath.Join(executableDir, "..", "..", "cli", "bin", "buzz"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("Buzz CLI not found; build it with `pnpm build:cli` or bundle it with Buzz.app")
}

func (s *CLIService) emit(event cliEvent) {
	if s.emitEvent != nil {
		s.emitEvent(event)
		return
	}
	if s.app != nil {
		emitWailsEvent(s.app.Event, "buzz:cli-event", event)
	}
}

func (s *CLIService) emitError(message string) {
	s.emit(cliEvent{ProtocolVersion: protocol.Version, Event: "error", Error: message})
}
