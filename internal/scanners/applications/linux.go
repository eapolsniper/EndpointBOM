package applications

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// scanLinuxApplications scans for installed applications on Linux
func scanLinuxApplications(cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Standard Linux application locations
	searchPaths := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
	}

	// Add user application folders
	if cfg.ScanAllUsers {
		userProfiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range userProfiles {
				searchPaths = append(searchPaths,
					filepath.Join(profile, ".local", "share", "applications"),
				)
			}
		}
	}

	for _, searchPath := range searchPaths {
		if cfg.IsPathExcluded(searchPath) {
			continue
		}

		apps, err := scanLinuxDirectory(searchPath, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning %s: %v\n", searchPath, err)
			}
			continue
		}
		components = append(components, apps...)
	}

	return components, nil
}

func scanLinuxDirectory(dir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Linux desktop files end with .desktop
		if !strings.HasSuffix(entry.Name(), ".desktop") {
			continue
		}

		desktopPath := filepath.Join(dir, entry.Name())
		if cfg.IsPathExcluded(desktopPath) {
			continue
		}

		// Parse desktop file
		appInfo := parseLinuxDesktopFile(desktopPath)

		comp := scanners.Component{
			Type:       "application",
			Name:       appInfo.Name,
			Version:    appInfo.Version,
			Location:   desktopPath,
			Properties: make(map[string]string),
		}

		if appInfo.Comment != "" {
			comp.Description = appInfo.Comment
		}

		if appInfo.Exec != "" {
			comp.Properties["exec"] = appInfo.Exec
		}

		components = append(components, comp)
	}

	return components, nil
}

type linuxAppInfo struct {
	Name    string
	Version string
	Comment string
	Exec    string
}

func parseLinuxDesktopFile(path string) linuxAppInfo {
	info := linuxAppInfo{}

	data, err := os.ReadFile(path)
	if err != nil {
		return info
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name=") {
			info.Name = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Version=") {
			info.Version = strings.TrimPrefix(line, "Version=")
		} else if strings.HasPrefix(line, "Comment=") {
			info.Comment = strings.TrimPrefix(line, "Comment=")
		} else if strings.HasPrefix(line, "Exec=") {
			info.Exec = strings.TrimPrefix(line, "Exec=")
		}
	}

	// If no name found, use filename
	if info.Name == "" {
		info.Name = filepath.Base(path)
		info.Name = strings.TrimSuffix(info.Name, ".desktop")
	}

	return info
}

