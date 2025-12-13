package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// BrewScanner scans for Homebrew installed packages (macOS/Linux)
type BrewScanner struct{}

func (s *BrewScanner) Name() string {
	return "brew"
}

func (s *BrewScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("brew") {
		return nil, nil
	}

	// Homebrew is primarily for macOS and Linux
	if runtime.GOOS == "windows" {
		return nil, nil
	}

	if !isCommandAvailable("brew") {
		if cfg.Debug {
			fmt.Println("brew not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get installed formulae
	cmd := exec.Command("brew", "list", "--formula", "--json")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("brew scan error: %v\n", err)
		}
		return nil, nil
	}

	var formulae []brewFormula
	if err := json.Unmarshal(output, &formulae); err != nil {
		if cfg.Debug {
			fmt.Printf("brew parse error: %v\n", err)
		}
		return nil, nil
	}

	for _, formula := range formulae {
		comp := scanners.Component{
			Type:           "application",
			Name:           formula.Name,
			Version:        formula.Version,
			PackageManager: "brew",
			Description:    formula.Desc,
			Location:       formula.Prefix,
			Properties:     make(map[string]string),
		}

		if formula.Homepage != "" {
			comp.Properties["homepage"] = formula.Homepage
		}

		components = append(components, comp)
	}

	// Get installed casks
	cmd = exec.Command("brew", "list", "--cask", "--json")
	output, err = cmd.Output()
	if err == nil {
		var casks []brewCask
		if err := json.Unmarshal(output, &casks); err == nil {
			for _, cask := range casks {
				comp := scanners.Component{
					Type:           "application",
					Name:           cask.Token,
					Version:        cask.Version,
					PackageManager: "brew-cask",
					Location:       cask.CaskroomPath,
					Properties:     make(map[string]string),
				}

				components = append(components, comp)
			}
		}
	}

	return components, nil
}

type brewFormula struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Desc     string `json:"desc"`
	Homepage string `json:"homepage"`
	Prefix   string `json:"prefix"`
}

type brewCask struct {
	Token        string `json:"token"`
	Version      string `json:"version"`
	CaskroomPath string `json:"caskroom_path"`
}

