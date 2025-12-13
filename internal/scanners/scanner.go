package scanners

import "github.com/eapolsniper/endpointbom/internal/config"

// Component represents a discovered software component
type Component struct {
	Type            string            // application, library, etc.
	Name            string            // Component name
	Version         string            // Version string
	Group           string            // Group/namespace (optional)
	Description     string            // Description (optional)
	PackageManager  string            // Source package manager (npm, pip, etc.)
	Location        string            // Installation location
	Dependencies    []Component       // Transitive dependencies
	Properties      map[string]string // Additional properties
}

// Scanner is the interface that all scanners must implement
type Scanner interface {
	// Name returns the scanner name
	Name() string

	// Scan performs the scan and returns discovered components
	Scan(cfg *config.Config) ([]Component, error)
}

// ScanResult contains the results of all scans
type ScanResult struct {
	Applications      []Component
	PackageManagers   []Component
	IDEExtensions     []Component
	BrowserExtensions []Component
}

