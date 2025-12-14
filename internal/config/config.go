package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// ExcludePaths are paths to exclude from scanning
	ExcludePaths []string `yaml:"exclude_paths"`

	// DisabledScanners lists scanners to disable
	DisabledScanners []string `yaml:"disabled_scanners"`

	// RequireAdmin forces admin/root privileges
	RequireAdmin bool `yaml:"require_admin"`

	// ScanAllUsers scans all user profiles (requires admin)
	ScanAllUsers bool `yaml:"scan_all_users"`

	// OutputDir specifies where to save SBOM files
	OutputDir string `yaml:"output_dir"`

	// Debug enables debug logging
	Debug bool `yaml:"debug"`

	// Verbose enables verbose output
	Verbose bool `yaml:"verbose"`

	// DisablePublicIP disables public IP address gathering from external services (default: true)
	DisablePublicIP bool `yaml:"disable_public_ip"`

	// HistoricalLookbackDays specifies how many days back to look for historical installations
	HistoricalLookbackDays int `yaml:"historical_lookback_days"`

	// IncludeHistorical enables historical package tracking from logs
	IncludeHistorical bool `yaml:"include_historical"`

	// IncludeRawLogs includes raw log files in the zip archive
	IncludeRawLogs bool `yaml:"include_raw_logs"`

	// CreateZipArchive creates a zip file with SBOMs and logs
	CreateZipArchive bool `yaml:"create_zip_archive"`
}

// DefaultConfig returns a Config with default values including sensitive path exclusions
func DefaultConfig() *Config {
	return &Config{
		// Add default exclusions for sensitive paths
		ExcludePaths: []string{
			// Unix/Linux sensitive paths
			"/etc/shadow",
			"/etc/gshadow",
			"/root/.ssh",
			"/var/log/audit",
			// User home directory sensitive paths (will be expanded per user)
			".ssh",
			".gnupg",
			".aws",
			".kube",
			".docker",
			".netrc",
			".git-credentials",
			// Windows sensitive paths
			"C:\\Windows\\System32\\config",
			"C:\\Windows\\repair",
		},
		DisabledScanners: []string{
			// Browser scanners disabled by default to avoid macOS TCC permission popups
			// Enable these if your environment has Full Disk Access configured via MDM
			"chrome-extensions",
			"firefox-extensions",
			"edge-extensions",
			"safari-extensions",
		},
		RequireAdmin:           false, // Don't require admin - auto-adjust based on privileges
		ScanAllUsers:           true,  // Default to true, will auto-adjust if not admin
		OutputDir:              "",    // Will be set to scans/ by main
		Debug:                  false,
		Verbose:                false,
		DisablePublicIP:        true,  // Default to true - don't fetch public IP from external services
		HistoricalLookbackDays: 30,
		IncludeHistorical:      true,
		IncludeRawLogs:         true,
		CreateZipArchive:       true,
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// IsScannerDisabled checks if a scanner is disabled
func (c *Config) IsScannerDisabled(scanner string) bool {
	for _, disabled := range c.DisabledScanners {
		if disabled == scanner {
			return true
		}
	}
	return false
}

// IsPathExcluded checks if a path should be excluded (improved with normalization)
func (c *Config) IsPathExcluded(path string) bool {
	// Normalize the input path
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute path if possible
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		// Fallback to cleaned path if abs fails
		absPath = cleanPath
	}
	
	for _, excluded := range c.ExcludePaths {
		// Normalize the excluded path
		cleanExcluded := filepath.Clean(excluded)
		absExcluded, err := filepath.Abs(cleanExcluded)
		if err != nil {
			absExcluded = cleanExcluded
		}
		
		// Case-insensitive comparison on Windows
		comparePath := absPath
		compareExcluded := absExcluded
		if runtime.GOOS == "windows" {
			comparePath = strings.ToLower(absPath)
			compareExcluded = strings.ToLower(absExcluded)
		}
		
		// Check exact match
		if comparePath == compareExcluded {
			return true
		}
		
		// Check if path is within excluded directory (prefix match)
		pathWithSep := comparePath + string(filepath.Separator)
		excludedWithSep := compareExcluded + string(filepath.Separator)
		if strings.HasPrefix(pathWithSep, excludedWithSep) {
			return true
		}
	}
	
	return false
}

