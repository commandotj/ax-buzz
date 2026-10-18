package gh

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"time"
)

type ghReleaseInfo struct {
	PublishDate time.Time `json:"publishedAt"`
	TagName     string    `json:"tagName"`
	Url         string    `json:"url"`
}

type ReleaseInfo struct {
	Date    time.Time
	Version string
	Url     string
}

const (
	gh            = "gh"
	releaseFields = "publishedAt,tagName,url"
)

var (
	githubRepoUrl = regexp.MustCompile(`^https://github.com/([^/\s]+)/([^/\.\s]+)`)
	githubPageUrl = regexp.MustCompile(`^https://([^.\s]+).github.io/([^/\s]+)`)
)

func GetGithubReleaseInfo(urls []string, homepage string) *ReleaseInfo {
	if !isGhInstalled() {
		return nil
	}

	for _, url := range urls {
		if matches := githubRepoUrl.FindStringSubmatch(url); len(matches) > 0 {
			// Package url matches a github repo
			return fetchLatestRelease(matches[1], matches[2])
		}
	}

	if matches := githubRepoUrl.FindStringSubmatch(homepage); len(matches) > 0 {
		// Package home page matches a github repo
		return fetchLatestRelease(matches[1], matches[2])
	} else if matches := githubPageUrl.FindStringSubmatch(homepage); len(matches) > 0 {
		// Package home page matches a github page
		return fetchLatestRelease(matches[1], matches[2])
	} else {
		return nil
	}
}

func isGhInstalled() bool {
	if _, err := exec.LookPath(gh); err == nil {
		return true
	} else {
		return false
	}
}

func fetchLatestRelease(ghOwner, ghRepo string) *ReleaseInfo {
	var note ghReleaseInfo
	cmd := exec.Command(gh, "release", "view", "--repo", fmt.Sprintf("%s/%s", ghOwner, ghRepo), "--json", releaseFields)

	body, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			log.Printf("Failed to get release info for %s/%s: %s", ghOwner, ghRepo, e.Stderr)
		}
		return nil
	}

	if err := json.Unmarshal(body, &note); err != nil {
		log.Printf("Failed to decode json from 'gh release view' response %s: %v", body, err)
		return nil
	} else {
		return toReleaseInfo(&note)
	}
}

func toReleaseInfo(info *ghReleaseInfo) *ReleaseInfo {
	return &ReleaseInfo{
		Date:    info.PublishDate,
		Version: info.TagName,
		Url:     info.Url,
	}
}
