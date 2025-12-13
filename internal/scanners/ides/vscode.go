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

// VSCodeScanner scans for VSCode extensions
type VSCodeScanner struct{}

func (s *VSCodeScanner) Name() string {
	return "vscode"
}

func (s *VSCodeScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("vscode") {
		return nil, nil
	}

	var components []scanners.Component
	var extensionDirs []string

	// Determine VSCode extension directory based on OS
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".vscode", "extensions"))
	case "windows":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".vscode", "extensions"))
	case "linux":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".vscode", "extensions"))
	}

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				extensionDirs = append(extensionDirs, filepath.Join(profile, ".vscode", "extensions"))
			}
		}
	}

	for _, extDir := range extensionDirs {
		if cfg.IsPathExcluded(extDir) {
			continue
		}

		exts, err := scanVSCodeExtensions(extDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning VSCode extensions in %s: %v\n", extDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	// Also scan for MCP servers in VSCode settings
	mcpServers := scanVSCodeMCPServers(cfg)
	components = append(components, mcpServers...)

	return components, nil
}

func scanVSCodeExtensions(extensionDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// VSCode extensions are typically named publisher.extension-version
		packagePath := filepath.Join(extensionDir, entry.Name(), "package.json")
		
		data, err := os.ReadFile(packagePath)
		if err != nil {
			continue
		}

		var pkgInfo vscodePackageJSON
		if err := json.Unmarshal(data, &pkgInfo); err != nil {
			continue
		}

		comp := scanners.Component{
			Type:        "ide-extension",
			Name:        pkgInfo.Name,
			Version:     pkgInfo.Version,
			Description: pkgInfo.Description,
			Location:    filepath.Join(extensionDir, entry.Name()),
			Properties: map[string]string{
				"ide":       "vscode",
				"publisher": pkgInfo.Publisher,
			},
		}

		if pkgInfo.DisplayName != "" {
			comp.Properties["display_name"] = pkgInfo.DisplayName
		}

		components = append(components, comp)
	}

	return components, nil
}

func scanVSCodeMCPServers(cfg *config.Config) []scanners.Component {
	var components []scanners.Component

	// VSCode MCP settings are typically in settings.json
	var settingsPaths []string

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		settingsPaths = append(settingsPaths,
			filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"),
		)
	case "windows":
		home, _ := os.UserHomeDir()
		settingsPaths = append(settingsPaths,
			filepath.Join(home, "AppData", "Roaming", "Code", "User", "settings.json"),
		)
	case "linux":
		home, _ := os.UserHomeDir()
		settingsPaths = append(settingsPaths,
			filepath.Join(home, ".config", "Code", "User", "settings.json"),
		)
	}

	for _, settingsPath := range settingsPaths {
		if cfg.IsPathExcluded(settingsPath) {
			continue
		}

		data, err := os.ReadFile(settingsPath)
		if err != nil {
			continue
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			continue
		}

		// Look for MCP server configurations
		// This is a placeholder - actual MCP config structure may vary
		if mcpConfig, ok := settings["mcp.servers"].(map[string]interface{}); ok {
			for serverName, serverConfig := range mcpConfig {
				comp := scanners.Component{
					Type:     "mcp-server",
					Name:     serverName,
					Location: settingsPath,
					Properties: map[string]string{
						"ide": "vscode",
					},
				}

				// Extract server details without secrets
				if configMap, ok := serverConfig.(map[string]interface{}); ok {
					if command, ok := configMap["command"].(string); ok {
						comp.Properties["command"] = command
					}
					if args, ok := configMap["args"].([]interface{}); ok {
						comp.Properties["args_count"] = fmt.Sprintf("%d", len(args))
					}
				}

				components = append(components, comp)
			}
		}
	}

	return components
}

type vscodePackageJSON struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	Description string `json:"description"`
}

