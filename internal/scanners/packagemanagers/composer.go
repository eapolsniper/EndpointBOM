package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// ComposerScanner scans for PHP Composer global packages
type ComposerScanner struct{}

func (s *ComposerScanner) Name() string {
	return "composer"
}

func (s *ComposerScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("composer") {
		return nil, nil
	}

	if !isCommandAvailable("composer") {
		if cfg.Debug {
			fmt.Println("composer not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	cmd := exec.Command("composer", "global", "show", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("composer scan error: %v\n", err)
		}
		return nil, nil
	}

	var result composerShowResult
	if err := json.Unmarshal(output, &result); err != nil {
		if cfg.Debug {
			fmt.Printf("composer parse error: %v\n", err)
		}
		return nil, nil
	}

	for _, pkg := range result.Installed {
		comp := scanners.Component{
			Type:           "library",
			Name:           pkg.Name,
			Version:        pkg.Version,
			PackageManager: "composer",
			Description:    pkg.Description,
			Properties:     make(map[string]string),
		}

		if len(pkg.Keywords) > 0 {
			comp.Properties["keywords"] = strings.Join(pkg.Keywords, ", ")
		}

		components = append(components, comp)
	}

	return components, nil
}

type composerShowResult struct {
	Installed []composerPackage `json:"installed"`
}

type composerPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

