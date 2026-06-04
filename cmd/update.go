package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

var Version = "dev"

const githubRepo = "ashuraits/commito"

func checkForUpdate() {
	if Version == "dev" {
		return
	}
	go func() {
		latest, err := fetchLatestVersion()
		if err != nil || latest == "" || latest == Version {
			return
		}
		fmt.Fprintf(os.Stderr, "\nUpdate available: %s → %s\nRun: commito update\n\n", Version, latest)
	}()
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + githubRepo + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func selfUpdate() error {
	fmt.Println("Checking for updates...")
	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("could not fetch latest version: %w", err)
	}

	if latest == Version {
		fmt.Printf("Already up to date (%s)\n", Version)
		return nil
	}

	fmt.Printf("Updating %s → %s\n", Version, latest)

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/commito-%s-%s",
		githubRepo, latest, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find current executable: %w", err)
	}

	tmp, err := os.CreateTemp("", "commito-update-*")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	if err := os.Rename(tmp.Name(), exe); err != nil {
		return fmt.Errorf("could not replace binary (try with sudo): %w", err)
	}

	fmt.Printf("Updated to %s — restart commito to apply\n", latest)
	return nil
}
