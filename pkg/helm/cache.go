package helm

import (
	"fmt"
	"helmvalues/pkg/charts"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func GetCacheHome() (string, error) {
	cacheHome := os.Getenv("HELM_CACHE_HOME")
	if cacheHome == "" {
		return "", fmt.Errorf("HELM_CACHE_HOME environment variable is not set")
	}
	return cacheHome, nil
}

func FindRepositoryIndex(logger *logrus.Logger, repoName string) (*Index, error) {
	cacheHome, err := GetCacheHome()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(cacheHome, "repository", repoName+"-index.yaml")

	if _, err := os.Stat(indexPath); err != nil {
		return nil, err
	}

	return LoadIndex(indexPath)
}

// GetChartFromCache retrieves chart information from the Helm cache
func GetChartFromCache(logger *logrus.Logger, chartRef string, version string) (*charts.ChartDetails, error) {
	// Parse chart reference
	parts := strings.SplitN(chartRef, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid chart reference: %s", chartRef)
	}
	repoName := parts[0]
	chartName := parts[1]

	// Find the repository index
	index, err := FindRepositoryIndex(logger, repoName)
	if err != nil {
		return nil, err
	}

	// Get the specific version or latest
	var chartVersion *charts.ChartDetails
	if version != "" {
		chartVersion, err = index.GetVersion(chartName, version)
		if err != nil {
			return nil, err
		}
		logger.Infof("Using chart %s version %s", chartName, version)
	} else {
		chartVersion, err = index.GetLatestVersion(chartName)
		if err != nil {
			return nil, err
		}
		logger.Infof("Using latest stable version of %s: %s", chartName, chartVersion.Version)
	}

	return chartVersion, nil
}
