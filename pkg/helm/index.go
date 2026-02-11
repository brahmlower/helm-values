package helm

import (
	"fmt"
	"helmvalues/pkg/charts"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
	"go.yaml.in/yaml/v4"
)

func RepositoryIndexFromCache(repoName string) (*Index, error) {
	repositoryCache := os.Getenv("HELM_REPOSITORY_CACHE")
	if repositoryCache == "" {
		return nil, fmt.Errorf("HELM_REPOSITORY_CACHE environment variable is not set")
	}

	indexPath := filepath.Join(repositoryCache, repoName+"-index.yaml")

	if _, err := os.Stat(indexPath); err != nil {
		return nil, err
	}

	return LoadIndex(indexPath)
}

type Index struct {
	Entries map[string][]*charts.ChartDetails `yaml:"entries"`
}

// LoadIndex loads and parses a Helm index.yaml file
func LoadIndex(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	return &index, nil
}

// FindChart finds chart versions by name in the index
func (i *Index) FindChart(chartName string) ([]*charts.ChartDetails, error) {
	versions, ok := i.Entries[chartName]
	if !ok {
		return nil, fmt.Errorf("chart %q not found in index", chartName)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for chart %q", chartName)
	}

	return versions, nil
}

// GetVersion finds a specific version of a chart
func (i *Index) GetVersion(chartName, version string) (*charts.ChartDetails, error) {
	versions, err := i.FindChart(chartName)
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, fmt.Errorf("version %q not found for chart %q", version, chartName)
}

// GetLatestVersion finds the latest stable version of a chart
// Stable means non-prerelease versions (no -alpha, -beta, -rc suffixes)
func (i *Index) GetLatestVersion(chartName string) (*charts.ChartDetails, error) {
	versions, err := i.FindChart(chartName)
	if err != nil {
		return nil, err
	}

	// Parse and sort versions
	semvers := []*semver.Version{}
	versionMap := map[string]*charts.ChartDetails{}

	for _, v := range versions {
		sv, err := semver.NewVersion(v.Version)
		if err != nil {
			// Skip invalid semver versions
			continue
		}
		semvers = append(semvers, sv)
		versionMap[v.Version] = v
	}

	if len(semvers) == 0 {
		return nil, fmt.Errorf("no valid semver versions found for chart %q", chartName)
	}

	// Sort in descending order
	sort.Sort(sort.Reverse(semver.Collection(semvers)))

	// Find the first stable (non-prerelease) version
	for _, sv := range semvers {
		if sv.Prerelease() == "" {
			return versionMap[sv.Original()], nil
		}
	}

	// If no stable version found, return the latest version regardless
	return versionMap[semvers[0].Original()], nil
}
