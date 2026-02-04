package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GithubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GithubRelease struct {
	TagName     string        `json:"tag_name"`
	PublishedAt string        `json:"published_at"`
	Assets      []GithubAsset `json:"assets"`
}

func (r *GithubRelease) PluginURL() (string, error) {
	asset, err := r.ReleaseArtifact()
	if err != nil {
		return "", err
	}
	return asset.BrowserDownloadURL, nil
}

func (r *GithubRelease) ReleaseArtifact() (*GithubAsset, error) {
	expectedName := fmt.Sprintf("values-%s.tgz", r.TagName)

	for _, asset := range r.Assets {
		if asset.Name == expectedName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("expected asset not found")
}

func GetLatestRelease(repository string) (*GithubRelease, error) {
	githubReleasesURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repository)
	resp, err := http.Get(githubReleasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %w", err)
	}

	return &release, nil
}
