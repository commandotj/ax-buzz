package brew

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindBrewExecutableSearchOrder(t *testing.T) {
	oldLook, oldStat := brewLookPath, brewStat
	defer func() { brewLookPath, brewStat = oldLook, oldStat }()
	brewLookPath = func(string) (string, error) { return "", errors.New("not in path") }
	file, err := os.CreateTemp(t.TempDir(), "brew")
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	dirInfo, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	brewStat = func(path string) (os.FileInfo, error) {
		switch path {
		case "/opt/homebrew/bin/brew":
			return dirInfo, nil
		case "/usr/local/bin/brew":
			return nil, os.ErrNotExist
		default:
			return fileInfo, nil
		}
	}
	if got, err := findBrewExecutable(); err != nil || got != "/home/linuxbrew/.linuxbrew/bin/brew" {
		t.Fatalf("fallback executable=%q err=%v", got, err)
	}
	brewStat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	if _, err := findBrewExecutable(); err == nil {
		t.Fatal("expected missing executable error")
	}
	brewLookPath = func(string) (string, error) { return "/custom/brew", nil }
	if got, err := findBrewExecutable(); err != nil || got != "/custom/brew" {
		t.Fatalf("PATH executable=%q err=%v", got, err)
	}
}

func TestBrewPrefixAndPinnedPackageDiscovery(t *testing.T) {
	if got := detectBrewPrefix("/bin/echo", nil); got != "--prefix" {
		t.Fatalf("brew prefix=%q", got)
	}
	if got := detectBrewPrefix("/does/not/exist", nil); got != "" {
		t.Fatalf("failed prefix=%q", got)
	}
	if got := detectBrewPrefix("/bin/echo", errors.New("not found")); got != "" {
		t.Fatalf("missing executable prefix=%q", got)
	}
	if got := findPinnedPackages(""); len(got) != 0 {
		t.Fatalf("empty prefix pinned=%v", got)
	}
	if got := findPinnedPackages(filepath.Join(t.TempDir(), "missing")); len(got) != 0 {
		t.Fatalf("missing pins=%v", got)
	}
	prefix := t.TempDir()
	pinDir := filepath.Join(prefix, "var/homebrew/pinned")
	if err := os.MkdirAll(pinDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pinDir, "formula"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if got := findPinnedPackages(prefix); !got["formula"] {
		t.Fatalf("pins=%v", got)
	}
}

func TestInstallDiscoveryAndCaskReceipts(t *testing.T) {
	result := make(chan []*installInfo, 1)
	fetchInstalledPackages(filepath.Join(t.TempDir(), "missing"), func(string) *installInfo { return nil }, result)
	if got := <-result; len(got) != 0 {
		t.Fatalf("missing dir results=%v", got)
	}
	root := t.TempDir()
	for _, name := range []string{".hidden", "linked", "pkg"} {
		_ = os.Mkdir(filepath.Join(root, name), 0700)
	}
	_ = os.Symlink(filepath.Join(root, "pkg"), filepath.Join(root, "linked"))
	fetchInstalledPackages(root, func(path string) *installInfo {
		if filepath.Base(path) == "pkg" {
			return &installInfo{name: "pkg"}
		}
		return nil
	}, result)
	if got := <-result; len(got) != 1 || got[0].name != "pkg" {
		t.Fatalf("filtered results=%+v", got)
	}
	oldPrefix := brewPrefix
	brewPrefix = t.TempDir()
	t.Cleanup(func() { brewPrefix = oldPrefix })
	if err := os.Mkdir(filepath.Join(brewPrefix, "Cellar"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(brewPrefix, "Caskroom"), 0700); err != nil {
		t.Fatal(err)
	}
	fetchInstalledFormula(false, result)
	if len(<-result) != 0 {
		t.Fatal("empty formula directory should yield no packages")
	}
	fetchInstalledCask(false, result)
	if len(<-result) != 0 {
		t.Fatal("empty cask directory should yield no packages")
	}
	formulaRoot := filepath.Join(brewPrefix, "Cellar", "tool")
	formulaVersion := filepath.Join(formulaRoot, "1.0")
	if err := os.MkdirAll(formulaVersion, 0700); err != nil {
		t.Fatal(err)
	}
	formulaReceipt := `{"source":{"versions":{"stable":"1.0"}}}`
	if err := os.WriteFile(filepath.Join(formulaVersion, "INSTALL_RECEIPT.json"), []byte(formulaReceipt), 0600); err != nil {
		t.Fatal(err)
	}
	fetchInstalledFormula(true, result)
	if got := <-result; len(got) != 1 || got[0].name != "tool" {
		t.Fatalf("formula scan=%+v", got)
	}
	caskRoot := filepath.Join(brewPrefix, "Caskroom", "app")
	if err := os.MkdirAll(filepath.Join(caskRoot, "1.0"), 0700); err != nil {
		t.Fatal(err)
	}
	fetchInstalledCask(true, result)
	if got := <-result; len(got) != 1 || got[0].name != "app" {
		t.Fatalf("cask scan=%+v", got)
	}
	brewPrefix = ""
	fetchInstalledFormula(false, result)
	if len(<-result) != 0 {
		t.Fatal("no brew formula scan should be empty")
	}
	fetchInstalledCask(false, result)
	if len(<-result) != 0 {
		t.Fatal("no brew cask scan should be empty")
	}
	cask := filepath.Join(t.TempDir(), "test-cask")
	_ = os.MkdirAll(filepath.Join(cask, ".metadata"), 0700)
	_ = os.Mkdir(filepath.Join(cask, "1.0"), 0700)
	receipt := `{"installed_as_dependency":true,"time":42,"source":{"tap":"custom/tap","path":"cask.rb"}}`
	_ = os.WriteFile(filepath.Join(cask, ".metadata", "INSTALL_RECEIPT.json"), []byte(receipt), 0600)
	info := getCaskInstallInfo(true, cask)
	if info.version != "1.0" || info.tap != "custom/tap" || !info.asDep || info.path != "cask.rb" || info.timestamp != 42 {
		t.Fatalf("cask receipt=%+v", info)
	}
	if got := getCaskInstallInfo(false, filepath.Join(t.TempDir(), "absent")); got == nil {
		t.Fatal("missing cask still yields package info")
	}
	if got := fetchDirSize(filepath.Join(t.TempDir(), "not-found"), false); got != 0 {
		t.Fatalf("missing dir size=%d", got)
	}
	formulaEmpty := filepath.Join(t.TempDir(), "empty-formula")
	_ = os.Mkdir(formulaEmpty, 0700)
	if got := getFormulaInstallInfo(false, formulaEmpty); got == nil || got.version != "" {
		t.Fatalf("empty formula info=%+v", got)
	}
	if got := getFormulaInstallInfo(false, filepath.Join(t.TempDir(), "missing-formula")); got != nil {
		t.Fatalf("missing formula info=%+v", got)
	}
	hiddenOnly := filepath.Join(t.TempDir(), "hidden-formula")
	_ = os.Mkdir(hiddenOnly, 0700)
	_ = os.Mkdir(filepath.Join(hiddenOnly, ".hidden"), 0700)
	if got := getFormulaInstallInfo(false, hiddenOnly); got == nil || got.version != ".hidden" {
		t.Fatalf("hidden-only formula info=%+v", got)
	}
	badRevision := filepath.Join(t.TempDir(), "formula")
	badVersion := filepath.Join(badRevision, "1.0_bad")
	_ = os.MkdirAll(badVersion, 0700)
	_ = os.WriteFile(filepath.Join(badVersion, "INSTALL_RECEIPT.json"), []byte(`{"source":{"versions":{"stable":"1.0"}}}`), 0600)
	if got := getFormulaInstallInfo(false, badRevision); got == nil || got.revision != 0 || got.version != "1.0" {
		t.Fatalf("bad revision fallback=%+v", got)
	}
	malformed := t.TempDir()
	_ = os.WriteFile(filepath.Join(malformed, "INSTALL_RECEIPT.json"), []byte("{"), 0600)
	if parseInstallReceipt(malformed) != nil {
		t.Fatal("malformed receipt should be ignored")
	}
	dangling := filepath.Join(t.TempDir(), "dangling-cask")
	_ = os.Mkdir(dangling, 0700)
	_ = os.Symlink(filepath.Join(dangling, "absent"), filepath.Join(dangling, "1.0"))
	if got := getCaskInstallInfo(false, dangling); got == nil || got.version != "1.0" || got.timestamp != 0 {
		t.Fatalf("cask with dangling version=%+v", got)
	}
}
