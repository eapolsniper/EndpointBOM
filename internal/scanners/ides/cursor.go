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

// CursorScanner scans for Cursor IDE extensions and MCP servers
type CursorScanner struct{}

func (s *CursorScanner) Name() string {
	return "cursor"
}

func (s *CursorScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("cursor") {
		return nil, nil
	}

	var components []scanners.Component
	var extensionDirs []string

	// Determine Cursor extension directory based on OS
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".cursor", "extensions"))
	case "windows":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".cursor", "extensions"))
	case "linux":
		home, _ := os.UserHomeDir()
		extensionDirs = append(extensionDirs, filepath.Join(home, ".cursor", "extensions"))
	}

	// Scan all user profiles if enabled
	if cfg.ScanAllUsers {
		profiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range profiles {
				extensionDirs = append(extensionDirs, filepath.Join(profile, ".cursor", "extensions"))
			}
		}
	}

	for _, extDir := range extensionDirs {
		if cfg.IsPathExcluded(extDir) {
			continue
		}

		exts, err := scanCursorExtensions(extDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning Cursor extensions in %s: %v\n", extDir, err)
			}
			continue
		}
		components = append(components, exts...)
	}

	// Scan for MCP servers in Cursor
	mcpServers := scanCursorMCPServers(cfg)
	components = append(components, mcpServers...)

	return components, nil
}

func scanCursorExtensions(extensionDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packagePath := filepath.Join(extensionDir, entry.Name(), "package.json")
		
		data, err := os.ReadFile(packagePath)
		if err != nil {
			continue
		}

		var pkgInfo cursorPackageJSON
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
				"ide":       "cursor",
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

func scanCursorMCPServers(cfg *config.Config) []scanners.Component {
	var components []scanners.Component

	// Cursor MCP settings location
	var mcpConfigPaths []string

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		mcpConfigPaths = append(mcpConfigPaths,
			filepath.Join(home, ".cursor", "mcp.json"),
			filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "mcp.json"),
		)
	case "windows":
		home, _ := os.UserHomeDir()
		mcpConfigPaths = append(mcpConfigPaths,
			filepath.Join(home, ".cursor", "mcp.json"),
			filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "globalStorage", "mcp.json"),
		)
	case "linux":
		home, _ := os.UserHomeDir()
		mcpConfigPaths = append(mcpConfigPaths,
			filepath.Join(home, ".cursor", "mcp.json"),
			filepath.Join(home, ".config", "Cursor", "User", "globalStorage", "mcp.json"),
		)
	}

	for _, configPath := range mcpConfigPaths {
		if cfg.IsPathExcluded(configPath) {
			continue
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var mcpConfig map[string]interface{}
		if err := json.Unmarshal(data, &mcpConfig); err != nil {
			continue
		}

		// Parse MCP server configurations
		if servers, ok := mcpConfig["mcpServers"].(map[string]interface{}); ok {
			for serverName, serverConfig := range servers {
				comp := scanners.Component{
					Type:     "mcp-server",
					Name:     serverName,
					Location: configPath,
					Properties: map[string]string{
						"ide": "cursor",
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
					if env, ok := configMap["env"].(map[string]interface{}); ok {
						// Count env vars but don't expose values
						comp.Properties["env_vars_count"] = fmt.Sprintf("%d", len(env))
					}
				}

				components = append(components, comp)
			}
		}
	}

	return components
}

type cursorPackageJSON struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	Description string `json:"description"`
}

