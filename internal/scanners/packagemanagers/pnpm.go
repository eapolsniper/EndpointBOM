package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// PnpmScanner scans for globally installed pnpm packages
type PnpmScanner struct{}

func (s *PnpmScanner) Name() string {
	return "pnpm"
}

func (s *PnpmScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("pnpm") {
		return nil, nil
	}

	if !isCommandAvailable("pnpm") {
		if cfg.Debug {
			fmt.Println("pnpm not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get global packages
	cmd := exec.Command("pnpm", "list", "-g", "--json", "--depth=Infinity")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("pnpm scan error: %v\n", err)
		}
		return nil, nil
	}

	var result []pnpmListResult
	if err := json.Unmarshal(output, &result); err != nil {
		if cfg.Debug {
			fmt.Printf("pnpm parse error: %v\n", err)
		}
		return nil, nil
	}

	for _, pkg := range result {
		if pkg.Dependencies != nil {
			components = parsePnpmDependencies(pkg.Dependencies)
		}
	}

	return components, nil
}

type pnpmListResult struct {
	Dependencies map[string]pnpmPackage `json:"dependencies"`
}

type pnpmPackage struct {
	Version      string                `json:"version"`
	Path         string                `json:"path"`
	Dependencies map[string]pnpmPackage `json:"dependencies"`
}

func parsePnpmDependencies(deps map[string]pnpmPackage) []scanners.Component {
	var components []scanners.Component

	for name, pkg := range deps {
		comp := scanners.Component{
			Type:           "library",
			Name:           name,
			Version:        pkg.Version,
			PackageManager: "pnpm",
			Location:       pkg.Path,
			Properties:     make(map[string]string),
		}

		// Parse transitive dependencies
		if pkg.Dependencies != nil {
			comp.Dependencies = parsePnpmDependencies(pkg.Dependencies)
		}

		components = append(components, comp)
	}

	return components
}

