package brew

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/commandotj/ax-buzz/modules/brew/data"
)

func commandMessages(t *testing.T, msg tea.Msg) []tea.Msg {
	t.Helper()
	first, ok := msg.(CommandOutputMsg)
	if !ok {
		t.Fatalf("initial message = %#v", msg)
	}
	var got []tea.Msg
	for next := range first.Ch {
		got = append(got, next)
	}
	return got
}

func TestCommandExecutionBranches(t *testing.T) {
	oldExe, oldErr := brewExecutable, brewExecutableErr
	brewExecutable = "/bin/echo"
	brewExecutableErr = nil
	defer func() { brewExecutable, brewExecutableErr = oldExe, oldErr }()
	pkg := &data.Package{Name: "demo", InstallSupported: true}
	msg := execute(BrewCommandInstall, []*data.Package{pkg}, "install", "demo")()
	items := commandMessages(t, msg)
	if len(items) < 3 {
		t.Fatalf("streamed messages=%#v", items)
	}
	if _, ok := items[len(items)-1].(CommandFinishMsg); !ok {
		t.Fatalf("finish message=%#v", items[len(items)-1])
	}
	unsupported := &data.Package{Name: "pkg-app", InstallSupported: false}
	for _, command := range []BrewCommand{BrewCommandInstall, BrewCommandUninstall} {
		got := commandMessages(t, execute(command, []*data.Package{unsupported}, string(command), unsupported.Name)())
		if len(got) != 3 {
			t.Fatalf("unsupported %s messages=%#v", command, got)
		}
	}
	brewExecutableErr = fmt.Errorf("not installed")
	if got := commandMessages(t, execute(BrewCommandCleanup, nil, "cleanup")()); len(got) != 2 {
		t.Fatalf("missing brew messages=%#v", got)
	}
	brewExecutableErr = nil
	brewExecutable = fmt.Sprintf("%s/missing", t.TempDir())
	if got := commandMessages(t, execute(BrewCommandCleanup, nil, "cleanup")()); len(got) != 2 {
		t.Fatalf("start error messages=%#v", got)
	}
	brewExecutable = "/usr/bin/false"
	got := commandMessages(t, execute(BrewCommandCleanup, nil, "cleanup")())
	finish := got[len(got)-1].(CommandFinishMsg)
	if finish.Err == nil {
		t.Fatal("expected command failure")
	}
}

func TestCommandHelpersAndStateUpdates(t *testing.T) {
	pkg := &data.Package{Name: "demo", IsCask: true, InstallSupported: true, Dependencies: []string{"dep"}}
	_ = UpgradeAllPackages([]*data.Package{pkg})
	_ = UpgradePackage(pkg)
	_ = InstallPackage(pkg)
	_ = UninstallPackage(pkg)
	_ = PinPackage(pkg)
	_ = UnpinPackage(pkg)
	_ = Cleanup()
	start := startCommand()()
	if _, ok := start.(CommandStartMsg); !ok {
		t.Fatalf("start msg=%T", start)
	}
	ch := make(chan tea.Msg, 1)
	ch <- CommandStartMsg{}
	if _, ok := StreamOutput(ch)().(CommandStartMsg); !ok {
		t.Fatal("stream output")
	}
	feed := make(chan tea.Msg, 2)
	feedOutput(feed, io.NopCloser(strings.NewReader("line\n")))
	if got := <-feed; got.(CommandOutputMsg).Line != "line" {
		t.Fatalf("feed output=%#v", got)
	}
	pkg.IsInstalled = true
	pkg.IsOutdated = true
	allBrewPackages = []*data.Package{pkg, {Name: "dep"}}
	sort.Slice(allBrewPackages, func(i, j int) bool { return allBrewPackages[i].Name < allBrewPackages[j].Name })
	for _, action := range []BrewCommand{BrewCommandUpgradeAll, BrewCommandUpgrade, BrewCommandInstall, BrewCommandUninstall, BrewCommandPin, BrewCommandUnpin, BrewCommandCleanup} {
		UpdatePackageForAction(action, []*data.Package{pkg})
	}
	if pkg.IsPinned {
		t.Fatal("last unpin should clear pin")
	}
	pkg.MarkUninstalled()
	UpdatePackageForAction(BrewCommandInstall, []*data.Package{pkg})
	if dep := GetPackage("dep"); dep == nil {
		t.Fatal("dependency lookup")
	} else if !dep.IsInstalled {
		t.Fatal("missing dependency was not marked installed")
	}
	UpdatePackageForAction("unknown", []*data.Package{pkg})
}

func TestCommandPipeSetupErrors(t *testing.T) {
	oldFactory, oldExe, oldErr := newBrewExecCommand, brewExecutable, brewExecutableErr
	defer func() { newBrewExecCommand, brewExecutable, brewExecutableErr = oldFactory, oldExe, oldErr }()
	brewExecutable = "/bin/echo"
	brewExecutableErr = nil
	newBrewExecCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(name, args...)
		cmd.Stdout = io.Discard
		return cmd
	}
	if got := commandMessages(t, execute(BrewCommandCleanup, nil, "cleanup")()); len(got) != 2 || got[1].(CommandFinishMsg).Err == nil {
		t.Fatalf("stdout pipe error=%#v", got)
	}
	newBrewExecCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(name, args...)
		cmd.Stderr = io.Discard
		return cmd
	}
	if got := commandMessages(t, execute(BrewCommandCleanup, nil, "cleanup")()); len(got) != 2 || got[1].(CommandFinishMsg).Err == nil {
		t.Fatalf("stderr pipe error=%#v", got)
	}
}
