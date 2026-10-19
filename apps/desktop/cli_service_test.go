package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestFindCLIPath(t *testing.T) {
	t.Run("override", func(t *testing.T) {
		got, err := findCLIPath("relative/buzz", "/bundle/MacOS/Buzz")
		if err != nil {
			t.Fatal(err)
		}
		want, _ := filepath.Abs("relative/buzz")
		if got != want {
			t.Fatalf("findCLIPath() = %q, want %q", got, want)
		}
	})

	t.Run("resources bundle", func(t *testing.T) {
		root := t.TempDir()
		executable := filepath.Join(root, "Buzz.app", "Contents", "MacOS", "Buzz")
		cli := filepath.Join(root, "Buzz.app", "Contents", "Resources", "buzz")
		if err := os.MkdirAll(filepath.Dir(cli), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cli, []byte("cli"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := findCLIPath("", executable)
		if err != nil || got != cli {
			t.Fatalf("findCLIPath() = %q, %v; want %q", got, err, cli)
		}
	})

	t.Run("development bundle", func(t *testing.T) {
		root := t.TempDir()
		executable := filepath.Join(root, "apps", "desktop", "build", "Buzz")
		cli := filepath.Join(root, "apps", "cli", "bin", "buzz")
		if err := os.MkdirAll(filepath.Dir(cli), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cli, []byte("cli"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := findCLIPath("", executable)
		if err != nil || got != cli {
			t.Fatalf("findCLIPath() = %q, %v; want %q", got, err, cli)
		}
	})

	t.Run("directory ignored", func(t *testing.T) {
		root := t.TempDir()
		executable := filepath.Join(root, "Buzz.app", "Contents", "MacOS", "Buzz")
		candidate := filepath.Join(root, "Buzz.app", "Contents", "Resources", "buzz")
		if err := os.MkdirAll(candidate, 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := findCLIPath("", executable); err == nil {
			t.Fatal("findCLIPath() unexpectedly accepted directory")
		}
	})
}

func TestResolveCLIPathExecutableFailure(t *testing.T) {
	old := executablePath
	defer func() { executablePath = old }()
	executablePath = func() (string, error) { return "", errors.New("no executable") }
	if _, err := resolveCLIPath(); err == nil || !strings.Contains(err.Error(), "locate Buzz.app executable") {
		t.Fatalf("resolveCLIPath() error = %v", err)
	}
}

func TestCLIServiceStartAndCancel(t *testing.T) {
	svc := NewCLIService()
	svc.SetApp(nil)
	if svc.CancelOperation() || svc.CancelHandshake() {
		t.Fatal("cancel without operation returned true")
	}
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	svc.cancel = cancel
	if !svc.CancelOperation() {
		t.Fatal("cancel active operation returned false")
	}
	if ctx.Err() == nil {
		t.Fatal("operation context was not cancelled")
	}

	svc = NewCLIService()
	svc.cancel = func() {}
	if err := svc.start(nil); err == nil || err.Error() != "Buzz CLI operation already running" {
		t.Fatalf("start with active operation error = %v", err)
	}
	svc.cancel()
	cli := writeCLI(t, `printf '{"protocolVersion":1,"event":"result","data":{}}\\n'`, "", false)
	t.Setenv("BUZZ_CLI_PATH", cli)
	svc = NewCLIService()
	if err := svc.StartHandshake(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		svc.mu.Lock()
		active := svc.cancel != nil
		svc.mu.Unlock()
		if !active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("started CLI did not complete")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestCLIServiceRunCLIEventsAndErrors(t *testing.T) {
	tests := []struct {
		name string
		out  string
		err  string
		want string
		fail bool
	}{
		{name: "success", out: `{"protocolVersion":1,"event":"progress","data":{"step":1}}` + "\n" + `{"protocolVersion":1,"event":"result","data":{"ok":true}}` + "\n", want: "result"},
		{name: "invalid json", out: "not json\n", want: "error"},
		{name: "invalid protocol", out: `{"protocolVersion":99,"event":"progress","data":{}}` + "\n", want: "error"},
		{name: "missing result", out: `{"protocolVersion":1,"event":"progress","data":{}}` + "\n", want: "error"},
		{name: "stderr failure", out: "", err: "cli failed\n", want: "error"},
		{name: "empty stderr failure", out: "", err: "", want: "error", fail: true},
		{name: "result on nonzero exit", out: `{"protocolVersion":1,"event":"result","data":{}}` + "\n", err: "failed after result", want: "result", fail: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cli := writeCLI(t, tc.out, tc.err, tc.fail)
			t.Setenv("BUZZ_CLI_PATH", cli)
			var events []cliEvent
			svc := &CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}
			svc.runCLI(context.Background(), nil)
			if len(events) == 0 || events[len(events)-1].Event != tc.want {
				t.Fatalf("events = %#v, want final event %q", events, tc.want)
			}
			if tc.want == "result" && events[len(events)-1].Event != "result" {
				t.Fatalf("result event not emitted: %#v", events)
			}
		})
	}
}

func TestCLIServiceRunCLIFailures(t *testing.T) {
	t.Run("CLI path resolution failure", func(t *testing.T) {
		old := executablePath
		defer func() { executablePath = old }()
		executablePath = func() (string, error) { return "", errors.New("missing executable") }
		var events []cliEvent
		(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(context.Background(), nil)
		if len(events) != 1 || !strings.Contains(events[0].Error, "locate Buzz.app executable") {
			t.Fatalf("events = %#v", events)
		}
	})
	t.Run("stdout pipe failure", func(t *testing.T) {
		t.Setenv("BUZZ_CLI_PATH", filepath.Join(t.TempDir(), "buzz"))
		old := newCLICommand
		defer func() { newCLICommand = old }()
		newCLICommand = func(context.Context, string, ...string) *exec.Cmd { return &exec.Cmd{Stdout: os.Stdout} }
		var events []cliEvent
		(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(context.Background(), nil)
		if len(events) != 1 || !strings.Contains(events[0].Error, "start Buzz CLI output") {
			t.Fatalf("events = %#v", events)
		}
	})
	t.Run("missing executable", func(t *testing.T) {
		t.Setenv("BUZZ_CLI_PATH", filepath.Join(t.TempDir(), "missing"))
		var events []cliEvent
		(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(context.Background(), nil)
		if len(events) != 1 || !strings.Contains(events[0].Error, "start Buzz CLI") {
			t.Fatalf("events = %#v", events)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "ready")
		cli, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("BUZZ_CLI_PATH", cli)
		t.Setenv("BUZZ_TEST_READY", marker)
		t.Setenv("BUZZ_TEST_CANCEL_CHILD", "1")
		ctx, cancel := context.WithCancel(context.Background())
		var events []cliEvent
		done := make(chan struct{})
		go func() {
			defer close(done)
			(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(ctx, []string{"-test.run=^TestCLIServiceCancelChild$"})
		}()
		deadline := time.Now().Add(3 * time.Second)
		for {
			if _, err := os.Stat(marker); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("CLI did not start")
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("cancelled process did not stop")
		}
		if len(events) != 1 || events[0].Error != "Buzz CLI operation cancelled" {
			t.Fatalf("events = %#v", events)
		}
	})
	t.Run("cancel after result", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "ready")
		cli, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("BUZZ_CLI_PATH", cli)
		t.Setenv("BUZZ_TEST_READY", marker)
		t.Setenv("BUZZ_TEST_CANCEL_CHILD", "1")
		t.Setenv("BUZZ_TEST_CANCEL_CHILD_RESULT", "1")
		ctx, cancel := context.WithCancel(context.Background())
		var events []cliEvent
		done := make(chan struct{})
		go func() {
			defer close(done)
			(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(ctx, []string{"-test.run=^TestCLIServiceCancelChild$"})
		}()
		deadline := time.Now().Add(3 * time.Second)
		for {
			if _, err := os.Stat(marker); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("CLI did not produce result")
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("cancelled process did not stop")
		}
		if len(events) != 1 || events[0].Event != "result" {
			t.Fatalf("events = %#v", events)
		}
	})
	t.Run("oversized line", func(t *testing.T) {
		cli, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("BUZZ_CLI_PATH", cli)
		t.Setenv("BUZZ_TEST_LARGE_LINE", "1")
		var events []cliEvent
		(&CLIService{emitEvent: func(event cliEvent) { events = append(events, event) }}).runCLI(context.Background(), []string{"-test.run=TestLargeLineChild"})
		if len(events) != 1 || !strings.Contains(events[0].Error, "read Buzz CLI output") {
			t.Fatalf("events = %#v", events)
		}
	})
}

func TestCLIServiceCancelChild(t *testing.T) {
	if os.Getenv("BUZZ_TEST_CANCEL_CHILD") != "1" {
		return
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	if os.Getenv("BUZZ_TEST_CANCEL_CHILD_RESULT") == "1" {
		_, _ = os.Stdout.Write([]byte("{\"protocolVersion\":1,\"event\":\"result\",\"data\":{}}\n"))
	}
	if err := os.WriteFile(os.Getenv("BUZZ_TEST_READY"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	<-signals
}

func TestLargeLineChild(t *testing.T) {
	if os.Getenv("BUZZ_TEST_LARGE_LINE") != "1" {
		return
	}
	_, _ = os.Stdout.Write([]byte(strings.Repeat("x", 17*1024*1024)))
}

func TestCLIServiceEmitWithoutApp(t *testing.T) {
	(&CLIService{}).emit(cliEvent{Event: "progress"})
	old := emitWailsEvent
	defer func() { emitWailsEvent = old }()
	var emitted bool
	emitWailsEvent = func(_ *application.EventManager, name string, data ...any) bool {
		emitted = name == "buzz:cli-event" && len(data) == 1
		return false
	}
	(&CLIService{app: &application.App{}}).emit(cliEvent{Event: "progress"})
	if !emitted {
		t.Fatal("app event was not emitted")
	}
}

func TestInterruptCLIProcessWithoutProcess(t *testing.T) {
	if err := interruptCLIProcess(&exec.Cmd{}); err != nil {
		t.Fatalf("interruptCLIProcess() = %v", err)
	}
}

func writeRawCLI(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buzz")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCLIServiceStartOperationArguments(t *testing.T) {
	svc := &CLIService{cancel: func() {}}
	if err := svc.StartHandshake(); err == nil {
		t.Fatal("StartHandshake did not reject concurrent operation")
	}
	if err := svc.StartOperation("install", "", false); err == nil {
		t.Fatal("StartOperation did not reject concurrent operation")
	}
	if err := svc.StartOperation("install", "wget", true); err == nil {
		t.Fatal("StartOperation did not reject concurrent operation")
	}
}

func writeCLI(t *testing.T, stdout, stderr string, fail bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buzz")
	script := "#!/bin/sh\n"
	if stdout != "" {
		script += "cat <<'BUZZ_OUT'\n" + stdout + "BUZZ_OUT\n"
	}
	if stderr != "" {
		script += "printf '%s' " + shellQuote(stderr) + " >&2\n"
	}
	if fail {
		script += "exit 1\n"
	}
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestCLIEventJSONShape(t *testing.T) {
	encoded, err := json.Marshal(cliEvent{ProtocolVersion: 1, Event: "error", Error: "failed"})
	if err != nil || !strings.Contains(string(encoded), `"error":"failed"`) {
		t.Fatalf("event JSON = %s, %v", encoded, err)
	}
}
