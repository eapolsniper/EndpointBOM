package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// NPMScanner scans for globally installed npm packages
type NPMScanner struct{}

func (s *NPMScanner) Name() string {
	return "npm"
}

func (s *NPMScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("npm") {
		return nil, nil
	}

	// Check if npm is installed
	if !isCommandAvailable("npm") {
		if cfg.Debug {
			fmt.Println("npm not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get global packages with all dependencies
	cmd := exec.Command("npm", "ls", "-g", "--all", "--json", "--depth=999")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("npm scan error: %v\n", err)
		}
		// npm ls returns non-zero even on success sometimes
		if len(output) == 0 {
			return nil, nil
		}
	}

	var result npmListResult
	if err := json.Unmarshal(output, &result); err != nil {
		if cfg.Debug {
			fmt.Printf("npm parse error: %v\n", err)
		}
		return nil, nil
	}

	// Parse dependencies recursively
	if result.Dependencies != nil {
		components = parseNPMDependencies(result.Dependencies, "")
	}

	return components, nil
}

type npmListResult struct {
	Dependencies map[string]npmPackage `json:"dependencies"`
}

type npmPackage struct {
	Version      string                `json:"version"`
	Resolved     string                `json:"resolved"`
	Dependencies map[string]npmPackage `json:"dependencies"`
}

func parseNPMDependencies(deps map[string]npmPackage, location string) []scanners.Component {
	var components []scanners.Component

	for name, pkg := range deps {
		comp := scanners.Component{
			Type:           "library",
			Name:           name,
			Version:        pkg.Version,
			PackageManager: "npm",
			Location:       location,
			Properties:     make(map[string]string),
		}

		if pkg.Resolved != "" {
			comp.Properties["resolved"] = pkg.Resolved
		}

		// Parse transitive dependencies
		if pkg.Dependencies != nil {
			comp.Dependencies = parseNPMDependencies(pkg.Dependencies, filepath.Join(location, name))
		}

		components = append(components, comp)
	}

	return components
}

