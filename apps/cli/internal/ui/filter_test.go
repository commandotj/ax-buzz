package ui

import (
	"strings"
	"testing"
)

func TestFilterOperationsAndParsing(t *testing.T) {
	if (Filter(0)).String() != "Unknown" || Filter(filterMax).String() != "Unknown" {
		t.Fatal("unknown filter string")
	}
	for _, pair := range []struct {
		s string
		f Filter
	}{{"Formulae", FilterFormulae}, {"Casks", FilterCasks}, {"Installed", FilterInstalled}, {"Outdated", FilterOutdated}, {"Expl. Installed", FilterExplicitlyInstalled}, {"Active", FilterActive}} {
		got, err := parseFilter(pair.s)
		if err != nil || got != pair.f {
			t.Errorf("parseFilter(%q) = %v, %v", pair.s, got, err)
		}
		if got.String() != pair.s {
			t.Errorf("filter String() = %q, want %q", got.String(), pair.s)
		}
	}
	if _, err := parseFilter("bogus"); err == nil {
		t.Fatal("unknown filter accepted")
	}
	var fg filterGroup
	if fg.String() != "None" {
		t.Fatal("empty group string")
	}
	fg.enableFilter(FilterFormulae)
	fg.enableFilter(FilterOutdated)
	if got := fg.String(); got != "Formulae & Outdated" {
		t.Fatalf("group string = %q", got)
	}
	if !fg.isFilterEnabled(FilterFormulae) || Filter(fg).getConflictFilters() != filterGroup(FilterFormulae|FilterCasks) {
		t.Fatal("conflicting group lookup failed")
	}
	fg.enableFilter(FilterCasks)
	if fg.isFilterEnabled(FilterFormulae) || !fg.isFilterEnabled(FilterCasks) {
		t.Fatal("enabling conflict did not replace previous filter")
	}
	fg.toggleFilter(FilterCasks)
	if fg.isFilterEnabled(FilterCasks) {
		t.Fatal("toggle did not disable filter")
	}
	fg.toggleFilter(FilterInstalled)
	fg.toggleFilter(FilterExplicitlyInstalled)
	if fg.isFilterEnabled(FilterInstalled) || !fg.isFilterEnabled(FilterExplicitlyInstalled) {
		t.Fatal("installed conflict did not replace")
	}
	got := fg.split()
	if len(got) != 2 {
		t.Fatalf("split = %v", got)
	}
	fg.reset()
	if fg != emptyFilterGroup {
		t.Fatal("reset failed")
	}
	for _, input := range [][]string{nil, {}, {"Formulae", "Outdated"}, {"Active", "Installed"}} {
		if _, err := parseFilterGroup(input); err != nil {
			t.Errorf("valid group %v: %v", input, err)
		}
	}
	for _, input := range [][]string{{"bad"}, {"Formulae", "Casks"}, {"Installed", "Expl. Installed"}} {
		if _, err := parseFilterGroup(input); err == nil {
			t.Errorf("invalid group accepted: %v", input)
		}
	}
	if _, _, ok := strings.Cut(parseFilterGroupError([]string{"Formulae", "Casks"}), "Conflicting filters"); !ok {
		t.Fatal("conflict error missing explanation")
	}
}

func parseFilterGroupError(filters []string) string {
	_, err := parseFilterGroup(filters)
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestFilterConflictHelpers(t *testing.T) {
	if enabled := (filterGroup(FilterActive)).isFilterEnabled(FilterActive); !enabled {
		t.Fatal("filter bit not detected")
	}
	if got := Filter(FilterActive).getConflictFilters(); got != emptyFilterGroup {
		t.Fatalf("non-conflicting filter group = %v", got)
	}
	if conflict, group := hasConflict(uint(FilterFormulae | FilterCasks)); !conflict || group != filterGroup(FilterFormulae|FilterCasks) {
		t.Fatal("formula/cask conflict not detected")
	}
	if conflict, _ := hasConflict(uint(FilterInstalled | FilterOutdated)); conflict {
		t.Fatal("independent filters flagged conflicting")
	}
}
