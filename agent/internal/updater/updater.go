// Package updater checks GitHub Releases and self-updates the agent binary.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const repo = "FireGams/managerhub"

// Info holds release version info.
type Info struct {
	Tag    string
	URL    string
	HasNew bool
}

// CheckForUpdate compares the running version with the latest GitHub release.
func CheckForUpdate(current string) Info {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return Info{Tag: current}
	}
	defer func() { _ = resp.Body.Close() }()
	var rel struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Info{Tag: current}
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	asset := fmt.Sprintf("managerhub-agent-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
	for _, a := range rel.Assets {
		if a.Name == asset {
			hasNew := rel.TagName != current && current != "dev"
			return Info{Tag: rel.TagName, URL: a.URL, HasNew: hasNew}
		}
	}
	return Info{Tag: rel.TagName}
}

// Apply downloads the new binary and replaces the current executable.
func Apply(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	tmp := exe + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	_ = f.Close()

	if runtime.GOOS == "windows" {
		old := exe + ".old"
		_ = os.Rename(exe, old)
		if err := os.Rename(tmp, exe); err != nil {
			_ = os.Rename(old, exe)
			return err
		}
		_ = os.Remove(old)
	} else {
		if err := os.Rename(tmp, exe); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	}
	return nil
}

// ShouldAutoUpdate reports whether auto-update is enabled.
func ShouldAutoUpdate() bool {
	v := os.Getenv("MH_AUTO_UPDATE")
	return v == "true" || v == "1"
}
