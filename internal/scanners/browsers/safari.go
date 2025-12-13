package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// SafariScanner scans for Safari browser extensions (macOS only)
type SafariScanner struct{}

func (s *SafariScanner) Name() string {
	return "safari-extensions"
}

func (s *SafariScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("safari-extensions") {
		return nil, nil
	}

	// Safari extensions only exist on macOS
	if runtime.GOOS != "darwin" {
		return nil, nil
	}

	var components []scanners.Component
	var extensionDirs []string

	home, _ := os.UserHomeDir()
	
	// Safari extensions can be in multiple locations
	extensionDirs = append(extensionDirs,
		filepath.Join(home, "Library", "Safari", "Extensions"),
		filepath.Join(home, "Library", "Containers"),
		"/Library/Application Support/App Store",
	)

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				extensionDirs = append(extensionDirs,
					filepath.Join(profile, "Library", "Safari", "Extensions"),
					filepath.Join(profile, "Library", "Containers"),
				)
			}
		}
	}

	for _, extDir := range extensionDirs {
		if cfg.IsPathExcluded(extDir) {
			continue
		}

		exts, err := scanSafariExtensions(extDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Safari extensions in %s: %v\n", extDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	return components, nil
}

func scanSafariExtensions(extensionDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Safari extensions are in .safariextension or .appex directories
		// Also check container bundles for Safari App Extensions
		extName := entry.Name()
		
		if strings.HasSuffix(extName, ".safariextension") {
			// Legacy Safari extension
			extPath := filepath.Join(extensionDir, extName)
			comp := scanLegacySafariExtension(extPath)
			if comp.Name != "" {
				components = append(components, comp)
			}
		} else if strings.Contains(extName, "Safari") {
			// Modern Safari App Extension (inside container)
			extPath := filepath.Join(extensionDir, extName)
			comp := scanModernSafariExtension(extPath)
			if comp.Name != "" {
				components = append(components, comp)
			}
		}
	}

	return components, nil
}

func scanLegacySafariExtension(extPath string) scanners.Component {
	// Legacy extensions have Info.plist
	// plistPath := filepath.Join(extPath, "Info.plist")
	
	// For simplicity, we'll try to read a manifest.json if it exists
	manifestPath := filepath.Join(extPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// Try reading basic info from directory name
		return scanners.Component{
			Type:     "browser-extension",
			Name:     filepath.Base(extPath),
			Version:  "unknown",
			Location: extPath,
			Properties: map[string]string{
				"browser": "safari",
				"type":    "legacy",
			},
		}
	}

	var manifest safariManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return scanners.Component{}
	}

	return scanners.Component{
		Type:        "browser-extension",
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Location:    extPath,
		Properties: map[string]string{
			"browser": "safari",
			"type":    "legacy",
		},
	}
}

func scanModernSafariExtension(containerPath string) scanners.Component {
	// Modern Safari extensions are inside app containers
	// Look for .appex bundles
	err := filepath.Walk(containerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}
		
		if info.IsDir() && strings.HasSuffix(info.Name(), ".appex") {
			// Found an app extension
			manifestPath := filepath.Join(path, "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil
			}

			var manifest safariManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil
			}

			// We found one, stop walking
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return scanners.Component{}
	}

	// Return basic info about the container
	return scanners.Component{
		Type:     "browser-extension",
		Name:     filepath.Base(containerPath),
		Version:  "unknown",
		Location: containerPath,
		Properties: map[string]string{
			"browser": "safari",
			"type":    "modern",
		},
	}
}

type safariManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

