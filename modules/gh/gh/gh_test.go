package gh

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestGetGithubReleaseInfo(t *testing.T) {
	date := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	body, err := json.Marshal(ghReleaseInfo{PublishDate: date, TagName: "v1.2.3", Url: "https://github.com/acme/tool/releases/tag/v1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		urls       []string
		homepage   string
		owner      string
		repo       string
		output     []byte
		commandErr error
		wantNil    bool
	}{
		{name: "release URL", urls: []string{"https://github.com/acme/tool"}, owner: "acme", repo: "tool", output: body},
		{name: "homepage repository", homepage: "https://github.com/acme/tool/", owner: "acme", repo: "tool", output: body},
		{name: "github pages homepage", homepage: "https://acme.github.io/tool/", owner: "acme", repo: "tool", output: body},
		{name: "unrecognized homepage", homepage: "https://example.com/tool", wantNil: true},
		{name: "invalid github release json", urls: []string{"https://github.com/acme/tool"}, owner: "acme", repo: "tool", output: []byte("not json"), wantNil: true},
		{name: "gh command fails", urls: []string{"https://github.com/acme/tool"}, owner: "acme", repo: "tool", commandErr: errors.New("command failed"), wantNil: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setGhInstalled(t)
			previous := ghReleaseView
			t.Cleanup(func() { ghReleaseView = previous })
			called := false
			ghReleaseView = func(owner, repo string) ([]byte, error) {
				called = true
				if owner != test.owner || repo != test.repo {
					t.Fatalf("gh release args = %s/%s, want %s/%s", owner, repo, test.owner, test.repo)
				}
				return test.output, test.commandErr
			}
			info := GetGithubReleaseInfo(test.urls, test.homepage)
			if test.wantNil {
				if info != nil {
					t.Fatalf("GetGithubReleaseInfo() = %#v, want nil", info)
				}
			} else {
				if info == nil {
					t.Fatal("GetGithubReleaseInfo() = nil, want release")
				}
				if info.Version != "v1.2.3" || info.Url != "https://github.com/acme/tool/releases/tag/v1.2.3" || !info.Date.Equal(date) {
					t.Fatalf("GetGithubReleaseInfo() = %#v, want release info", info)
				}
			}
			if called != (test.owner != "") {
				t.Fatalf("gh release command called = %t, want %t", called, test.owner != "")
			}
		})
	}
}

func TestGetGithubReleaseInfoWithoutGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if info := GetGithubReleaseInfo([]string{"https://github.com/acme/tool"}, ""); info != nil {
		t.Fatalf("GetGithubReleaseInfo() = %#v, want nil when gh is missing", info)
	}
}

func TestRunGhReleaseViewWithoutGh(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := runGhReleaseView("acme", "tool"); err == nil {
		t.Fatal("runGhReleaseView() error = nil, want missing gh error")
	}
}

func TestFetchLatestReleaseLogsCommandFailure(t *testing.T) {
	setGhInstalled(t)
	previous := ghReleaseView
	t.Cleanup(func() { ghReleaseView = previous })
	ghReleaseView = func(string, string) ([]byte, error) {
		return nil, &os.PathError{Op: "exec", Path: "gh", Err: errors.New("missing")}
	}
	if info := fetchLatestRelease("acme", "tool"); info != nil {
		t.Fatalf("fetchLatestRelease() = %#v, want nil", info)
	}
}

func TestFetchLatestReleaseLogsExitFailure(t *testing.T) {
	setGhInstalled(t)
	previous := ghReleaseView
	t.Cleanup(func() { ghReleaseView = previous })
	ghReleaseView = func(string, string) ([]byte, error) {
		return nil, &exec.ExitError{Stderr: []byte("not found")}
	}
	if info := fetchLatestRelease("acme", "tool"); info != nil {
		t.Fatalf("fetchLatestRelease() = %#v, want nil", info)
	}
}

func setGhInstalled(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, gh)
	if err := os.WriteFile(path, []byte(""), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("gh fixture unavailable: %v", err)
	}
}
