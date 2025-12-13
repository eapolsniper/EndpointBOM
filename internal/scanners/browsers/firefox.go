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

// FirefoxScanner scans for Firefox browser extensions
type FirefoxScanner struct{}

func (s *FirefoxScanner) Name() string {
	return "firefox-extensions"
}

func (s *FirefoxScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("firefox-extensions") {
		return nil, nil
	}

	var components []scanners.Component
	var profileDirs []string

	// Get Firefox profile directories based on OS
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		profileDirs = append(profileDirs,
			filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"),
		)
	case "windows":
		home, _ := os.UserHomeDir()
		profileDirs = append(profileDirs,
			filepath.Join(home, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles"),
		)
	case "linux":
		home, _ := os.UserHomeDir()
		profileDirs = append(profileDirs,
			filepath.Join(home, ".mozilla", "firefox"),
		)
	}

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				switch runtime.GOOS {
				case "darwin":
					profileDirs = append(profileDirs,
						filepath.Join(profile, "Library", "Application Support", "Firefox", "Profiles"),
					)
				case "windows":
					profileDirs = append(profileDirs,
						filepath.Join(profile, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles"),
					)
				case "linux":
					profileDirs = append(profileDirs,
						filepath.Join(profile, ".mozilla", "firefox"),
					)
				}
			}
		}
	}

	for _, profileDir := range profileDirs {
		if cfg.IsPathExcluded(profileDir) {
			continue
		}

		exts, err := scanFirefoxExtensions(profileDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Firefox extensions in %s: %v\n", profileDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	return components, nil
}

func scanFirefoxExtensions(profilesDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Firefox can have multiple profiles
	profiles, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		if !profile.IsDir() {
			continue
		}

		// Extensions are in the extensions subdirectory
		extensionsDir := filepath.Join(profilesDir, profile.Name(), "extensions")
		
		entries, err := os.ReadDir(extensionsDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			// Firefox extensions are .xpi files or directories
			var manifestPath string
			
			if entry.IsDir() {
				manifestPath = filepath.Join(extensionsDir, entry.Name(), "manifest.json")
			} else if filepath.Ext(entry.Name()) == ".xpi" {
				// .xpi files are zip archives, would need to extract
				// For now, we'll skip these and only scan unpacked extensions
				continue
			} else {
				continue
			}

			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var manifest firefoxManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				continue
			}

			comp := scanners.Component{
				Type:        "browser-extension",
				Name:        manifest.Name,
				Version:     manifest.Version,
				Description: manifest.Description,
				Location:    filepath.Join(extensionsDir, entry.Name()),
				Properties: map[string]string{
					"browser":    "firefox",
					"addon_id":   manifest.Applications.Gecko.ID,
					"profile":    profile.Name(),
				},
			}

			// Add permissions
			if len(manifest.Permissions) > 0 {
				permStr := ""
				for i, perm := range manifest.Permissions {
					if i > 0 {
						permStr += ", "
					}
					permStr += perm
				}
				comp.Properties["permissions"] = permStr
			}

			components = append(components, comp)
		}
	}

	return components, nil
}

type firefoxManifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Permissions  []string `json:"permissions"`
	Applications struct {
		Gecko struct {
			ID string `json:"id"`
		} `json:"gecko"`
	} `json:"applications"`
}

