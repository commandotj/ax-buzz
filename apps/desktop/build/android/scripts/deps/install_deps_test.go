package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMainSuccessAndFailure(t *testing.T) {
	t.Run("all dependencies available", func(t *testing.T) {
		setupAndroidSuccess(t, "printf 'Pixel_7\\n'\n")
		if code := run(); code != 0 {
			t.Fatalf("run() = %d, want 0", code)
		}
	})
	t.Run("create AVD when emulator list empty", func(t *testing.T) {
		setupAndroidSuccess(t, "exit 0\n")
		if code := run(); code != 0 {
			t.Fatalf("run() = %d, want 0", code)
		}
	})
	t.Run("missing dependencies", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, "Android", "Sdk"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("HOME", home)
		t.Setenv("ANDROID_HOME", "")
		t.Setenv("ANDROID_SDK_ROOT", "")
		t.Setenv("ANDROID_NDK_HOME", "")
		if code := run(); code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}
	})
	t.Run("main forwards exit code", func(t *testing.T) {
		oldExit := exitProcess
		defer func() { exitProcess = oldExit }()
		called := false
		exitProcess = func(code int) {
			called = true
			if code != 1 {
				t.Errorf("exit code = %d", code)
			}
		}
		t.Setenv("PATH", t.TempDir())
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ANDROID_HOME", "")
		t.Setenv("ANDROID_SDK_ROOT", "")
		t.Setenv("ANDROID_NDK_HOME", "")
		main()
		if !called {
			t.Fatal("main did not call exitProcess")
		}
	})
}

func setupAndroidSuccess(t *testing.T, emulatorBody string) {
	t.Helper()
	bin := t.TempDir()
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(goPath, filepath.Join(bin, "go")); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"adb": "exit 0\n", "emulator": emulatorBody, "java": "exit 0\n", "avdmanager": "exit 0\n",
	} {
		writeCommand(t, bin, name, body)
	}
	sdk := t.TempDir()
	if err := os.MkdirAll(filepath.Join(sdk, "ndk", "27.1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(sdk, "system-images", "android-35", "google_apis", androidABI()), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", sdk)
	t.Setenv("ANDROID_NDK_HOME", "")
	t.Setenv("TASK_FORCE_YES", "true")
}

func TestCheckCommand(t *testing.T) {
	bin := t.TempDir()
	writeCommand(t, bin, "ok", "exit 0\n")
	writeCommand(t, bin, "bad", "exit 1\n")
	t.Setenv("PATH", bin)
	if !checkCommand("ok") || checkCommand("bad") || checkCommand("missing") {
		t.Fatal("checkCommand returned wrong result")
	}
}

func TestOfferCreateAVD(t *testing.T) {
	t.Run("no image or manager", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		output := captureOutput(t, func() { offerCreateAVD(t.TempDir()) })
		if !strings.Contains(output, "No Android Virtual Devices found") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("decline installed image", func(t *testing.T) {
		sdk := t.TempDir()
		if err := os.MkdirAll(filepath.Join(sdk, "system-images", "android-35", "google_apis", androidABI()), 0o755); err != nil {
			t.Fatal(err)
		}
		bin := t.TempDir()
		writeCommand(t, bin, "avdmanager", "exit 0\n")
		t.Setenv("PATH", bin)
		t.Setenv("CI", "")
		t.Setenv("TASK_FORCE_YES", "")
		setInput(t, "n\n")
		output := captureOutput(t, func() { offerCreateAVD(sdk) })
		if !strings.Contains(output, "Skipping AVD creation") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("create installed image", func(t *testing.T) {
		sdk := t.TempDir()
		if err := os.MkdirAll(filepath.Join(sdk, "system-images", "android-35", "google_apis", androidABI()), 0o755); err != nil {
			t.Fatal(err)
		}
		bin := t.TempDir()
		writeCommand(t, bin, "avdmanager", "exit 0\n")
		t.Setenv("PATH", bin)
		t.Setenv("TASK_FORCE_YES", "true")
		output := captureOutput(t, func() { offerCreateAVD(sdk) })
		if !strings.Contains(output, "AVD created") {
			t.Fatalf("output = %s", output)
		}
	})
	t.Run("avdmanager failure", func(t *testing.T) {
		sdk := t.TempDir()
		if err := os.MkdirAll(filepath.Join(sdk, "system-images", "android-35", "google_apis", androidABI()), 0o755); err != nil {
			t.Fatal(err)
		}
		bin := t.TempDir()
		writeCommand(t, bin, "avdmanager", "echo failed >&2\nexit 1\n")
		t.Setenv("PATH", bin)
		t.Setenv("TASK_FORCE_YES", "true")
		output := captureOutput(t, func() { offerCreateAVD(sdk) })
		if !strings.Contains(output, "Failed to create AVD") {
			t.Fatalf("output = %s", output)
		}
	})
}

func TestFindAVDManager(t *testing.T) {
	bin := t.TempDir()
	pathAVD := writeCommand(t, bin, "avdmanager", "exit 0\n")
	t.Setenv("PATH", bin)
	if got := findAVDManager(t.TempDir()); got != pathAVD {
		t.Fatalf("PATH manager = %q, want %q", got, pathAVD)
	}
	t.Setenv("PATH", t.TempDir())
	sdk := t.TempDir()
	for _, version := range []string{"9.0", "11.0", "latest", "bad"} {
		writeCommand(t, filepath.Join(sdk, "cmdline-tools", version, "bin"), "avdmanager", "exit 0\n")
	}
	if got := findAVDManager(sdk); !strings.Contains(got, "latest") {
		t.Fatalf("latest manager = %q", got)
	}
	if got := findAVDManager(t.TempDir()); got != "" {
		t.Fatalf("missing manager = %q", got)
	}
	if got := findAVDManager(""); got != "" {
		t.Fatalf("empty SDK manager = %q", got)
	}
}

func TestAndroidSDKRelativePath(t *testing.T) {
	if got := androidSDKRelativePath("darwin"); got != "Library/Android/sdk" {
		t.Fatalf("macOS path = %q", got)
	}
	if got := androidSDKRelativePath("linux"); got != "Android/Sdk" {
		t.Fatalf("Linux path = %q", got)
	}
}

func TestPromptUser(t *testing.T) {
	t.Setenv("CI", "1")
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
		setInput(t, input)
		if got := promptUser("question"); got != want {
			t.Errorf("promptUser(%q) = %t, want %t", input, got, want)
		}
	}
}

func TestFindHighestInstalledImageParsing(t *testing.T) {
	// Numeric version and malformed preview coverage comes from directory names.
	sdk := t.TempDir()
	for _, api := range []string{"9", "35", "35-ext14", "abc"} {
		if err := os.MkdirAll(filepath.Join(sdk, "system-images", "android-"+api, "google_apis", androidABI()), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	bin := t.TempDir()
	writeCommand(t, bin, "avdmanager", "exit 0\n")
	t.Setenv("PATH", bin)
	t.Setenv("TASK_FORCE_YES", "")
	setInput(t, "n\n")
	output := captureOutput(t, func() { offerCreateAVD(sdk) })
	if !strings.Contains(output, "android-35") || strings.Contains(output, "android-9;google") {
		t.Fatalf("selected image output = %s", output)
	}
}

func androidABI() string {
	if runtime.GOARCH == "arm64" {
		return "arm64-v8a"
	}
	return "x86_64"
}

func writeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
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
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	return string(data)
}

func setInput(t *testing.T, input string) {
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
