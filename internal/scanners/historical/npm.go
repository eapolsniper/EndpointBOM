package historical

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// NPMHistoricalScanner scans npm installation logs for historical package data
type NPMHistoricalScanner struct{}

// Name returns the scanner name
func (s *NPMHistoricalScanner) Name() string {
	return "npm-historical"
}

// Scan performs the historical scan (best effort - parse simple log entries)
func (s *NPMHistoricalScanner) Scan(cfg *config.Config) ([]Component, error) {
	if !cfg.IncludeHistorical {
		return []Component{}, nil
	}

	var components []Component
	lookbackDuration := time.Duration(cfg.HistoricalLookbackDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-lookbackDuration)

	// Find npm log directory
	logDir := filepath.Join(os.Getenv("HOME"), ".npm", "_logs")
	if os.Getenv("USERPROFILE") != "" { // Windows
		logDir = filepath.Join(os.Getenv("USERPROFILE"), ".npm", "_logs")
	}

	if _, err := os.Stat(logDir); err != nil {
		return components, nil // No logs, skip silently
	}

	// Read log files within lookback period
	logs, err := filepath.Glob(filepath.Join(logDir, "*-debug*.log"))
	if err != nil {
		return components, nil
	}

	seen := make(map[string]bool) // Deduplicate packages

	for _, logFile := range logs {
		fileInfo, err := os.Stat(logFile)
		if err != nil || !fileInfo.ModTime().After(cutoffTime) {
			continue
		}

		// Simple parse - just look for "npm install package@version" lines
		parsed := parseNPMLogSimple(logFile, fileInfo.ModTime())
		for _, comp := range parsed {
			key := comp.Name + "@" + comp.Version
			if !seen[key] {
				components = append(components, comp)
				seen[key] = true
			}
		}
	}

	return components, nil
}

// GetLogFiles returns relevant log files for archiving
func (s *NPMHistoricalScanner) GetLogFiles(cfg *config.Config) ([]string, error) {
	if !cfg.IncludeRawLogs {
		return []string{}, nil
	}

	lookbackDuration := time.Duration(cfg.HistoricalLookbackDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-lookbackDuration)

	logDir := filepath.Join(os.Getenv("HOME"), ".npm", "_logs")
	if os.Getenv("USERPROFILE") != "" {
		logDir = filepath.Join(os.Getenv("USERPROFILE"), ".npm", "_logs")
	}

	logs, err := filepath.Glob(filepath.Join(logDir, "*-debug*.log"))
	if err != nil {
		return []string{}, nil
	}

	var relevantLogs []string
	for _, log := range logs {
		fileInfo, err := os.Stat(log)
		if err == nil && fileInfo.ModTime().After(cutoffTime) {
			relevantLogs = append(relevantLogs, log)
		}
	}

	return relevantLogs, nil
}

// parseNPMLogSimple does simple best-effort parsing
func parseNPMLogSimple(logPath string, logTime time.Time) []Component {
	var components []Component

	file, err := os.Open(logPath)
	if err != nil {
		return components
	}
	defer file.Close()

	// Simple regex to find package installations
	// Matches lines like: "npm install react@18.2.0" or "added 5 packages"
	installRegex := regexp.MustCompile(`(?:npm (?:install|i|add)\s+|added\s+\d+\s+packages?)([a-zA-Z0-9@/_-]+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		matches := installRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			pkg := strings.TrimSpace(matches[1])
			if pkg == "" || strings.HasPrefix(pkg, "-") {
				continue
			}

			name := pkg
			version := ""

			// Split package@version
			if strings.Contains(pkg, "@") && !strings.HasPrefix(pkg, "@") {
				parts := strings.Split(pkg, "@")
				name = parts[0]
				if len(parts) > 1 {
					version = parts[1]
				}
			}

			if name != "" {
				components = append(components, scanners.Component{
					Type:           "library",
					Name:           name,
					Version:        version,
					PackageManager: "npm",
					Location:       logPath,
					Properties: map[string]string{
						"install_date": logTime.Format(time.RFC3339),
						"install_type": getInstallType(logTime),
						"source":       "npm_log",
						"log_file":     filepath.Base(logPath),
					},
				})
			}
		}
	}

	return components
}

// getInstallType returns "historical" - will be overridden to "current" during deduplication
// if the package is still installed
func getInstallType(installTime time.Time) string {
	// All items from logs start as "historical"
	// Main scanner will change to "current" if package is still installed
	return "historical"
}

// Component is a local type alias
type Component = scanners.Component
