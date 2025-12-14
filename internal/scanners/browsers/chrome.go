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

// ChromeScanner scans for Chrome browser extensions
type ChromeScanner struct{}

func (s *ChromeScanner) Name() string {
	return "chrome-extensions"
}

func (s *ChromeScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("chrome-extensions") {
		return nil, nil
	}

	var components []scanners.Component
	var extensionDirs []string

	// Get Chrome base directory
	home, _ := os.UserHomeDir()
	var chromeBase string

	switch runtime.GOOS {
	case "darwin":
		chromeBase = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	case "windows":
		chromeBase = filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data")
	case "linux":
		chromeBase = filepath.Join(home, ".config", "google-chrome")
	}

	// Discover all Chrome profiles dynamically
	profiles, err := discoverChromeProfiles(chromeBase)
	if err != nil && cfg.Debug {
		fmt.Printf("Could not discover Chrome profiles: %v\n", err)
	}

	// Add extension directories for each discovered profile
	profileExtDirs := make(map[string]string) // map[extensionDir]profileName
	for _, profile := range profiles {
		extDir := filepath.Join(chromeBase, profile, "Extensions")
		extensionDirs = append(extensionDirs, extDir)
		profileExtDirs[extDir] = profile
	}

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				switch runtime.GOOS {
				case "darwin":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, "Library", "Application Support", "Google", "Chrome", "Default", "Extensions"),
					)
				case "windows":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Extensions"),
					)
				case "linux":
					extensionDirs = append(extensionDirs,
						filepath.Join(profile, ".config", "google-chrome", "Default", "Extensions"),
					)
				}
			}
		}
	}

	for _, extDir := range extensionDirs {
		if cfg.IsPathExcluded(extDir) {
			continue
		}

		profileName := profileExtDirs[extDir]
		exts, err := scanChromeExtensions(extDir, profileName, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Chrome extensions in %s: %v\n", extDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	return components, nil
}

func scanChromeExtensions(extensionDir string, profileName string, cfg *config.Config) ([]scanners.Component, error) {
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

		// Get the latest version (usually only one, but could be multiple during update)
		var latestVersion string
		var latestManifest chromeManifest
		
		for _, versionDir := range versions {
			if !versionDir.IsDir() {
				continue
			}

			manifestPath := filepath.Join(extensionPath, versionDir.Name(), "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var manifest chromeManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				continue
			}

			latestVersion = versionDir.Name()
			latestManifest = manifest
		}

		if latestVersion != "" {
			// Resolve internationalized name if needed
			displayName := latestManifest.Name
			if len(displayName) > 6 && displayName[:6] == "__MSG_" {
				// Try to resolve from _locales
				resolvedName := resolveI18nName(extensionPath, latestVersion, latestManifest.Name)
				if resolvedName != "" {
					displayName = resolvedName
				}
			}

			comp := scanners.Component{
				Type:        "browser-extension",
				Name:        displayName,
				Version:     latestManifest.Version,
				Description: latestManifest.Description,
				Location:    extensionPath,
				Properties: map[string]string{
					"browser":          "chrome",
					"extension_id":     extensionID,
					"manifest_version": fmt.Sprintf("%d", latestManifest.ManifestVersion),
					"profile":          profileName,
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

			// Add host permissions (which sites can the extension access)
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

type chromeManifest struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Description     string   `json:"description"`
	ManifestVersion int      `json:"manifest_version"`
	Permissions     []string `json:"permissions"`
	HostPermissions []string `json:"host_permissions"`
}

// discoverChromeProfiles finds all Chrome profile directories
func discoverChromeProfiles(chromeBase string) ([]string, error) {
	var profiles []string

	entries, err := os.ReadDir(chromeBase)
	if err != nil {
		return profiles, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Include Default and any Profile directories
		if name == "Default" || (len(name) > 7 && name[:7] == "Profile") {
			profiles = append(profiles, name)
		}
	}

	return profiles, nil
}

// resolveI18nName resolves internationalized extension names from _locales
func resolveI18nName(extensionPath, version, msgKey string) string {
	// Extract the message key from __MSG_key__
	if len(msgKey) < 8 {
		return ""
	}
	key := msgKey[6 : len(msgKey)-2] // Remove __MSG_ and __

	// Try to read from _locales/en/messages.json (default to English)
	localesPath := filepath.Join(extensionPath, version, "_locales", "en", "messages.json")
	data, err := os.ReadFile(localesPath)
	if err != nil {
		// Try en_US as fallback
		localesPath = filepath.Join(extensionPath, version, "_locales", "en_US", "messages.json")
		data, err = os.ReadFile(localesPath)
		if err != nil {
			return ""
		}
	}

	var messages map[string]map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return ""
	}

	if msg, exists := messages[key]; exists {
		if message, ok := msg["message"].(string); ok {
			return message
		}
	}

	return ""
}

