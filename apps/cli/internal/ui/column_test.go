package ui

import (
	"strings"
	"testing"

	"github.com/commandotj/ax-buzz/modules/brew/data"
)

func TestColumns(t *testing.T) {
	wants := []string{" ", "Name", "Version", "Tap", "Description", "Installs", "Size", "Status"}
	for i, want := range wants {
		if got := packageTableColumn(i).String(); got != want {
			t.Errorf("column %d String() = %q, want %q", i, got, want)
		}
	}
	if colUnknown.String() != "Unknown" || packageTableColumn(totalNumColumns+1).String() != "Unknown" {
		t.Fatal("unknown column name not handled")
	}
	for _, name := range wants[1:] {
		if _, err := parseColumnName(name); err != nil {
			t.Errorf("parseColumnName(%q): %v", name, err)
		}
	}
	if _, err := parseColumnName("bogus"); err == nil || !strings.Contains(err.Error(), "Unknown column") {
		t.Fatal("unknown column should return error")
	}
	for c := packageTableColumn(0); c < totalNumColumns; c++ {
		_ = c.hideable()
		_ = c.sortable()
		_ = c.reverseSort()
		_ = c.rightAligned()
		_ = c.width()
		if c.nextColumn().prevColumn() != c {
			t.Errorf("column cycle failed for %v", c)
		}
	}
	p := &data.Package{Name: "demo", Version: "1.2.3", Tap: "homebrew/core", Desc: "description", Installs90d: 42, FormattedSize: "5 MB", IsInstalled: true}
	checks := map[packageTableColumn]string{colSymbol: p.Symbol(), colName: "demo", colVersion: p.ShortVersion(), colTap: "homebrew/core", colDescription: "description", colInstalls: "42", colSize: "5 MB", colStatus: p.Status(), colUnknown: ""}
	for c, want := range checks {
		if got := c.getColumnData(p); got != want {
			t.Errorf("%v data = %q, want %q", c, got, want)
		}
	}
	p.IsInstalled = false
	if got := colSize.getColumnData(p); got != "N/A" {
		t.Fatalf("uninstalled size = %q", got)
	}
}
