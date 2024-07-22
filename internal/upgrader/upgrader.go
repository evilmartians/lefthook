// Package upgrader contains the auto-upgrade implementation for the lefthook executable.
package upgrader

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	timeout                       = 10 * time.Second
	latestReleaseURL              = "https://api.github.com/repos/evilmartians/lefthook/releases/latest"
	checksumsFilename             = "lefthook_checksums.txt"
	modExecutable     os.FileMode = 0o755
)

var (
	errNoAsset        = errors.New("Couldn't find an asset to download. Please submit an issue to https://github.com/evilmartians/lefthook")
	errInvalidHashsum = errors.New("SHA256 sums differ, it's not safe to use the downloaded binary.\nIf you have problems upgrading lefthook please submit an issue to https://github.com/evilmartians/lefthook")

	osNames = map[string]string{
		"windows": "Windows",
		"darwin":  "MacOS",
		"linux":   "Linux",
		"freebsd": "Freebsd",
	}

	archNames = map[string]string{
		"amd64": "x86_64",
		"arm64": "arm64",
		"386":   "i386",
	}
)

type release struct {
	TagName string `json:"tag_name"`
	Assets  []asset
}

type asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type Upgrader struct {
	client *http.Client
}

func New() *Upgrader {
	return &Upgrader{
		client: &http.Client{Timeout: timeout},
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, yes, force bool) error {
	rel, ferr := u.fetchLatestRelease(ctx)
	if ferr != nil {
		return fmt.Errorf("latest release fetch failed: %w", ferr)
	}

	latestVersion := strings.TrimPrefix(rel.TagName, "v")

	if latestVersion == version.Version(false) && !force {
		log.Infof("Already installed the latest version: %s\n", latestVersion)
		return nil
	}

	wantedAsset := fmt.Sprintf("lefthook_%s_%s_%s", latestVersion, osNames[runtime.GOOS], archNames[runtime.GOARCH])
	if runtime.GOOS == "windows" {
		wantedAsset += ".exe"
	}

	log.Debugf("Searching assets for %s", wantedAsset)

	var downloadURL string
	var checksumURL string
	for i := range rel.Assets {
		asset := rel.Assets[i]
		if len(downloadURL) == 0 && asset.Name == wantedAsset {
			downloadURL = asset.DownloadURL
			if len(checksumURL) > 0 {
				break
			}
		}

		if len(checksumURL) == 0 && asset.Name == checksumsFilename {
			checksumURL = asset.DownloadURL
			if len(downloadURL) > 0 {
				break
			}
		}
	}

	if len(downloadURL) == 0 {
		log.Warnf("Couldn't find the right asset to download. Wanted: %s\n", wantedAsset)
		return errNoAsset
	}

	if len(checksumURL) == 0 {
		log.Warn("Couldn't find checksums")
	}

	if !yes {
		log.Infof("Upgrade %s to %s? %s ", log.Cyan("lefthook"), log.Yellow(latestVersion), log.Gray("[Y/n]"))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		ans := scanner.Text()

		if len(ans) > 0 && ans[0] != 'y' && ans[0] != 'Y' {
			log.Debug("Upgrade rejected")
			return nil
		}
	}

	lefthookExePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine the binary path: %w", err)
	}

	if realPath, serr := filepath.EvalSymlinks(lefthookExePath); serr == nil {
		lefthookExePath = realPath
	}

	destPath := lefthookExePath + "." + latestVersion
	defer os.Remove(destPath)

	log.Debugf("Downloading to %s", destPath)
	ok, err := u.download(ctx, wantedAsset, downloadURL, checksumURL, destPath)
	if err != nil {
		return err
	}
	if !ok {
		return errInvalidHashsum
	}

	backupPath := lefthookExePath + ".bak"
	defer os.Remove(backupPath)

	log.Debugf("Renaming %s -> %s", lefthookExePath, backupPath)
	if err = os.Rename(lefthookExePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup lefthook executable: %w", err)
	}

	log.Debugf("Renaming %s -> %s", destPath, lefthookExePath)
	err = os.Rename(destPath, lefthookExePath)
	if err != nil {
		log.Errorf("Failed to replace the lefthook executable: %s\n", err)
		if err = os.Rename(backupPath, lefthookExePath); err != nil {
			return fmt.Errorf("failed to recover from backup: %w", err)
		}

		return nil
	}

	log.Debugf("Making file %s executable", lefthookExePath)
	if err = os.Chmod(lefthookExePath, modExecutable); err != nil {
		return fmt.Errorf("failed to set mod executable: %w", err)
	}

	return nil
}

func (u *Upgrader) fetchLatestRelease(ctx context.Context) (*release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var rel release
	if err = json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse the Github response: %w", err)
	}

	return &rel, nil
}

func (u *Upgrader) download(ctx context.Context, name, fileURL, checksumURL, path string) (bool, error) {
	filereq, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to build download request: %w", err)
	}

	sumreq, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to build checksum download request: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return false, fmt.Errorf("failed to create destination path (%s): %w", path, err)
	}
	defer file.Close()

	resp, err := u.client.Do(filereq)
	if err != nil {
		return false, fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	checksumResp, err := u.client.Do(sumreq)
	if err != nil {
		return false, fmt.Errorf("checksum download request failed: %w", err)
	}
	defer checksumResp.Body.Close()

	bar := progressbar.DefaultBytes(resp.ContentLength+checksumResp.ContentLength, name)

	fileHasher := sha256.New()
	if _, err = io.Copy(io.MultiWriter(file, fileHasher, bar), resp.Body); err != nil {
		return false, fmt.Errorf("failed to download the file: %w", err)
	}
	log.Debug()

	hashsum := hex.EncodeToString(fileHasher.Sum(nil))

	scanner := bufio.NewScanner(checksumResp.Body)
	for scanner.Scan() {
		sums := strings.Fields(scanner.Text())
		log.Debugf("Checking %s %s", sums[0], sums[1])
		if sums[1] == name {
			if sums[0] == hashsum {
				if err = bar.Finish(); err != nil {
					log.Debugf("Progressbar error: %s", err)
				}

				log.Debugf("Match %s %s", sums[0], sums[1])

				return true, nil
			} else {
				return false, nil
			}
		}
	}

	log.Debugf("No matches found for %s %s\n", name, hashsum)

	return false, nil
}
