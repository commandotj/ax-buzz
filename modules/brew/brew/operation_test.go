package brew

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRunPackageAction(t *testing.T) {
	oldExe, oldErr := brewExecutable, brewExecutableErr
	brewExecutable = "/bin/echo"
	brewExecutableErr = nil
	t.Cleanup(func() { brewExecutable, brewExecutableErr = oldExe, oldErr })
	if err := RunPackageAction(context.Background(), "upgrade", "", false, nil); err == nil {
		t.Fatal("empty package name should fail")
	}
	var lines []string
	if err := RunPackageAction(context.Background(), "upgrade", "demo", true, func(s string) { lines = append(lines, s) }); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || !strings.Contains(lines[0], "--cask demo") {
		t.Fatalf("progress=%v", lines)
	}
	brewExecutableErr = context.DeadlineExceeded
	if err := RunPackageAction(context.Background(), "install", "demo", false, nil); err == nil {
		t.Fatal("missing brew should fail")
	}
	brewExecutableErr = nil
	brewExecutable = "/usr/bin/false"
	if err := RunPackageAction(context.Background(), "install", "demo", false, nil); err == nil || !strings.Contains(err.Error(), "brew install demo") {
		t.Fatalf("failure=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := RunPackageAction(ctx, "install", "demo", false, nil); err == nil {
		t.Fatal("canceled command should fail")
	}
	brewExecutable = "/bin/sleep"
	ctx, cancel = context.WithCancel(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	if err := runBrew(ctx, []string{"2"}, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled running command=%v", err)
	}
}

func TestRunBrewPipeSetupErrors(t *testing.T) {
	oldFactory, oldExe, oldErr := newBrewExecCommandContext, brewExecutable, brewExecutableErr
	defer func() { newBrewExecCommandContext, brewExecutable, brewExecutableErr = oldFactory, oldExe, oldErr }()
	brewExecutable = "/bin/echo"
	brewExecutableErr = nil
	newBrewExecCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout = io.Discard
		return cmd
	}
	if err := runBrew(context.Background(), []string{"ok"}, nil); err == nil {
		t.Fatal("expected stdout pipe setup error")
	}
	newBrewExecCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stderr = io.Discard
		return cmd
	}
	if err := runBrew(context.Background(), []string{"ok"}, nil); err == nil {
		t.Fatal("expected stderr pipe setup error")
	}
}
