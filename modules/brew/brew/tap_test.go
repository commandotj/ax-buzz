package brew

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCustomTapPackageParsing(t *testing.T) {
	for _, tc := range []struct {
		name, body, version string
		wantErr             bool
	}{
		{"version", "version \"1.2\"\ndesc \"d\"\nhomepage \"h\"\nurl \"https://x/1.2.tar.gz\"", "1.2", false},
		{"latest", "version :latest\ndesc \"d\"\nhomepage \"h\"", "latest", false},
		{"tag", "tag: \"v2.0\"\ndesc \"d\"\nhomepage \"h\"", "2.0", false},
		{"url-version", "url \"https://x/v3.4.zip\"\ndesc \"d\"\nhomepage \"h\"", "3.4", false},
		{"no-version", "desc \"d\"\nhomepage \"h\"", "", true},
		{"no-desc", "version \"1\"\nhomepage \"h\"", "", true},
		{"no-homepage", "version \"1\"\ndesc \"d\"", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "custom.rb")
			if err := os.WriteFile(path, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			pkg, err := getCustomTapPackage(&installInfo{name: "custom", tap: "user/tap", path: path})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected validation error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if pkg.Version != tc.version {
				t.Fatalf("version=%q, want %q", pkg.Version, tc.version)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "missing.rb")
	if _, err := getCustomTapPackage(&installInfo{path: path}); err == nil {
		t.Fatal("expected read error")
	}
	path = filepath.Join(t.TempDir(), "formula.rb")
	body := "version \"1\"\nrevision 3\ndesc \"d\"\nhomepage \"h\"\nurl \"https://x/file.tar.xz\"\nurl \"https://x/other.tgz\"\nlicense \"MIT\"\ndepends_on \"build\" => :build\ndepends_on \"runtime\"\nconflicts_with \"other\"\ndisable!\ndeprecate!"
	_ = os.WriteFile(path, []byte(body), 0600)
	pkg, err := getCustomTapPackage(&installInfo{path: path})
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Revision != 3 || pkg.License != "MIT" || len(pkg.Urls) != 2 || len(pkg.BuildDependencies) != 1 || len(pkg.Dependencies) != 1 || len(pkg.Conflicts) != 1 || !pkg.IsDisabled || !pkg.IsDeprecated {
		t.Fatalf("parsed package=%+v", pkg)
	}
}

func TestParseVersionFromURL(t *testing.T) {
	for _, tc := range []struct{ url, want string }{
		{"https://x/v1.2.tar.gz", "1.2"}, {"https://x/v1.2.tar.xz", "1.2"}, {"https://x/v1.2.tar.bz2", "1.2"}, {"https://x/v1.2.tgz", "1.2"}, {"https://x/v1.2.zip", "1.2"}, {"https://x/no-version", ""},
	} {
		if got := parseVersionFromUrl(tc.url); got != tc.want {
			t.Errorf("%q => %q, want %q", tc.url, got, tc.want)
		}
	}
}
