package ides

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// JetBrainsScanner scans for JetBrains IDE plugins
type JetBrainsScanner struct{}

func (s *JetBrainsScanner) Name() string {
	return "jetbrains"
}

func (s *JetBrainsScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("jetbrains") {
		return nil, nil
	}

	var components []scanners.Component

	// Get JetBrains config directories
	configDirs := getJetBrainsConfigDirs()

	for _, configDir := range configDirs {
		if cfg.IsPathExcluded(configDir) {
			continue
		}

		plugins, err := scanJetBrainsPlugins(configDir, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning JetBrains plugins in %s: %v\n", configDir, err)
			}
			continue
		}
		components = append(components, plugins...)
	}

	return components, nil
}

func getJetBrainsConfigDirs() []string {
	var dirs []string
	home, _ := os.UserHomeDir()

	// Common JetBrains product directories
	jetbrainsProducts := []string{
		"IntelliJIdea", "PyCharm", "WebStorm", "PhpStorm",
		"GoLand", "RubyMine", "CLion", "Rider", "DataGrip",
		"AndroidStudio", "AppCode",
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/JetBrains/{Product}{Version}
		baseDir := filepath.Join(home, "Library", "Application Support", "JetBrains")
		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					dirs = append(dirs, filepath.Join(baseDir, entry.Name()))
				}
			}
		}
	case "windows":
		// Windows: %APPDATA%\JetBrains\{Product}{Version}
		baseDir := filepath.Join(home, "AppData", "Roaming", "JetBrains")
		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					dirs = append(dirs, filepath.Join(baseDir, entry.Name()))
				}
			}
		}
	case "linux":
		// Linux: ~/.config/JetBrains/{Product}{Version}
		baseDir := filepath.Join(home, ".config", "JetBrains")
		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					dirs = append(dirs, filepath.Join(baseDir, entry.Name()))
				}
			}
		}
		
		// Also check legacy locations
		for _, product := range jetbrainsProducts {
			legacyDir := filepath.Join(home, fmt.Sprintf(".%s", product))
			if _, err := os.Stat(legacyDir); err == nil {
				dirs = append(dirs, legacyDir)
			}
		}
	}

	return dirs
}

func scanJetBrainsPlugins(configDir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Plugins are typically in config/plugins directory
	pluginsDir := filepath.Join(configDir, "plugins")
	
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	// Extract IDE name from config directory
	ideName := extractIDEName(configDir)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Look for plugin.xml
		pluginXMLPath := filepath.Join(pluginsDir, entry.Name(), "META-INF", "plugin.xml")
		
		pluginInfo, err := parseJetBrainsPluginXML(pluginXMLPath)
		if err != nil {
			// Try alternate location
			pluginXMLPath = filepath.Join(pluginsDir, entry.Name(), "lib", "META-INF", "plugin.xml")
			pluginInfo, err = parseJetBrainsPluginXML(pluginXMLPath)
			if err != nil {
				continue
			}
		}

		comp := scanners.Component{
			Type:        "ide-extension",
			Name:        pluginInfo.Name,
			Version:     pluginInfo.Version,
			Description: pluginInfo.Description,
			Location:    filepath.Join(pluginsDir, entry.Name()),
			Properties: map[string]string{
				"ide":       "jetbrains",
				"ide_name":  ideName,
				"plugin_id": pluginInfo.ID,
			},
		}

		if pluginInfo.Vendor != "" {
			comp.Properties["vendor"] = pluginInfo.Vendor
		}

		components = append(components, comp)
	}

	return components, nil
}

func extractIDEName(configDir string) string {
	baseName := filepath.Base(configDir)
	
	// Remove version numbers
	for i, c := range baseName {
		if c >= '0' && c <= '9' {
			return baseName[:i]
		}
	}
	
	return baseName
}

type jetBrainsPluginXML struct {
	XMLName     xml.Name `xml:"idea-plugin"`
	ID          string   `xml:"id"`
	Name        string   `xml:"name"`
	Version     string   `xml:"version"`
	Description string   `xml:"description"`
	Vendor      string   `xml:"vendor"`
}

func parseJetBrainsPluginXML(path string) (*jetBrainsPluginXML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin jetBrainsPluginXML
	if err := xml.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	// Clean up description (remove extra whitespace)
	plugin.Description = strings.TrimSpace(plugin.Description)

	return &plugin, nil
}

