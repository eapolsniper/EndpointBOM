package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// EdgeScanner scans for Microsoft Edge browser extensions
type EdgeScanner struct{}

func (s *EdgeScanner) Name() string {
	return "edge-extensions"
}

func (s *EdgeScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("edge-extensions") {
		return nil, nil
	}

	var components []scanners.Component
	var extensionDirs []string

	// Get Edge extension directories based on OS
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs,
			filepath.Join(home, "Library", "Application Support", "Microsoft Edge", "Default", "Extensions"),
			filepath.Join(home, "Library", "Application Support", "Microsoft Edge", "Profile 1", "Extensions"),
		)
	case "windows":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs,
			filepath.Join(home, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions"),
			filepath.Join(home, "AppData", "Local", "Microsoft", "Edge", "User Data", "Profile 1", "Extensions"),
		)
	case "linux":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs,
			filepath.Join(home, ".config", "microsoft-edge", "Default", "Extensions"),
			filepath.Join(home, ".config", "microsoft-edge", "Profile 1", "Extensions"),
		)
	}

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				switch runtime.GOOS {
				case "darwin":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, "Library", "Application Support", "Microsoft Edge", "Default", "Extensions"),
					)
				case "windows":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions"),
					)
				case "linux":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, ".config", "microsoft-edge", "Default", "Extensions"),
					)
				}
			}
		}
	}

	for _, extDir := range extensionDirs {
		if cfg.IsPathExcluded(extDir) {
			continue
		}

		exts, err := scanEdgeExtensions(extDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Edge extensions in %s: %v\n", extDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	return components, nil
}

func scanEdgeExtensions(extensionDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionID := entry.Name()
		extensionPath := filepath.Join(extensionDir, extensionID)

		// Each extension can have multiple versions
		versions, err := os.ReadDir(extensionPath)
		if err != nil {
			continue
		}

		// Get the latest version
		var latestVersion string
		var latestManifest edgeManifest
		
		for _, versionDir := range versions {
			if !versionDir.IsDir() {
				continue
			}

			manifestPath := filepath.Join(extensionPath, versionDir.Name(), "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var manifest edgeManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				continue
			}

			latestVersion = versionDir.Name()
			latestManifest = manifest
		}

		if latestVersion != "" {
			comp := scanners.Component{
				Type:        "browser-extension",
				Name:        latestManifest.Name,
				Version:     latestManifest.Version,
				Description: latestManifest.Description,
				Location:    extensionPath,
				Properties: map[string]string{
					"browser":      "edge",
					"extension_id": extensionID,
					"manifest_version": fmt.Sprintf("%d", latestManifest.ManifestVersion),
				},
			}

			// Add permissions (important for security analysis)
			if len(latestManifest.Permissions) > 0 {
				permStr := ""
				for i, perm := range latestManifest.Permissions {
					if i > 0 {
						permStr += ", "
					}
					permStr += perm
				}
				comp.Properties["permissions"] = permStr
			}

			// Add host permissions
			if len(latestManifest.HostPermissions) > 0 {
				hostStr := ""
				for i, host := range latestManifest.HostPermissions {
					if i > 0 {
						hostStr += ", "
					}
					hostStr += host
				}
				comp.Properties["host_permissions"] = hostStr
			}

			components = append(components, comp)
		}
	}

	return components, nil
}

type edgeManifest struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Description     string   `json:"description"`
	ManifestVersion int      `json:"manifest_version"`
	Permissions     []string `json:"permissions"`
	HostPermissions []string `json:"host_permissions"`
}

