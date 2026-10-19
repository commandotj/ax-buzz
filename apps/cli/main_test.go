package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/commandotj/ax-buzz/modules/brew/brew"
	"github.com/commandotj/ax-buzz/modules/brew/data"
	"github.com/commandotj/ax-buzz/modules/protocol/protocol"
)

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

type failAfterWriter struct{ writes, allowed int }

func (w *failAfterWriter) Write(data []byte) (int, error) {
	if w.writes >= w.allowed {
		return 0, errors.New("write failed")
	}
	w.writes++
	return len(data), nil
}

type quitModel struct{}

func (quitModel) Init() tea.Cmd                         { return tea.Quit }
func (m quitModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (quitModel) View() string                          { return "" }

func TestRunMachine(t *testing.T) {
	var out bytes.Buffer
	if err := runMachine(&out, []string{"handshake"}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "Buzz CLI") || !strings.Contains(got, "\"event\":\"result\"") {
		t.Fatalf("unexpected handshake: %s", got)
	}
	for _, args := range [][]string{nil, {"bad"}, {"handshake", "extra"}} {
		if err := runMachine(io.Discard, args); err == nil {
			t.Errorf("runMachine(%v) accepted unsupported arguments", args)
		}
	}
	if err := runMachine(failedWriter{}, []string{"handshake"}); err == nil {
		t.Fatal("writer failure ignored")
	}
}

func TestParseStreamOperation(t *testing.T) {
	cases := []struct {
		args      []string
		name, pkg string
		cask      bool
		wantErr   bool
	}{
		{nil, "", "", false, true}, {[]string{"list"}, "list", "", false, false}, {[]string{"installed"}, "installed", "", false, false}, {[]string{"refresh"}, "refresh", "", false, false},
		{[]string{"install", "wget"}, "install", "wget", false, false}, {[]string{"uninstall", "app", "--cask"}, "uninstall", "app", true, false}, {[]string{"upgrade", "wget"}, "upgrade", "wget", false, false},
		{[]string{"list", "extra"}, "", "", false, true}, {[]string{"list", "--cask"}, "", "", false, true}, {[]string{"install"}, "", "", false, true}, {[]string{"invalid"}, "", "", false, true},
		{[]string{"install", "first", "second"}, "", "", false, true},
	}
	for _, tc := range cases {
		got, err := parseStreamOperation(tc.args)
		if (err != nil) != tc.wantErr {
			t.Errorf("parse(%v) err=%v", tc.args, err)
			continue
		}
		if err == nil && (got.name != tc.name || got.packageID != tc.pkg || got.isCask != tc.cask) {
			t.Errorf("parse(%v)=%+v", tc.args, got)
		}
	}
}

func TestPackageHelpersAndFinishStream(t *testing.T) {
	a := &data.Package{Name: "a", IsInstalled: true}
	b := &data.Package{Name: "b"}
	c := &data.Package{Name: "a", IsCask: true}
	if findPackage([]*data.Package{a, b, c}, "a", false) != a || findPackage([]*data.Package{a, b, c}, "a", true) != c || findPackage([]*data.Package{a}, "missing", false) != nil {
		t.Fatal("package lookup incorrect")
	}
	got := installedPackages([]*data.Package{a, b, c})
	if len(got) != 1 || got[0] != a {
		t.Fatalf("installed filter = %#v", got)
	}
	var out, diagnostics bytes.Buffer
	w := protocol.NewWriter(&out)
	err := finishStream(w, errors.New("failed operation"))
	if err == nil || diagnostics.Len() != 0 || !strings.Contains(out.String(), "\"success\":false") {
		t.Fatalf("stream failure was not reported: %v %q %q", err, out.String(), diagnostics.String())
	}
}

func TestMainMachineRoute(t *testing.T) {
	oldArgs, oldOut := os.Args, os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"buzz", "--machine", "handshake"}
	os.Stdout = w
	main()
	_ = w.Close()
	os.Args, os.Stdout = oldArgs, oldOut
	got, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("Buzz CLI")) {
		t.Fatalf("machine route output missing: %s", got)
	}
}

func TestMainErrorAndRunRoutes(t *testing.T) {
	oldArgs, oldExit := os.Args, processExit
	defer func() { os.Args, processExit = oldArgs, oldExit }()
	var exitCode int
	processExit = func(code int) { exitCode = code }
	os.Args = []string{"buzz", "--machine", "unsupported"}
	main()
	if exitCode != 1 {
		t.Fatal("main did not report failure")
	}
	var output, diagnostics bytes.Buffer
	if err := run([]string{"--machine", "handshake"}, &output, &diagnostics); err != nil || !strings.Contains(output.String(), "Buzz CLI") {
		t.Fatalf("machine dispatch: %v %q", err, output.String())
	}
	if err := run([]string{"--machine", "bad"}, &output, &diagnostics); err == nil || diagnostics.Len() != 0 {
		t.Fatal("machine error dispatch missing")
	}
	if err := run([]string{"--stream-json", "unknown"}, &output, &diagnostics); err == nil || diagnostics.Len() != 0 {
		t.Fatal("stream error dispatch missing")
	}
	*flagShowVersion = true
	if err := run([]string{"--version"}, &output, &diagnostics); err != nil || !strings.Contains(output.String(), strings.Fields(version)[0]) {
		t.Fatal("version route failed")
	}
	*flagShowVersion = false
	*flagShowHelp = true
	if err := run([]string{"--help"}, &output, &diagnostics); err != nil || diagnostics.Len() == 0 {
		t.Fatal("help route failed")
	}
	*flagShowHelp = false
	if err := run([]string{"--not-a-flag"}, &output, &diagnostics); err == nil {
		t.Fatal("invalid flag accepted")
	}
	oldStart := startProgram
	t.Cleanup(func() { startProgram = oldStart })
	startProgram = func(*tea.Program) error { return errors.New("program failed") }
	t.Setenv("BUZZ_LOG", t.TempDir()+"/buzz.log")
	if err := run(nil, &output, &diagnostics); err == nil || !strings.Contains(err.Error(), "program failed") {
		t.Fatal("interactive failure not returned")
	}
	startProgram = func(*tea.Program) error { return nil }
	if err := run(nil, &output, &diagnostics); err != nil {
		t.Fatalf("interactive route failed: %v", err)
	}
	if err := runProgram(tea.NewProgram(quitModel{}, tea.WithInput(strings.NewReader("")), tea.WithOutput(io.Discard), tea.WithoutRenderer())); err != nil {
		t.Fatalf("program runner failed: %v", err)
	}
	t.Setenv("BUZZ_LOG", t.TempDir()+"/missing"+"/buzz.log")
	if err := run(nil, &output, &diagnostics); err == nil || !strings.Contains(err.Error(), "failed to create log file") {
		t.Fatal("log creation failure not returned")
	}
}

func TestMainReportsMachineAndStreamErrorsOnce(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"machine", []string{"--machine", "bad"}, "unsupported machine command; expected handshake\n"},
		{"stream-json", []string{"--stream-json", "unknown"}, "unsupported operation \"unknown\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stderr, exitCode := captureMain(t, tc.args)
			if exitCode != 1 || stderr != tc.want {
				t.Fatalf("main stderr=%q exit=%d, want one line %q and exit 1", stderr, exitCode, tc.want)
			}
		})
	}
}

func captureMain(t *testing.T, args []string) (string, int) {
	t.Helper()
	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	oldArgs, oldStdout, oldStderr, oldExit := os.Args, os.Stdout, os.Stderr, processExit
	var exitCode int
	os.Args = append([]string{"buzz"}, args...)
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	processExit = func(code int) { exitCode = code }
	main()
	os.Args, os.Stdout, os.Stderr, processExit = oldArgs, oldStdout, oldStderr, oldExit
	if err := stdoutFile.Close(); err != nil {
		t.Fatal(err)
	}
	if err := stderrFile.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(content), exitCode
}

func TestRunStreamJSONInvalidArgs(t *testing.T) {
	var out, diagnostics bytes.Buffer
	err := runStreamJSON(&out, []string{"bogus"})
	if err == nil || !strings.Contains(out.String(), "\"success\":false") || diagnostics.Len() != 0 {
		t.Fatalf("invalid operation not returned as protocol error: %v", err)
	}
}

func TestRunStreamJSONLoadActionAndListPaths(t *testing.T) {
	oldLoad, oldAction := loadCatalog, performAction
	t.Cleanup(func() { loadCatalog, performAction = oldLoad, oldAction })
	packages := []*data.Package{
		{Name: "formula", InstallSupported: true},
		{Name: "cask", InstallSupported: true, IsCask: true, IsInstalled: true},
		{Name: "unsupported"},
	}
	loadCatalog = func(_ context.Context, force bool, progress func(string)) (*brew.CatalogSnapshot, error) {
		if force {
			progress("refreshing")
		}
		return &brew.CatalogSnapshot{Packages: packages}, nil
	}
	performAction = func(_ context.Context, action, name string, cask bool, progress func(string)) error {
		progress("action output")
		if name == "formula" {
			return errors.New("action failed")
		}
		return nil
	}
	for _, tc := range []struct {
		args     []string
		contains string
		wantErr  bool
	}{
		{[]string{"installed"}, `"count":1`, false},
		{[]string{"list"}, `"count":3`, false},
		{[]string{"refresh"}, `"operation":"refresh"`, false},
		{[]string{"install", "missing"}, `"package \"missing\" was not found"`, true},
		{[]string{"install", "unsupported"}, `not supported by Buzz`, true},
		{[]string{"upgrade", "formula"}, `not installed`, true},
	} {
		var output bytes.Buffer
		err := runStreamJSON(&output, tc.args)
		if (err != nil) != tc.wantErr || !strings.Contains(output.String(), tc.contains) {
			t.Errorf("%v => %v %q", tc.args, err, output.String())
		}
	}
	var output, diagnostics bytes.Buffer
	err := runStreamJSON(&output, []string{"install", "formula"})
	if err == nil || !strings.Contains(output.String(), `"success":false`) || diagnostics.Len() != 0 {
		t.Fatal("action error path not reported")
	}
	output.Reset()
	diagnostics.Reset()
	if err := runStreamJSON(&output, []string{"uninstall", "cask", "--cask"}); err != nil || !strings.Contains(output.String(), `"success":true`) {
		t.Fatalf("successful action failed: %v %s", err, output.String())
	}
	loadCatalog = func(context.Context, bool, func(string)) (*brew.CatalogSnapshot, error) {
		return nil, errors.New("catalog failed")
	}
	output.Reset()
	diagnostics.Reset()
	if err := runStreamJSON(&output, []string{"list"}); err == nil || !strings.Contains(output.String(), `"success":false`) {
		t.Fatal("catalog failure path missing")
	}
}

func TestRunStreamJSONChunkingAndWriterErrors(t *testing.T) {
	oldLoad := loadCatalog
	t.Cleanup(func() { loadCatalog = oldLoad })
	packages := make([]*data.Package, 205)
	for i := range packages {
		packages[i] = &data.Package{Name: fmt.Sprintf("pkg-%03d", i)}
	}
	loadCatalog = func(context.Context, bool, func(string)) (*brew.CatalogSnapshot, error) {
		return &brew.CatalogSnapshot{Packages: packages}, nil
	}
	var out bytes.Buffer
	if err := runStreamJSON(&out, []string{"list"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(out.String(), `"event":"data"`); got != 3 {
		t.Fatalf("data chunks = %d, want 3", got)
	}
	if err := runStreamJSON(failedWriter{}, []string{"list"}); err == nil || !strings.Contains(err.Error(), "write package data") {
		t.Fatalf("data writer failure = %v", err)
	}
	loadCatalog = func(context.Context, bool, func(string)) (*brew.CatalogSnapshot, error) {
		return &brew.CatalogSnapshot{Packages: packages[:1]}, nil
	}
	resultWriter := &failAfterWriter{allowed: 1}
	if err := runStreamJSON(resultWriter, []string{"list"}); err == nil || err.Error() != "write failed" {
		t.Fatalf("result writer failure = %v", err)
	}
	oldSignalContext := newSignalContext
	t.Cleanup(func() { newSignalContext = oldSignalContext })
	newSignalContext = func(context.Context, ...os.Signal) (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx, func() {}
	}
	for _, list := range [][]*data.Package{packages, nil} {
		loadCatalog = func(context.Context, bool, func(string)) (*brew.CatalogSnapshot, error) {
			return &brew.CatalogSnapshot{Packages: list}, nil
		}
		if err := runStreamJSON(io.Discard, []string{"list"}); !errors.Is(err, context.Canceled) {
			t.Errorf("canceled list returned %v", err)
		}
	}
	newSignalContext = func(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
		if len(signals) != 2 || signals[0] != os.Interrupt || signals[1] != syscall.SIGTERM {
			t.Errorf("signal context registered %v", signals)
		}
		return context.WithCancel(parent)
	}
}
