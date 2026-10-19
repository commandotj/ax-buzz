package brew

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/commandotj/ax-buzz/modules/brew/data"
	"github.com/commandotj/ax-buzz/modules/brew/loading"
	"github.com/commandotj/ax-buzz/modules/brew/util"
	"github.com/commandotj/ax-buzz/modules/gh/gh"
)

func TestCatalogTransformAndQueries(t *testing.T) {
	assertPackageDomainCoverage(t)
	oldRelease := *flagFetchReleaseInfo
	*flagFetchReleaseInfo = false
	t.Cleanup(func() { *flagFetchReleaseInfo = oldRelease })
	formula := &apiFormula{Name: "alpha", Aliases: []string{"a"}, Tap: coreTap, Desc: "formula", Dependencies: []string{"dep"}, BuildDependencies: []string{"tool"}, Conflicts: []string{"bad"}, Revision: 2}
	formula.Versions.Stable = "2"
	formula.Urls.Stable.Url = "https://x/alpha-2.tar.gz"
	formula.Urls.Head.Url = "head"
	formula.Homepage = "https://x"
	formula.License = "MIT"
	cask := &apiCask{Name: "zeta", Tap: caskTap, Desc: "cask", Version: "1", Url: "https://x/zeta.dmg"}
	cask.Dependencies.Formulae = []string{"alpha"}
	cask.Dependencies.Casks = []string{"other"}
	cask.Conflicts.Formulae = []string{"bad"}
	formulaInfo := &installInfo{name: "alpha", version: "1", revision: 1, pinned: true, asDep: true, size: 2048, timestamp: 42}
	caskInfo := &installInfo{name: "zeta", version: "old", timestamp: 43}
	packages := processAllData([]*apiFormula{formula}, []*apiCask{cask}, apiFormulaAnalytics{Items: []struct {
		Name  string `json:"formula"`
		Count string `json:"count"`
	}{{"alpha", "1,234"}}}, apiCaskAnalytics{Items: []struct {
		Name  string `json:"cask"`
		Count string `json:"count"`
	}{{"zeta", "bad"}}}, []*installInfo{formulaInfo}, []*installInfo{caskInfo})
	if len(packages) != 2 || packages[0].Name != "alpha" || packages[0].Installs90d != 1234 || !packages[0].IsOutdated || !packages[0].IsPinned || packages[0].Dependents[0] != "zeta" {
		t.Fatalf("packages=%+v", packages)
	}
	if packages[1].Installs90d != 0 || !packages[1].IsOutdated || !packages[1].IsCask {
		t.Fatalf("cask=%+v", packages[1])
	}
	if parseInstallCount("1,234") != 1234 || parseInstallCount("bad") != 0 {
		t.Fatal("install count parse failed")
	}
	if got := mapFormulaeInstalls(apiFormulaAnalytics{}); len(got) != 0 {
		t.Fatalf("empty formula installs=%v", got)
	}
	if got := mapCaskInstalls(apiCaskAnalytics{}); len(got) != 0 {
		t.Fatalf("empty cask installs=%v", got)
	}
	if got := mapInstallInfo([]*installInfo{{name: "x"}}); got["x"] == nil {
		t.Fatalf("install map=%v", got)
	}
	if p := packageFromFormula(formula, 1, nil); !p.InstallSupported || p.IsInstalled {
		t.Fatalf("formula=%+v", p)
	}
	if p := packageFromCask(cask, 1, nil); !p.InstallSupported || p.IsInstalled {
		t.Fatalf("cask=%+v", p)
	}
	cask.Url = "https://x/a.pkg?download=1"
	if isInstallSupported(cask.Url) {
		t.Fatal("pkg cask must be unsupported")
	}
	cask.Url = "https://x/a.dmg"
	if !isInstallSupported(cask.Url) {
		t.Fatal("dmg cask must be supported")
	}
	allBrewPackages = packages
	if GetPackage("alpha") == nil || GetPackage("missing") != nil {
		t.Fatal("package binary search")
	}
	if len(GetOutdatedPackages()) != 2 {
		t.Fatalf("outdated=%v", GetOutdatedPackages())
	}
	installed := &data.Package{Name: "dep", IsInstalled: true}
	missing := &data.Package{Name: "root", Dependencies: []string{"dep"}}
	packages = append(packages, installed, missing)
	sort.Slice(packages, func(i, j int) bool { return packages[i].Name < packages[j].Name })
	allBrewPackages = packages
	if got := GetRecursiveMissingDeps("root"); len(got) != 1 || got[0] != "dep" {
		t.Fatalf("missing deps=%v", got)
	}
	if got := GetRecursiveMissingDeps("dep"); len(got) != 0 {
		t.Fatalf("installed deps=%v", got)
	}
	root := GetPackage("alpha")
	root.IsInstalled = true
	root.Dependents = []string{"zeta"}
	GetPackage("zeta").IsInstalled = true
	GetPackage("zeta").Dependents = nil
	if got := GetRecursiveInstalledDependents("alpha"); len(got) < 1 {
		t.Fatalf("installed dependents=%v", got)
	}
	GetPackage("root").IsInstalled = false
	if got := GetRecursiveInstalledDependents("root"); len(got) != 0 {
		t.Fatalf("uninstalled dependents=%v", got)
	}
}

func assertPackageDomainCoverage(t *testing.T) {
	t.Helper()
	p := &data.Package{Name: "demo", Aliases: []string{"d"}, Desc: "utility", Tap: "user/tap", Homepage: "https://example.test", Version: "2.0", InstalledVersion: "1.0", Revision: 1, InstalledRevision: 1}
	_ = p.Symbol()
	p.IsCask = true
	_ = p.Symbol()
	p.IsCask = false
	for _, status := range []func(*data.Package){func(x *data.Package) { x.IsDisabled = true }, func(x *data.Package) { x.IsDisabled = false; x.IsDeprecated = true }, func(x *data.Package) { x.IsDeprecated = false; x.IsPinned = true }, func(x *data.Package) { x.IsPinned = false; x.IsOutdated = true }, func(x *data.Package) { x.IsOutdated = false; x.InstalledAsDependency = true }, func(x *data.Package) { x.InstalledAsDependency = false; x.IsInstalled = true }, func(x *data.Package) { x.IsInstalled = false }} {
		status(p)
		_ = p.Status()
	}
	_ = p.BrewUrl()
	p.IsCask = true
	_ = p.BrewUrl()
	p.IsCask = false
	for _, state := range []struct{ outdated, pinned bool }{{false, false}, {true, false}, {false, true}} {
		p.IsOutdated, p.IsPinned = state.outdated, state.pinned
		_ = p.ShortVersion()
		_ = p.LongVersion()
	}
	p.Revision = 0
	p.InstalledRevision = 0
	p.IsOutdated = false
	p.IsPinned = false
	_ = p.ShortVersion()
	_ = p.LongVersion()
	p.IsPinned = true
	_ = p.ShortVersion()
	_ = p.LongVersion()
	_ = p.MatchKeywords([]string{"n:d", "d:util", "t:user", "h:example"})
	_ = p.MatchKeywords([]string{"-demo"})
	p.Aliases = []string{"alias-only"}
	_ = p.MatchKeywords([]string{"n:alias-only"})
	_ = p.MatchKeywords([]string{"n:absent"})
	_ = p.MatchKeywords([]string{"absent"})
	_ = p.MatchKeywords(nil)
	p.MarkInstalledAsDep()
	p.MarkUninstalled()
	p.MarkPinned()
	p.MarkUnpinned()
	_ = util.Sort([]string{"b", "a"})
	_ = util.SortAndUniq([]string{"b", "a", "a"})
	for _, size := range []int64{0, 1, 1024, 1536, 1048576, 1572864} {
		_ = util.FormatSize(size)
	}
	lp := loading.NewLoadingProgress()
	task := make(chan int)
	lp.AddTask(task, "task")
	lp.MarkCompleted(task)
	_ = lp.String(" done")
	lp.Reset()
	_ = lp.String(" done")
}

func TestCustomTapMergeAndReleaseFetch(t *testing.T) {
	oldRelease, oldFetch := *flagFetchReleaseInfo, fetchGithubReleaseInfo
	*flagFetchReleaseInfo = true
	called := make(chan struct{}, 2)
	fetchGithubReleaseInfo = func([]string, string) *gh.ReleaseInfo { called <- struct{}{}; return nil }
	t.Cleanup(func() { *flagFetchReleaseInfo = oldRelease; fetchGithubReleaseInfo = oldFetch })
	makeTap := func(name, body string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), name+".rb")
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	formulaPath := makeTap("formula", "version \"1\"\ndesc \"formula\"\nhomepage \"https://f\"\ndepends_on \"base\"")
	caskPath := makeTap("cask", "version \"2\"\ndesc \"cask\"\nhomepage \"https://c\"\nurl \"https://x/app.pkg\"")
	formulaInfos := []*installInfo{{name: "third-formula", tap: "user/formula", path: formulaPath, version: "1", timestamp: 1}, {name: "broken", tap: "user/broken", path: filepath.Join(t.TempDir(), "absent")}, {name: "core", tap: coreTap}}
	caskInfos := []*installInfo{{name: "third-cask", tap: "user/cask", path: caskPath, version: "2", timestamp: 2}, {name: "broken-cask", tap: "user/cask", path: filepath.Join(t.TempDir(), "absent")}, {name: "standard", tap: caskTap}}
	formulas := []*apiFormula{{Name: "base", Tap: coreTap, Desc: "base", Dependencies: []string{"leaf"}}, {Name: "unused", Tap: coreTap, Desc: "unused"}}
	casks := []*apiCask{{Name: "dependent", Tap: caskTap, Version: "1", Url: "https://x/app.dmg"}, {Name: "cask-child", Tap: caskTap, Version: "1", Url: "https://x/child.dmg"}}
	casks[1].Dependencies.Casks = []string{"dependent"}
	packages := processAllData(formulas, casks, apiFormulaAnalytics{}, apiCaskAnalytics{}, formulaInfos, caskInfos)
	for range 2 {
		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatal("release fetch goroutine did not run")
		}
	}
	byName := map[string]*data.Package{}
	for _, pkg := range packages {
		byName[pkg.Name] = pkg
	}
	if !byName["third-formula"].InstallSupported || byName["base"].Dependents[0] != "third-formula" {
		t.Fatalf("custom formula=%+v", byName["third-formula"])
	}
	if byName["third-cask"].InstallSupported {
		t.Fatalf("pkg custom cask should be unsupported: %+v", byName["third-cask"])
	}
	if byName["dependent"].Dependents[0] != "cask-child" {
		t.Fatalf("cask dependents=%v", byName["dependent"].Dependents)
	}
	if byName["base"].Dependents[0] != "third-formula" && byName["base"].Dependents[0] != "dependent" {
		t.Fatalf("formula dependents=%v", byName["base"].Dependents)
	}
	inst := updateInstallInfo(&data.Package{Name: "same", Version: "1", Revision: 2}, &installInfo{version: "1", revision: 2})
	if inst.IsOutdated {
		t.Fatalf("matching install marked outdated: %+v", inst)
	}
	auto := updateInstallInfo(&data.Package{Name: "auto", Version: "2", IsCask: true, AutoUpdate: true}, &installInfo{version: "1"})
	if auto.IsOutdated || auto.InstalledVersion != "2" {
		t.Fatalf("auto-updated cask=%+v", auto)
	}
}

func TestLoadDataWithCache(t *testing.T) {
	oldCache, oldPrefix, oldInvalidate := taproomCacheDir, brewPrefix, *flagInvalidateCache
	taproomCacheDir = t.TempDir()
	brewPrefix = ""
	*flagInvalidateCache = false
	t.Cleanup(func() { taproomCacheDir = oldCache; brewPrefix = oldPrefix; *flagInvalidateCache = oldInvalidate })
	write := func(name string, value any) {
		t.Helper()
		body, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(taproomCacheDir, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	formulaPayload, _ := json.Marshal([]*apiFormula{{Name: "cached-formula", Tap: coreTap}})
	caskPayload, _ := json.Marshal([]*apiCask{})
	write(formulaJwsJson, jwsJson{Payload: string(formulaPayload)})
	write(caskJwsJson, jwsJson{Payload: string(caskPayload)})
	write(formulaAnalyticsJson, apiFormulaAnalytics{})
	write(caskAnalyticsJson, apiCaskAnalytics{})
	for _, analytics := range []bool{false, true} {
		cmd := LoadData(analytics, false, loading.NewLoadingProgress())
		msg := cmd()
		loaded, ok := msg.(DataLoadedMsg)
		if !ok || len(loaded.Packages) != 1 || loaded.Packages[0].Name != "cached-formula" {
			t.Fatalf("LoadData(%v) = %#v", analytics, msg)
		}
	}
	if err := os.WriteFile(filepath.Join(taproomCacheDir, formulaJwsJson), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if msg := LoadData(false, false, loading.NewLoadingProgress())(); msg == nil {
		t.Fatal("loading failure should return a message")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ctx.Err() == nil {
		t.Fatal("test setup")
	}
}

func TestLoadPackagesAndCacheAge(t *testing.T) {
	oldCache, oldPrefix := taproomCacheDir, brewPrefix
	taproomCacheDir, brewPrefix = t.TempDir(), t.TempDir()
	t.Cleanup(func() { taproomCacheDir, brewPrefix = oldCache, oldPrefix })
	for _, dir := range []string{"Cellar", "Caskroom"} {
		if err := os.Mkdir(filepath.Join(brewPrefix, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{formulaJwsJson, caskJwsJson} {
		if err := os.WriteFile(filepath.Join(taproomCacheDir, name), []byte(`{"payload":"[]"}`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	progress := 0
	snapshot, err := LoadPackages(context.Background(), false, func(string) { progress++ })
	if err != nil || snapshot == nil || len(snapshot.Packages) != 0 || progress != 3 || snapshot.Stale {
		t.Fatalf("snapshot=%+v progress=%d err=%v", snapshot, progress, err)
	}
	old := time.Now().Add(-urlCacheTtl - time.Minute)
	for _, name := range []string{formulaJwsJson, caskJwsJson} {
		if err := os.Chtimes(filepath.Join(taproomCacheDir, name), old, old); err != nil {
			t.Fatal(err)
		}
	}
	if age, ok := catalogCacheAge(); !ok || age < urlCacheTtl {
		t.Fatalf("old cache age=%s ok=%v", age, ok)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) {
		return httpResponse(503, io.NopCloser(strings.NewReader("offline"))), nil
	}))
	if _, err := LoadPackages(ctx, true, nil); err == nil {
		t.Fatal("canceled/failed request should fail")
	}
	if err := os.WriteFile(filepath.Join(taproomCacheDir, caskJwsJson), []byte(`{"payload":"[]"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(taproomCacheDir, caskJwsJson)); err != nil {
		t.Fatal(err)
	}
	if _, ok := catalogCacheAge(); ok {
		t.Fatal("partial cache should have no age")
	}
	if err := os.WriteFile(filepath.Join(taproomCacheDir, caskJwsJson), []byte(`{"payload":"[]"}`), 0600); err != nil {
		t.Fatal(err)
	}
	oldReader := readCatalogCacheAge
	readCatalogCacheAge = func() (time.Duration, bool) { return 0, false }
	t.Cleanup(func() { readCatalogCacheAge = oldReader })
	if _, err := LoadPackages(context.Background(), false, nil); err == nil {
		t.Fatal("unavailable cache timestamps should fail")
	}
	oldFormula, oldCask, oldInstalledFormula, oldInstalledCask := fetchCatalogFormula, fetchCatalogCask, fetchCatalogInstalledFormula, fetchCatalogInstalledCask
	release := make(chan struct{})
	fetchCatalogFormula = func(context.Context, chan []*apiFormula, chan error, bool) { <-release }
	fetchCatalogCask = func(context.Context, chan []*apiCask, chan error, bool) { <-release }
	fetchCatalogInstalledFormula = func(bool, chan []*installInfo) { <-release }
	fetchCatalogInstalledCask = func(bool, chan []*installInfo) { <-release }
	fetchCatalogFormula = func(_ context.Context, _ chan []*apiFormula, errCh chan error, _ bool) {
		errCh <- errors.New("catalog fetch failed")
	}
	if _, err := LoadPackages(context.Background(), false, nil); err == nil {
		t.Fatal("catalog fetch error should return")
	}
	fetchCatalogFormula = func(context.Context, chan []*apiFormula, chan error, bool) { <-release }
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	if _, err := LoadPackages(ctx2, false, nil); err == nil {
		t.Fatal("canceled load should return context error")
	}
	close(release)
	fetchCatalogFormula, fetchCatalogCask, fetchCatalogInstalledFormula, fetchCatalogInstalledCask = oldFormula, oldCask, oldInstalledFormula, oldInstalledCask
}
