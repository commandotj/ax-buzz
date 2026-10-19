package util

import (
	"reflect"
	"testing"
)

func TestSortAndUniq(t *testing.T) {
	if got := SortAndUniq(nil); got == nil { // Empty input stays empty.
	} else {
		t.Fatalf("SortAndUniq(nil) = %#v", got)
	}
	input := []string{"b", "a", "b", "c", "a"}
	if got := SortAndUniq(input); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("SortAndUniq() = %#v", got)
	}
}

func TestSort(t *testing.T) {
	input := []string{"z", "a"}
	if got := Sort(input); !reflect.DeepEqual(got, []string{"a", "z"}) {
		t.Fatalf("Sort() = %#v", got)
	}
}

func TestFormatSize(t *testing.T) {
	for _, tc := range []struct {
		kbs  int64
		want string
	}{{0, "0"}, {1, "1KB"}, {1024, "1MB"}, {1536, "1.5MB"}, {1048576, "1GB"}, {1572864, "1.5GB"}} {
		if got := FormatSize(tc.kbs); got != tc.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.kbs, got, tc.want)
		}
	}
}
