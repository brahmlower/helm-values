// Package internal implements the helm-values plugin's supporting
// commands: self-update (via GitHub releases) and pre-commit hook
// installation.
package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GithubAsset is a single downloadable file attached to a GitHub release.
type GithubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GithubRelease is a GitHub release, as returned by the releases API.
type GithubRelease struct {
	TagName     string        `json:"tag_name"`
	PublishedAt string        `json:"published_at"`
	Assets      []GithubAsset `json:"assets"`
}

// PluginURL returns the download URL for this release's helm-values plugin
// archive.
func (r *GithubRelease) PluginURL() (string, error) {
	asset, err := r.ReleaseArtifact()
	if err != nil {
		return "", err
	}

	return asset.BrowserDownloadURL, nil
}

// ReleaseArtifact finds this release's helm-values plugin archive asset.
func (r *GithubRelease) ReleaseArtifact() (*GithubAsset, error) {
	expectedName := fmt.Sprintf("values-%s.tgz", r.TagName)

	for _, asset := range r.Assets {
		if asset.Name == expectedName {
			return &asset, nil
		}
	}

	return nil, errors.New("expected asset not found")
}

// GetLatestRelease fetches the latest GitHub release for repository (in
// "owner/name" form).
func GetLatestRelease(ctx context.Context, repository string) (*GithubRelease, error) {
	githubReleasesURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repository)

	// githubReleasesURL is built from a fixed, hardcoded GitHub API endpoint
	// with the repository slug interpolated (supplied at build time via
	// -ldflags, not user input), so this is not an SSRF-style variable URL.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubReleasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %w", err)
	}

	return &release, nil
}
