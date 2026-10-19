package model

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/commandotj/ax-buzz/apps/cli/internal/ui"
	"github.com/commandotj/ax-buzz/modules/brew/brew"
	"github.com/commandotj/ax-buzz/modules/brew/data"
	"github.com/commandotj/ax-buzz/modules/gh/gh"
)

func TestInitialModelLoadAndLayout(t *testing.T) {
	*flagHideHelp = false
	m := InitialModel()
	if m.Init() == nil {
		t.Fatal("Init should load data")
	}
	if m.loadData() == nil {
		t.Fatal("loadData should batch commands")
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(model)
	if m.width != 120 || m.height != 40 {
		t.Fatal("window dimensions not stored")
	}
	if got := m.View(); got == "" {
		t.Fatal("initial model view is empty")
	}
	*flagHideHelp = true
	if got := m.View(); strings.Contains(got, ": quit") {
		t.Fatal("help should be hidden")
	}
	*flagHideHelp = false
}

func TestModelDataFilteringAndMessages(t *testing.T) {
	m := InitialModel()
	m.width, m.height = 130, 40
	pkgs := []*data.Package{
		{Name: "alpha", Desc: "editor", Tap: "homebrew/core", IsInstalled: true},
		{Name: "beta", Desc: "editor", Tap: "homebrew/core", IsCask: true, IsInstalled: true, IsOutdated: true},
		{Name: "gamma", Desc: "tool", Tap: "homebrew/core", IsInstalled: true, InstalledAsDependency: true, IsDisabled: true},
		{Name: "delta", Desc: "tool", Tap: "homebrew/core", IsDeprecated: true},
	}
	updated, cmd := m.Update(brew.DataLoadedMsg{Packages: pkgs})
	m = updated.(model)
	if cmd == nil || len(m.table.Packages()) != 4 {
		t.Fatal("data load should populate table")
	}
	_ = m.loadingView.StopLoading()
	if got := m.View(); got == "" {
		t.Fatal("loaded model view is empty")
	}
	updated, _ = m.Update(ui.TableSelectionChangedMsg{Selected: pkgs[0]})
	m = updated.(model)
	if !strings.Contains(m.detailPanel.View(), "alpha") {
		t.Fatal("selection not forwarded to details")
	}
	updated, _ = m.Update(brew.DataLoadingErrMsg{Err: errors.New("load failed")})
	m = updated.(model)
	if !strings.Contains(m.loadingView.View(), "load failed") {
		t.Fatal("load error should appear in loading view")
	}
	updated, cmd = m.Update(spinner.TickMsg{})
	m = updated.(model)
	_ = cmd
	updated, cmd = m.Update(stopwatch.TickMsg{})
	m = updated.(model)
	_ = cmd
	updated, cmd = m.Update(stopwatch.StartStopMsg{})
	m = updated.(model)
	_ = cmd
	updated, cmd = m.Update(stopwatch.ResetMsg{})
	m = updated.(model)
	_ = cmd
	updated, _ = m.Update(brew.CommandStartMsg{})
	m = updated.(model)
	if !m.isExecuting {
		t.Fatal("command start state missing")
	}
	ch := make(chan tea.Msg, 1)
	updated, cmd = m.Update(brew.CommandOutputMsg{Line: "line", Ch: ch})
	m = updated.(model)
	if cmd == nil || !strings.Contains(m.outputView.View(), "line") {
		t.Fatal("command output not streamed")
	}
	_, _ = m.Update(brew.CommandOutputMsg{Line: "", Ch: ch})
	updated, _ = m.Update(brew.CommandFinishMsg{Command: brew.BrewCommandUpgrade, Pkgs: pkgs})
	m = updated.(model)
	if m.isExecuting || m.outputView.View() != "" {
		t.Fatal("successful finish state incorrect")
	}
	updated, _ = m.Update(brew.CommandStartMsg{})
	m = updated.(model)
	updated, _ = m.Update(brew.CommandOutputMsg{Line: "failure", Ch: ch})
	m = updated.(model)
	updated, _ = m.Update(brew.CommandFinishMsg{Err: errors.New("failed")})
	m = updated.(model)
	if m.isExecuting || m.outputView.View() == "" {
		t.Fatal("failed finish state incorrect")
	}
	updated, _ = m.Update(ui.SearchMsg{})
	m = updated.(model)
	updated, _ = m.Update(ui.FilterChangedMsg{})
	m = updated.(model)
	_ = m.loadingView.StartLoading()
	_ = m.loadingView.StopLoading()
	m.outputView.Append("visible command output")
	if got := m.View(); !strings.Contains(got, "visible command output") {
		t.Fatal("output section missing from loaded view")
	}
	*flagHideHelp = true
	if got := m.View(); strings.Contains(got, ": quit") {
		t.Fatal("help section should be hidden")
	}
	*flagHideHelp = false
}

func TestModelFilteringPredicates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filter ui.Filter
		want   int
	}{{"all", 0, 4}, {"formula", ui.FilterFormulae, 3}, {"cask", ui.FilterCasks, 1}, {"installed", ui.FilterInstalled, 3}, {"outdated", ui.FilterOutdated, 1}, {"explicit", ui.FilterExplicitlyInstalled, 2}, {"active", ui.FilterActive, 2}} {
		m := InitialModel()
		m.allPackages = []*data.Package{{Name: "a", IsInstalled: true}, {Name: "b", IsInstalled: true, IsOutdated: true, IsCask: true}, {Name: "c", IsInstalled: true, InstalledAsDependency: true, IsDisabled: true}, {Name: "d", IsDeprecated: true}}
		if tc.filter != 0 {
			filterKey := map[ui.Filter]string{ui.FilterFormulae: "f", ui.FilterCasks: "c", ui.FilterInstalled: "i", ui.FilterOutdated: "o", ui.FilterExplicitlyInstalled: "e", ui.FilterActive: "v"}[tc.filter]
			m.filterView, _ = m.filterView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(filterKey)})
		}
		m.filterPackages()
		if len(m.table.Packages()) != tc.want {
			t.Errorf("%s filter got %d packages want %d", tc.name, len(m.table.Packages()), tc.want)
		}
	}
	m := InitialModel()
	m.allPackages = []*data.Package{{Name: "Alpha", Desc: "Text Editor"}, {Name: "Beta", Desc: "database"}}
	m.search.SetFocused(true)
	for _, r := range "editor" {
		m.search, _ = m.search.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.filterPackages()
	if len(m.table.Packages()) != 1 || m.table.Packages()[0].Name != "Alpha" {
		t.Fatal("keyword search filter failed")
	}
}

func TestModelKeyboardRoutes(t *testing.T) {
	oldOpen, oldOutdated := openExternalURL, getOutdatedPackages
	t.Cleanup(func() { openExternalURL, getOutdatedPackages = oldOpen, oldOutdated })
	var opened []string
	openExternalURL = func(url string) error { opened = append(opened, url); return nil }
	getOutdatedPackages = func() []*data.Package { return []*data.Package{{Name: "old"}} }
	m := InitialModel()
	m.width, m.height = 120, 40
	m.allPackages = []*data.Package{{Name: "a", Homepage: "https://example.test", IsInstalled: true, IsOutdated: true, InstallSupported: true, ReleaseInfo: &gh.ReleaseInfo{Url: "https://example.test/release"}}, {Name: "b", IsCask: true, InstallSupported: true}}
	m.filterPackages()
	key := func(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	if m.focusMode != focusDetail {
		t.Fatal("tab should focus details")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	updated, _ = m.Update(key("/"))
	m = updated.(model)
	if m.focusMode != focusSearch {
		t.Fatal("slash should focus search")
	}
	updated, _ = m.Update(key("x"))
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.focusMode != focusTable {
		t.Fatal("enter should exit search")
	}
	updated, _ = m.Update(key("/"))
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.focusMode != focusTable || m.search.Value() != "" {
		t.Fatal("escape should clear search")
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	m = updated.(model)
	for _, k := range []string{"h", "b", "r", "U", "u", "t", "x", "p", "P", "L", "z"} {
		if k == "t" {
			updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
			m = updated.(model)
		}
		if k == "x" || k == "p" {
			updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
			m = updated.(model)
		}
		if k == "P" {
			m.allPackages[0].IsPinned = true
		}
		updated, _ = m.Update(key(k))
		m = updated.(model)
	}
	if len(opened) != 3 || opened[0] != "https://example.test" || !strings.Contains(opened[1], "/formula/a") || opened[2] != "https://example.test/release" {
		t.Fatalf("open url calls = %v", opened)
	}
	if cmd := m.handleTableKeys(key("U")); cmd == nil {
		t.Fatal("upgrade all should create command when outdated packages exist")
	}
	m.isExecuting = true
	for _, k := range []string{"U", "u", "t", "x", "p", "P"} {
		updated, _ = m.Update(key(k))
		m = updated.(model)
	}
	m.isExecuting = false
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(key("z"))
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || cmd() == nil {
		t.Fatal("quit key should quit")
	}
	m = updated.(model)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	empty := InitialModel()
	for _, k := range []string{"h", "b", "r"} {
		_, _ = empty.Update(key(k))
	}
}
