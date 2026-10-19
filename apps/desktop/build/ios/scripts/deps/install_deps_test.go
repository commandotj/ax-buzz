package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDependencyChecks(t *testing.T) {
	t.Run("fully configured", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 (26.0) - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "iPhone 17 (available)\n", "exit 0\n")
		if code := run(); code != 0 {
			t.Fatalf("run() = %d, want 0", code)
		}
	})
	t.Run("missing tools", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("main forwards result", func(t *testing.T) {
		old := exitProcess
		defer func() { exitProcess = old }()
		called := false
		exitProcess = func(code int) {
			called = true
			if code != 1 {
				t.Errorf("exit code = %d", code)
			}
		}
		t.Setenv("PATH", t.TempDir())
		main()
		if !called {
			t.Fatal("main did not call exitProcess")
		}
	})
	t.Run("runtime absent and user declines device", func(t *testing.T) {
		setupIOS(t, true, true, "", "", "exit 0\n")
		withInput(t, "n\n")
		output := captureOutput(t, func() { _ = run() })
		if !strings.Contains(output, "No iOS runtime found") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("runtime present and user declines device", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "", "exit 0\n")
		withInput(t, "n\n")
		output := captureOutput(t, func() { _ = run() })
		if !strings.Contains(output, "Skipping simulator creation") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("simulator creation fails", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "", "exit 1\n")
		setenvYes(t)
		output := captureOutput(t, func() { _ = run() })
		if !strings.Contains(output, "Failed to create simulator") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("simulator creation succeeds", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "", "exit 0\n")
		setenvYes(t)
		output := captureOutput(t, func() { _ = run() })
		if !strings.Contains(output, "iPhone 15 Pro simulator created") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("devices command fails", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "", "exit 0\n")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "xcrun", `case "$*" in "simctl list devices") exit 1;; *) exit 0;; esac`)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("device listing output failure", func(t *testing.T) {
		setupIOS(t, true, true, "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n", "", "exit 0\n")
		counter := filepath.Join(t.TempDir(), "count")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "xcrun", `case "$*" in
"simctl list devices") n=0; [ -f "$COUNTER" ] && n=$(cat "$COUNTER"); n=$((n+1)); echo "$n" > "$COUNTER"; [ "$n" -eq 1 ] && exit 0; exit 1;;
*) exit 0;; esac`)
		t.Setenv("COUNTER", counter)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("invalid developer and incomplete SDK paths", func(t *testing.T) {
		setupIOS(t, true, false, "", "", "exit 0\n")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "xcode-select", `echo /Library/Developer/CommandLineTools`)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("xcode command output fails", func(t *testing.T) {
		setupIOS(t, true, true, "", "iPhone available\n", "exit 0\n")
		counter := filepath.Join(t.TempDir(), "xcode-count")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "xcodebuild", `n=0; [ -f "$XCODE_COUNTER" ] && n=$(cat "$XCODE_COUNTER"); n=$((n+1)); echo "$n" > "$XCODE_COUNTER"; [ "$n" -eq 1 ] && echo Xcode && exit 0; exit 1`)
		t.Setenv("XCODE_COUNTER", counter)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("runtime command output fails", func(t *testing.T) {
		setupIOS(t, true, true, "", "iPhone available\n", "exit 0\n")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "xcrun", `case "$*" in "simctl list runtimes") exit 1;; "simctl list devices") echo iPhone;; *) exit 0;; esac`)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("install command success", func(t *testing.T) {
		setupIOS(t, false, true, "", "", "exit 0\n")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "sudo", "exit 0")
		outputCommand = func(name string, args ...string) ([]byte, error) {
			if name == "xcode-select" {
				return nil, errors.New("developer path unavailable")
			}
			return defaultOutputCommand(name, args...)
		}
		setenvYes(t)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("install command fails", func(t *testing.T) {
		setupIOS(t, false, true, "", "", "exit 0\n")
		writeCommand(t, filepath.Join(os.Getenv("PATH")), "sudo", "exit 1")
		outputCommand = func(name string, args ...string) ([]byte, error) {
			if name == "xcode-select" {
				return nil, errors.New("developer path unavailable")
			}
			return defaultOutputCommand(name, args...)
		}
		setenvYes(t)
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
}

func TestCheckCommandAndPrompt(t *testing.T) {
	if checkCommand(nil) {
		t.Fatal("empty command succeeded")
	}
	bin := t.TempDir()
	writeCommand(t, bin, "ok", "exit 0")
	writeCommand(t, bin, "bad", "exit 1")
	t.Setenv("PATH", bin)
	if !checkCommand([]string{"ok"}) || checkCommand([]string{"bad"}) {
		t.Fatal("checkCommand result mismatch")
	}
	t.Setenv("CI", "true")
	if !promptUser("question") {
		t.Fatal("CI prompt not accepted")
	}
	t.Setenv("CI", "")
	t.Setenv("TASK_FORCE_YES", "true")
	if !promptUser("question") {
		t.Fatal("forced prompt not accepted")
	}
	t.Setenv("TASK_FORCE_YES", "")
	for input, want := range map[string]bool{"y\n": true, "YES\n": true, "n\n": false, "": false} {
		withInput(t, input)
		if got := promptUser("question"); got != want {
			t.Errorf("promptUser(%q) = %t, want %t", input, got, want)
		}
	}
}

func TestRunExternalCommandFailures(t *testing.T) {
	baseOutput, baseRun := outputCommand, runCommand
	defer func() { outputCommand, runCommand = baseOutput, baseRun }()
	tests := []struct {
		name   string
		modify func(t *testing.T)
		output func(string, ...string) ([]byte, error)
	}{
		{name: "xcode version command error", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcodebuild" {
				return nil, errors.New("xcodebuild error")
			}
			return baseOutput(name, args...)
		}},
		{name: "empty xcode version", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcodebuild" {
				return nil, nil
			}
			return baseOutput(name, args...)
		}},
		{name: "developer path missing", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcode-select" {
				return []byte(filepath.Join(t.TempDir(), "missing")), nil
			}
			return baseOutput(name, args...)
		}},
		{name: "developer path points at command line tools", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcode-select" {
				return []byte(t.TempDir()), nil
			}
			return baseOutput(name, args...)
		}},
		{name: "sdk path command error", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcrun" && len(args) > 0 && args[len(args)-1] == "--show-sdk-path" {
				return nil, errors.New("sdk path error")
			}
			return baseOutput(name, args...)
		}},
		{name: "runtime listing command error", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcrun" && len(args) == 3 && args[0] == "simctl" && args[2] == "runtimes" {
				return nil, errors.New("runtime error")
			}
			return baseOutput(name, args...)
		}},
		{name: "device listing output command error", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcrun" && len(args) == 3 && args[0] == "simctl" && args[2] == "devices" {
				return nil, errors.New("devices error")
			}
			return baseOutput(name, args...)
		}},
		{name: "runtime creation listing error", modify: func(t *testing.T) {}, output: func(name string, args ...string) ([]byte, error) {
			if name == "xcrun" && len(args) == 3 && args[0] == "simctl" && args[2] == "runtimes" {
				runtimeOutputCall++
				if runtimeOutputCall == 2 {
					return nil, errors.New("runtime create list error")
				}
			}
			return baseOutput(name, args...)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupIOS(t, true, true, "", "", "exit 0\n")
			tc.modify(t)
			runtimeOutputCall = 0
			outputCommand = tc.output
			if tc.name == "runtime creation listing error" {
				t.Setenv("IOS_RUNTIMES", "iOS 26.0 - com.apple.CoreSimulator.SimRuntime.iOS-26-0\n")
			}
			_ = run()
		})
	}
	t.Run("installer run errors", func(t *testing.T) {
		setupIOS(t, false, true, "", "", "exit 0\n")
		setenvYes(t)
		runCommand = func(cmd *exec.Cmd) error {
			if filepath.Base(cmd.Path) == "sudo" {
				return errors.New("sudo failed")
			}
			return baseRun(cmd)
		}
		_ = run()
	})
}

var runtimeOutputCall int

func setupIOS(t *testing.T, xcode, sdk bool, runtimes, devices, createResult string) {
	t.Helper()
	bin := t.TempDir()
	developer := filepath.Join(t.TempDir(), "Xcode.app", "Contents", "Developer")
	if err := os.MkdirAll(developer, 0o755); err != nil {
		t.Fatal(err)
	}
	sdkPath := t.TempDir()
	if sdk {
		if err := os.MkdirAll(filepath.Join(sdkPath, "System", "Library", "Frameworks", "UIKit.framework"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeCommand(t, bin, "xcodebuild", `echo 'Xcode 26.0'`)
	if !xcode {
		writeCommand(t, bin, "xcodebuild", "exit 1")
	}
	writeCommand(t, bin, "xcode-select", `printf '%s\n' "$DEVELOPER_PATH"`)
	writeCommand(t, bin, "xcrun", `case "$*" in
"--sdk iphonesimulator --show-sdk-path") printf '%s\n' "$IOS_SDK";;
"--sdk iphonesimulator --show-sdk-version") echo 26.0;;
"simctl help") exit 0;;
"simctl list runtimes") printf '%s' "$IOS_RUNTIMES";;
"simctl list devices") printf '%s' "$IOS_DEVICES";;
"simctl create"*) `+createResult+`;;
*) exit 1;; esac`)
	t.Setenv("PATH", bin)
	t.Setenv("DEVELOPER_PATH", developer)
	t.Setenv("IOS_SDK", sdkPath)
	t.Setenv("IOS_RUNTIMES", runtimes)
	t.Setenv("IOS_DEVICES", devices)
	t.Setenv("TASK_FORCE_YES", "")
	t.Setenv("CI", "")
}

func setenvYes(t *testing.T) { t.Helper(); t.Setenv("TASK_FORCE_YES", "true") }

func writeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func withInput(t *testing.T, input string) {
	t.Helper()
	old := os.Stdin
	file, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	os.Stdin = file
	t.Cleanup(func() { os.Stdin = old; _ = file.Close() })
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	file, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = file
	fn()
	os.Stdout = old
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	return string(output)
}
