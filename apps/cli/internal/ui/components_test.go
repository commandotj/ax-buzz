package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/commandotj/ax-buzz/modules/brew/data"
	"github.com/commandotj/ax-buzz/modules/gh/gh"
)

func TestBorderTitle(t *testing.T) {
	for _, tc := range []struct {
		title string
		width int
		want  string
	}{{"ABC", 0, ""}, {"ABC", 2, "AB"}, {"ABC", 3, "ABC"}, {"ABC", 4, "─ABC"}, {"ABC", 5, "─ABC─"}, {"ABC", 8, "─ABC────"}} {
		if got := getBorderTopWithTitle(tc.title, tc.width); got != tc.want {
			t.Errorf("getBorderTopWithTitle(%q,%d)=%q want %q", tc.title, tc.width, got, tc.want)
		}
	}
	if got := getRoundedBorderWithTitle("Hi", 5); got.Top != "─Hi──" || got.Bottom != roundedBorder.Bottom {
		t.Fatalf("rounded border not composed: %#v", got)
	}
}

func TestOutputHelpStatsAndSearch(t *testing.T) {
	out := NewOutputModel()
	if out.View() != "" {
		t.Fatal("empty output should be hidden")
	}
	out.SetWidth(40)
	for i := 0; i < outputMaxLines+2; i++ {
		out.Append(string(rune('a' + i)))
	}
	if got := out.View(); got == "" {
		t.Fatal("non-error output should render")
	}
	if got := out.View(); !strings.Contains(got, "│ c ") || !strings.Contains(got, "│ l ") || strings.Contains(got, "│ a ") {
		t.Fatalf("output did not retain last lines: %q", got)
	}
	out.SetError()
	if got := out.View(); got == "" {
		t.Fatal("error output hidden")
	}
	out.Clear()
	if len(out.lines) != 0 || out.hasError || out.View() != "" {
		t.Fatal("clear failed")
	}
	help := NewHelpModel()
	if help.View() == "" {
		t.Fatal("help view empty")
	}
	help.SetWidth(80)
	stats := NewStatsModel()
	stats.SetWidth(80)
	stats.SetPackages([]*data.Package{{Name: "f", Size: 100, IsInstalled: true, InstalledAsDependency: true}, {Name: "c", Size: 200, IsInstalled: true, IsCask: true}, {Name: "available"}, {Name: "cask", IsCask: true}})
	statsText := ansi.Strip(stats.View())
	for _, expected := range []string{"2 Formulae available", "2 Casks available", "1 Formulae (incl. 1 deps)", "installed taking 100KB", "1 Casks installed"} {
		if !strings.Contains(statsText, expected) {
			t.Errorf("stats missing %q", expected)
		}
	}
	search := NewSearchInputModel()
	if search.Value() != "" {
		t.Fatal("new search has value")
	}
	search.SetWidth(30)
	search.SetFocused(true)
	_, cmd := search.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	_ = cmd
	search.Clear()
	if search.Value() != "" {
		t.Fatal("search clear failed")
	}
	if search.sendSearchMsg()().(SearchMsg).Query != "" {
		t.Fatal("search message mismatch")
	}
	search.SetFocused(false)
	_ = search.View()
}

func TestLoadingScreen(t *testing.T) {
	*flagShowLoadTimer = false
	m := NewLoadingScreenModel()
	if m.Progress() == nil || m.View() == "" {
		t.Fatal("loading model not initialized")
	}
	_ = m.StartLoading()
	m, _ = m.Update(spinner.TickMsg{})
	m, _ = m.Update(stopwatch.TickMsg{})
	_ = m.SetError("bad")
	if !strings.Contains(m.View(), "bad") {
		t.Fatal("error screen failed")
	}
	_ = m.StopLoading()
	if !strings.Contains(m.View(), "bad") {
		t.Fatal("stopped screen should be blank")
	}
	clearModel := NewLoadingScreenModel()
	_ = clearModel.StopLoading()
	if clearModel.View() != "" {
		t.Fatal("stopped screen should be blank")
	}
	*flagShowLoadTimer = true
	m = NewLoadingScreenModel()
	_ = m.StartLoading()
	if !strings.Contains(m.View(), "Loading") {
		t.Fatal("timed loading screen missing")
	}
	_ = m.StopLoading()
	_, _ = m.Update(stopwatch.StartStopMsg{})
	_, _ = m.Update(stopwatch.ResetMsg{})
	*flagShowLoadTimer = false
}

func TestFilterViewKeysAndMessages(t *testing.T) {
	*flagFilters = []string{"Formulae"}
	m := NewFilterViewModel()
	m.SetWidth(40)
	_ = m.View()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd == nil || len(updated.Value()) != 1 || updated.Value()[0] != FilterCasks {
		t.Fatal("cask key should replace formula filter")
	}
	if msg, ok := cmd().(FilterChangedMsg); !ok || msg.filters[0] != FilterCasks {
		t.Fatal("filter changed message missing")
	}
	for _, key := range []rune{'a', 'f', 'i', 'o', 'e', 'v'} {
		updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		if cmd == nil {
			t.Errorf("key %q did not change filter", key)
		}
	}
	_, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if cmd != nil {
		t.Fatal("unhandled key emitted command")
	}
	_, cmd = updated.Update(tea.WindowSizeMsg{})
	if cmd != nil {
		t.Fatal("non-key emitted command")
	}
	*flagFilters = []string{"bad"}
	invalid := NewFilterViewModel()
	if got := invalid.Value(); len(got) != 0 {
		t.Fatal("invalid CLI filters should fall back to empty")
	}
	*flagFilters = []string{}
}

func TestDetailsPanelBasicStatesAndStatusSymbols(t *testing.T) {
	for _, pkg := range []*data.Package{{}, {IsDisabled: true}, {IsDeprecated: true}, {IsPinned: true}, {IsOutdated: true}, {IsInstalled: true, InstalledAsDependency: true}, {IsInstalled: true}} {
		if got := formatStatus(pkg); got == "" || formatStatusSymbol(pkg) == "" {
			t.Errorf("status missing for %#v", pkg)
		}
	}
	if got := hyperLink("https://example.test", "label"); !strings.Contains(got, "https://example.test") || !strings.Contains(got, "label") {
		t.Fatalf("link malformed: %q", got)
	}
	m := NewDetailsPanelModel()
	m.SetDimension(40, 10)
	m.SetPackage(nil)
	if !strings.Contains(m.View(), "No packages selected") {
		t.Fatal("empty detail panel missing placeholder")
	}
	m.SetPackage(&data.Package{Name: "demo", Version: "1.2", Desc: "desc", Tap: "homebrew/core", Homepage: "https://example.test", License: "MIT", Installs90d: 5, IsInstalled: true, FormattedSize: "5 MB", InstalledDate: "2026-10-18", ReleaseInfo: &gh.ReleaseInfo{Date: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)}, Conflicts: []string{"missing-conflict"}, Dependencies: []string{"missing-dependency"}, BuildDependencies: []string{"missing-build-dependency"}, Dependents: []string{"missing-dependent"}})
	if got := m.View(); !strings.Contains(got, "demo") || !strings.Contains(got, "Version: 1.2") {
		t.Fatalf("detail view missing package: %q", got)
	}
	m.SetFocused(true)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.SetFocused(false)
}

func TestDetailsPanelRelations(t *testing.T) {
	oldPackage, oldMissing, oldDependents := getPackage, getRecursiveMissingDeps, getRecursiveInstalledDependents
	t.Cleanup(func() {
		getPackage, getRecursiveMissingDeps, getRecursiveInstalledDependents = oldPackage, oldMissing, oldDependents
	})
	graph := map[string]*data.Package{
		"conflict":                {Name: "conflict"},
		"dependency":              {Name: "dependency"},
		"leaf":                    {Name: "leaf"},
		"installed-dependency":    {Name: "installed-dependency", IsInstalled: true},
		"build":                   {Name: "build"},
		"dependent":               {Name: "dependent", IsInstalled: true},
		"not-installed-dependent": {Name: "not-installed-dependent"},
		"explicit":                {Name: "explicit", IsInstalled: true},
		"transitive-dependency":   {Name: "transitive-dependency", IsInstalled: true, InstalledAsDependency: true},
	}
	getPackage = func(name string) *data.Package { return graph[name] }
	getRecursiveMissingDeps = func(string) []string { return []string{"leaf", "dependency", "missing-recursive"} }
	getRecursiveInstalledDependents = func(string) []string { return []string{"explicit", "transitive-dependency", "missing-recursive"} }
	pkg := &data.Package{
		Name: "root", Version: "2.0", Desc: "root package", IsInstalled: true, FormattedSize: "10 MB", InstalledDate: "today",
		ReleaseInfo:       &gh.ReleaseInfo{Date: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)},
		Conflicts:         []string{"conflict", "missing-conflict"},
		Dependencies:      []string{"dependency", "installed-dependency", "missing-dependency"},
		BuildDependencies: []string{"build", "missing-build"},
		Dependents:        []string{"dependent", "not-installed-dependent", "missing-dependent"},
	}
	m := NewDetailsPanelModel()
	m.SetDimension(50, 100)
	m.SetPackage(pkg)
	view := ansi.Strip(m.View())
	for _, part := range []string{"Released on: 2026-10-18", "Conflicts:", "Dependencies:", "leaf", "Build dependencies:", "Required By:", "explicit", "not-installed-dependent"} {
		if !strings.Contains(view, part) {
			t.Errorf("details missing %q: %s", part, view)
		}
	}
}

func TestPackageTableRowsColumnsSortingAndSelection(t *testing.T) {
	*flagHideCols = []string{}
	*flagSortColumn = "Name"
	m := NewPackageTableModel()
	m.SetDimensions(160, 8)
	if !m.ShowPackageInstalls() || !m.ShowPackageSizes() {
		t.Fatal("default columns missing")
	}
	pkgs := []*data.Package{{Name: "z", Tap: "b", Installs90d: 1, Size: 10}, {Name: "a", Tap: "a", Installs90d: 9, Size: 20}}
	_ = m.SetPackages(pkgs)
	m.table.SetCursor(0)
	if m.Selected().Name != "a" || len(m.Packages()) != 2 {
		t.Fatal("name sort or package access failed")
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd == nil {
		t.Fatal("table selection change should emit message")
	}
	m.table.SetCursor(0)
	for _, column := range []packageTableColumn{colTap, colInstalls, colSize, colStatus, colName} {
		m.sortColumn = column
		m.sortRows()
	}
	m.sortColumn = colInstalls
	_, sortedCols := m.getVisibleCols()
	if len(sortedCols) == 0 {
		t.Fatal("sorted columns should render")
	}
	if msg := m.sendSelectionChangedMsg()().(TableSelectionChangedMsg); msg.Selected.Name != "a" {
		t.Fatal("selection message incorrect")
	}
	if cmd := m.SetPackages([]*data.Package{{Name: "replacement"}}); cmd == nil {
		t.Fatal("changed selection should emit message")
	}
	if cmd := m.SetPackages(m.Packages()); cmd != nil {
		t.Fatal("same selected package should not emit message")
	}
	if m.Selected() == nil {
		t.Fatal("selection lost during sort")
	}
	m.sortColumn = colName
	m.sortNextColumn()
	m.sortPrevColumn()
	m.SetFocused(true)
	m.SetFocused(false)
	_ = m.View()
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if got := m.Selected(); got == nil {
		t.Fatal("cursor selection missing")
	}
	_ = getTableStyles()
	*flagHideCols = []string{"Description"}
	hidden := NewPackageTableModel()
	hidden.SetDimensions(60, 8)
	if hidden.isColumnEnabled(colDescription) {
		t.Fatal("hidden column still enabled")
	}
	hidden.SetDimensions(160, 8)
	if hidden.isColumnVisible(colDescription) {
		t.Fatal("description should remain hidden")
	}
	visible, rendered := hidden.getVisibleCols()
	if len(visible) == 0 || len(rendered) == 0 {
		t.Fatal("visible column calculation returned empty result")
	}
	hidden.SetPackages([]*data.Package{{Name: "first"}, {Name: "second"}})
	hidden.table.SetCursor(1)
	hidden.packages = hidden.packages[:1]
	hidden.UpdateRows()
	if hidden.table.Cursor() != 0 {
		t.Fatal("out-of-range cursor was not reset")
	}
	*flagHideCols = []string{}
}

func TestTableConfigValidationAndExit(t *testing.T) {
	if columns, sort, err := parseTableConfig([]string{"Description"}, "Name"); err != nil || sort != colName || len(columns) != int(totalNumColumns)-1 {
		t.Fatalf("valid config: %v %v %v", columns, sort, err)
	}
	for _, tc := range []struct {
		hidden   []string
		sort     string
		fragment string
	}{{[]string{"bogus"}, "Name", "Unknown column"}, {[]string{"Name"}, "Name", "can not be hidden"}, {nil, "bogus", "Unknown column"}, {nil, "Version", "Can not sort"}} {
		if _, _, err := parseTableConfig(tc.hidden, tc.sort); err == nil || !strings.Contains(err.Error(), tc.fragment) {
			t.Errorf("config (%v,%s) err=%v", tc.hidden, tc.sort, err)
		}
	}
	oldHide, oldSort, oldExit := *flagHideCols, *flagSortColumn, exitInvalidTableConfig
	t.Cleanup(func() { *flagHideCols, *flagSortColumn, exitInvalidTableConfig = oldHide, oldSort, oldExit })
	*flagHideCols = []string{"bogus"}
	*flagSortColumn = "Name"
	exitCode := 0
	exitInvalidTableConfig = func(code int) { exitCode = code }
	if model := NewPackageTableModel(); model.table.Columns() != nil {
		t.Fatal("invalid configuration returned live table")
	}
	if exitCode != 1 {
		t.Fatalf("configuration exit status=%d", exitCode)
	}
}
