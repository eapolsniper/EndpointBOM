package applications

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// scanDarwinApplications scans for installed applications on macOS
func scanDarwinApplications(cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	// Standard macOS application locations
	searchPaths := []string{
		"/Applications",
		"/System/Applications",
		"/System/Library/CoreServices",
	}

	// Add user Applications folders
	if cfg.ScanAllUsers {
		userProfiles, err := getUserProfiles()
		if err == nil {
			for _, profile := range userProfiles {
				searchPaths = append(searchPaths, filepath.Join(profile, "Applications"))
			}
		}
	}

	for _, searchPath := range searchPaths {
		if cfg.IsPathExcluded(searchPath) {
			continue
		}

		apps, err := scanMacOSDirectory(searchPath, cfg)
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

func scanMacOSDirectory(dir string, cfg *config.Config) ([]scanners.Component, error) {
	var components []scanners.Component

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// macOS applications end with .app
		if !strings.HasSuffix(entry.Name(), ".app") {
			continue
		}

		appPath := filepath.Join(dir, entry.Name())
		if cfg.IsPathExcluded(appPath) {
			continue
		}

		// Read Info.plist for version information
		infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
		appInfo := parseMacOSAppInfo(infoPlistPath)

		comp := scanners.Component{
			Type:     "application",
			Name:     strings.TrimSuffix(entry.Name(), ".app"),
			Version:  appInfo.Version,
			Location: appPath,
			Properties: map[string]string{
				"bundle_identifier": appInfo.BundleIdentifier,
			},
		}

		if appInfo.DisplayName != "" {
			comp.Name = appInfo.DisplayName
		}

		components = append(components, comp)
	}

	return components, nil
}

type macOSAppInfo struct {
	BundleIdentifier string
	Version          string
	DisplayName      string
}

func parseMacOSAppInfo(plistPath string) macOSAppInfo {
	info := macOSAppInfo{}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		return info
	}

	content := string(data)

	// Simple plist parsing - extract key values
	info.BundleIdentifier = extractPlistValue(content, "CFBundleIdentifier")
	info.Version = extractPlistValue(content, "CFBundleShortVersionString")
	if info.Version == "" {
		info.Version = extractPlistValue(content, "CFBundleVersion")
	}
	info.DisplayName = extractPlistValue(content, "CFBundleDisplayName")

	return info
}

func extractPlistValue(content, key string) string {
	// Look for <key>KeyName</key><string>Value</string>
	keyTag := "<key>" + key + "</key>"
	idx := strings.Index(content, keyTag)
	if idx == -1 {
		return ""
	}

	// Find the next <string> tag
	startTag := "<string>"
	endTag := "</string>"
	
	remainder := content[idx+len(keyTag):]
	startIdx := strings.Index(remainder, startTag)
	if startIdx == -1 {
		return ""
	}
	
	startIdx += len(startTag)
	endIdx := strings.Index(remainder[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return remainder[startIdx : startIdx+endIdx]
}

