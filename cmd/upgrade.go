package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	timeout          = 10 * time.Second
	latestReleaseURL = "https://api.github.com/repos/evilmartians/lefthook/releases/latest"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []Asset
}

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func newUpgradeCmd(_ *lefthook.Options) *cobra.Command {
	var yes bool
	upgradeCmd := cobra.Command{
		Use:               "upgrade",
		Short:             "Upgrade lefthook executable",
		Example:           "lefthook upgrade",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return upgrade(yes)
		},
	}

	upgradeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "no prompt")

	return &upgradeCmd
}

func upgrade(ask bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Determine the download URL
	// curl -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" https://api.github.com/repos/evilmartians/lefthook/releases/latest
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize a request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var release Release
	json.NewDecoder(resp.Body).Decode(&release)

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	if latestVersion == version.Version(false) {
		log.Infof("Already installed the latest version: %s", latestVersion)
		return nil
	}
	// Download the file and the hashsums
	// .assets[N].browser_download_url

	// Check hashsums

	// If ok - replace the binary
	return nil
}
