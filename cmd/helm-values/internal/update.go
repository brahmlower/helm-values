package internal

import (
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Update(logger *logrus.Logger, repository string, currentVersion string) error {
	logger.Info("Fetching latest release information...")

	// Get the latest release from GitHub
	release, err := GetLatestRelease(repository)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	logger.Infof("Latest version: %s", release.TagName)

	// Check if we're already at the latest version
	if currentVersion == release.TagName {
		fmt.Printf("Already at the latest version (%s)\n", currentVersion)
		return nil
	}

	pluginURL, err := release.PluginURL()
	if err != nil {
		return fmt.Errorf("failed to get plugin URL: %w", err)
	}

	// Uninstall the current plugin
	uninstallCmd := exec.Command("helm", "plugin", "uninstall", "values")
	if err := uninstallCmd.Run(); err != nil {
		logger.Warnf("Failed to uninstall plugin (it may not be installed): %v", err)
	}

	// Install the new version
	installCmd := exec.Command("helm", "plugin", "install", pluginURL)
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Printf("Successfully updated helm-values to %s\n", release.TagName)
	return nil
}
