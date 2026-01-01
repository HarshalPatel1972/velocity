package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitHubOwner = "HarshalPatel1972"
	GitHubRepo  = "velocity"
	APIEndpoint = "https://api.github.com/repos/" + GitHubOwner + "/" + GitHubRepo + "/releases/latest"
)

// Asset represents a downloadable file from a GitHub release
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available    bool
	NewVersion   string
	DownloadURL  string
	InstallerName string
}

// CheckForUpdates queries GitHub API for the latest release
func CheckForUpdates(currentVersion string) (*UpdateInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GitHub API requires User-Agent
	req.Header.Set("User-Agent", "Velocity-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No releases yet
		return &UpdateInfo{Available: false}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Compare versions (simple string comparison for vX.Y.Z format)
	if !isNewerVersion(release.TagName, currentVersion) {
		return &UpdateInfo{Available: false}, nil
	}

	// Find the installer asset
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, "Velocity_Setup") && strings.HasSuffix(asset.Name, ".exe") {
			return &UpdateInfo{
				Available:     true,
				NewVersion:    release.TagName,
				DownloadURL:   asset.BrowserDownloadURL,
				InstallerName: asset.Name,
			}, nil
		}
	}

	return nil, fmt.Errorf("no installer found in release %s", release.TagName)
}

// isNewerVersion compares two version strings (v1.0.0 format)
func isNewerVersion(remote, local string) bool {
	// Strip 'v' prefix if present
	remote = strings.TrimPrefix(remote, "v")
	local = strings.TrimPrefix(local, "v")

	remoteParts := strings.Split(remote, ".")
	localParts := strings.Split(local, ".")

	for i := 0; i < len(remoteParts) && i < len(localParts); i++ {
		if remoteParts[i] > localParts[i] {
			return true
		}
		if remoteParts[i] < localParts[i] {
			return false
		}
	}

	return len(remoteParts) > len(localParts)
}

// DownloadInstaller downloads the installer to temp folder
func DownloadInstaller(url, filename string) (string, error) {
	tempDir := os.TempDir()
	destPath := filepath.Join(tempDir, filename)

	// Create the file
	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Write to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save installer: %w", err)
	}

	return destPath, nil
}

// LaunchInstallerAndExit starts the installer and exits the current process
func LaunchInstallerAndExit(installerPath string) error {
	cmd := exec.Command(installerPath, "/SILENT")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	// Critical: Exit immediately to unlock the executable
	os.Exit(0)
	return nil // Never reached
}
