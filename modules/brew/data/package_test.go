package data

import (
	"testing"
)

func TestPackagePresentation(t *testing.T) {
	pkg := Package{Name: "demo", Version: "2.1", InstalledVersion: "1.9", Revision: 2, InstalledRevision: 1}
	if got := pkg.Symbol(); got != formulaSymbol {
		t.Fatalf("formula Symbol() = %q", got)
	}
	pkg.IsCask = true
	if got := pkg.Symbol(); got != caskSymbol {
		t.Fatalf("cask Symbol() = %q", got)
	}
	pkg.IsCask = false

	states := []struct {
		mutate func(*Package)
		want   string
	}{
		{func(p *Package) { p.IsDisabled = true }, statusDisabled},
		{func(p *Package) { p.IsDisabled = false; p.IsDeprecated = true }, statusDeprecated},
		{func(p *Package) { p.IsDeprecated = false; p.IsPinned = true }, statusPinned},
		{func(p *Package) { p.IsPinned = false; p.IsOutdated = true }, statusOutdated},
		{func(p *Package) { p.IsOutdated = false; p.InstalledAsDependency = true }, statusInstalledAsDep},
		{func(p *Package) { p.InstalledAsDependency = false; p.IsInstalled = true }, statusInstalled},
		{func(p *Package) { p.IsInstalled = false }, statusUninstalled},
	}
	for _, state := range states {
		state.mutate(&pkg)
		if got := pkg.Status(); got != state.want {
			t.Errorf("Status() = %q, want %q", got, state.want)
		}
	}
	if got := pkg.BrewUrl(); got != "https://formulae.brew.sh/formula/demo" {
		t.Fatalf("formula URL = %q", got)
	}
	pkg.IsCask = true
	if got := pkg.BrewUrl(); got != "https://formulae.brew.sh/cask/demo" {
		t.Fatalf("cask URL = %q", got)
	}

	pkg.IsCask, pkg.IsOutdated, pkg.IsPinned = false, false, false
	pkg.Version, pkg.Revision = "2.1", 2
	if got := pkg.ShortVersion(); got != "2.1_2" {
		t.Errorf("ShortVersion() = %q", got)
	}
	if got := pkg.LongVersion(); got != "2.1_2" {
		t.Errorf("LongVersion() = %q", got)
	}
	pkg.IsOutdated = true
	if got := pkg.ShortVersion(); got != "2.1_2 (New)" {
		t.Errorf("outdated short = %q", got)
	}
	if got := pkg.LongVersion(); got != "1.9_1 -> 2.1_2" {
		t.Errorf("outdated long = %q", got)
	}
	pkg.IsOutdated, pkg.IsPinned = false, true
	if got := pkg.ShortVersion(); got != "1.9_1 (Pin)" {
		t.Errorf("pinned short = %q", got)
	}
	if got := pkg.LongVersion(); got != "1.9_1 (Pinned)" {
		t.Errorf("pinned long = %q", got)
	}
	pkg.Revision, pkg.InstalledRevision = 0, 0
	if pkg.versionWithRev() != "2.1" || pkg.installedVersionWithRev() != "1.9" {
		t.Fatal("zero revisions should omit suffix")
	}
}

func TestPackageStateMutations(t *testing.T) {
	pkg := Package{Version: "3.0", IsOutdated: true}
	pkg.MarkInstalled()
	if !pkg.IsInstalled || pkg.IsOutdated || pkg.InstalledVersion != "3.0" || pkg.InstalledDate == "" {
		t.Fatalf("MarkInstalled() state = %+v", pkg)
	}
	pkg.MarkInstalledAsDep()
	if !pkg.InstalledAsDependency {
		t.Fatal("MarkInstalledAsDep() did not set dependency state")
	}
	pkg.MarkPinned()
	if !pkg.IsPinned {
		t.Fatal("MarkPinned() did not pin")
	}
	pkg.MarkUnpinned()
	if pkg.IsPinned {
		t.Fatal("MarkUnpinned() did not unpin")
	}
	pkg.MarkUninstalled()
	if pkg.IsInstalled || pkg.InstalledVersion != "" || pkg.IsOutdated || pkg.IsPinned || pkg.InstalledAsDependency {
		t.Fatalf("MarkUninstalled() state = %+v", pkg)
	}
}

func TestMatchKeywords(t *testing.T) {
	pkg := Package{Name: "DemoTool", Aliases: []string{"dt"}, Desc: "Fast Utility", Tap: "User/Tap", Homepage: "https://Example.test"}
	for _, tc := range []struct {
		words []string
		want  bool
	}{
		{nil, true}, {[]string{"dem"}, true}, {[]string{"n:dt"}, true}, {[]string{"d:fast"}, true}, {[]string{"t:user"}, true}, {[]string{"h:example"}, true}, {[]string{"utility", "-missing"}, true}, {[]string{"dem", "-tool"}, false}, {[]string{"unknown"}, false}, {[]string{"d:unknown"}, false}, {[]string{"t:unknown"}, false}, {[]string{"h:unknown"}, false}, {[]string{"n:unknown"}, false},
	} {
		if got := pkg.MatchKeywords(tc.words); got != tc.want {
			t.Errorf("MatchKeywords(%q) = %v, want %v", tc.words, got, tc.want)
		}
	}
}
