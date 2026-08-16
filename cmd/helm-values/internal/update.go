package internal

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Update checks repository's latest GitHub release against currentVersion
// and, if newer, uninstalls and reinstalls the helm-values plugin via helm.
func Update(ctx context.Context, logger *logrus.Logger, repository string, currentVersion string) error {
	logger.Info("Fetching latest release information...")

	// Get the latest release from GitHub
	release, err := GetLatestRelease(ctx, repository)
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
	uninstallCmd := exec.CommandContext(ctx, "helm", "plugin", "uninstall", "values")
	if err := uninstallCmd.Run(); err != nil {
		logger.Warnf("Failed to uninstall plugin (it may not be installed): %v", err)
	}

	// Install the new version. This is a deliberate, user-initiated subprocess
	// call to the helm plugin manager as part of this tool's self-update flow;
	// pluginURL comes from the GitHub releases API response for this project's
	// own repository, not from attacker-controlled input.
	//nolint:gosec // see comment above
	installCmd := exec.CommandContext(ctx, "helm", "plugin", "install", pluginURL)
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Printf("Successfully updated helm-values to %s\n", release.TagName)

	return nil
}
