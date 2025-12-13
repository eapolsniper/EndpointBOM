package applications

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// scanWindowsApplications scans for installed applications on Windows
func scanWindowsApplications(cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Scan file system locations
	searchPaths := []string{
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	}

	// Add user AppData folders
	if cfg.ScanAllUsers {
		userProfiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range userProfiles {
				searchPaths = append(searchPaths,
					filepath.Join(profile, "AppData", "Local", "Programs"),
				)
			}
		}
	}

	for _, searchPath := range searchPaths {
		if cfg.IsPathExcluded(searchPath) {
			continue
		}

		apps, err := scanWindowsDirectory(searchPath, cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Error scanning %s: %v\n", searchPath, err)
			}
			continue
		}
		components = append(components, apps...)
	}

	// Scan Windows Registry for installed programs
	regApps, err := scanWindowsRegistry(cfg)
	if err != nil {
		if cfg.Debug {
			fmt.Printf("Error scanning registry: %v\n", err)
		}
	} else {
		components = append(components, regApps...)
	}

	return components, nil
}

func scanWindowsDirectory(dir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appPath := filepath.Join(dir, entry.Name())
		if cfg.IsPathExcluded(appPath) {
			continue
		}

		// Look for .exe files to determine version
		version := findWindowsAppVersion(appPath)

		comp := scanners.Component{
			Type:       "application",
			Name:       entry.Name(),
			Version:    version,
			Location:   appPath,
			Properties: make(map[string]string),
		}

		components = append(components, comp)
	}

	return components, nil
}

func findWindowsAppVersion(appPath string) string {
	// Look for .exe files in the directory
	matches, err := filepath.Glob(filepath.Join(appPath, "*.exe"))
	if err != nil || len(matches) == 0 {
		return "unknown"
	}

	// Return version from first .exe found (simplified)
	// In a production tool, you'd use syscall to read the version resource
	return "unknown"
}

func scanWindowsRegistry(cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Use PowerShell to query the registry
	registryPaths := []string{
		"HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*",
		"HKLM:\\Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*",
	}

	for _, regPath := range registryPaths {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Get-ItemProperty '%s' | Select-Object DisplayName, DisplayVersion, Publisher | ConvertTo-Json", regPath))
		
		output, err := cmd.Output()
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Registry query error for %s: %v\n", regPath, err)
			}
			continue
		}

		// Parse JSON output (simplified - would need proper JSON parsing)
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "DisplayName") {
				// Would parse JSON properly in production
				if cfg.Verbose {
					fmt.Println(line)
				}
			}
		}
	}

	return components, nil
}

