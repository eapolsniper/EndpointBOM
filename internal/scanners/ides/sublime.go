package ides

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// SublimeScanner scans for Sublime Text packages
type SublimeScanner struct{}

func (s *SublimeScanner) Name() string {
	return "sublime"
}

func (s *SublimeScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("sublime") {
		return nil, nil
	}

	var components []scanners.Component
	var packageDirs []string

	home, _ := os.UserHomeDir()

	// Determine Sublime Text packages directory based on OS
	switch runtime.GOOS {
	case "darwin":
		packageDirs = append(packageDirs,
			filepath.Join(home, "Library", "Application Support", "Sublime Text", "Packages"),
			filepath.Join(home, "Library", "Application Support", "Sublime Text 3", "Packages"),
		)
	case "windows":
		packageDirs = append(packageDirs,
			filepath.Join(home, "AppData", "Roaming", "Sublime Text", "Packages"),
			filepath.Join(home, "AppData", "Roaming", "Sublime Text 3", "Packages"),
		)
	case "linux":
		packageDirs = append(packageDirs,
			filepath.Join(home, ".config", "sublime-text", "Packages"),
			filepath.Join(home, ".config", "sublime-text-3", "Packages"),
		)
	}

	for _, pkgDir := range packageDirs {
		if cfg.IsPathExcluded(pkgDir) {
			continue
		}

		pkgs, err := scanSublimePackages(pkgDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Sublime packages in %s: %v\n", pkgDir, err)
			}
			continue
		}
		components = append(components, pkgs...)
	}

	return components, nil
}

func scanSublimePackages(packageDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip built-in packages
		if entry.Name() == "User" || entry.Name() == "Default" {
			continue
		}

		packagePath := filepath.Join(packageDir, entry.Name())
		
		// Try to read package.json if it exists
		packageJSONPath := filepath.Join(packagePath, "package.json")
		pkgInfo := sublimePackageInfo{Name: entry.Name()}
		
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			json.Unmarshal(data, &pkgInfo)
		}

		comp := scanners.Component{
			Type:        "ide-extension",
			Name:        pkgInfo.Name,
			Version:     pkgInfo.Version,
			Description: pkgInfo.Description,
			Location:    packagePath,
			Properties: map[string]string{
				"ide": "sublime",
			},
		}

		if pkgInfo.Author != "" {
			comp.Properties["author"] = pkgInfo.Author
		}

		components = append(components, comp)
	}

	return components, nil
}

type sublimePackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

