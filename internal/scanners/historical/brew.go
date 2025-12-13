package historical

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// BrewHistoricalScanner uses Homebrew's built-in install date tracking
type BrewHistoricalScanner struct{}

// Name returns the scanner name
func (s *BrewHistoricalScanner) Name() string {
	return "brew-historical"
}

// Scan performs the historical scan using `brew info --json=v2 --installed`
// Homebrew already tracks install dates - easy win!
func (s *BrewHistoricalScanner) Scan(cfg *config.Config) ([]Component, error) {
	if !cfg.IncludeHistorical {
		return []Component{}, nil
	}

	// Check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return []Component{}, nil
	}

	lookbackDuration := time.Duration(cfg.HistoricalLookbackDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-lookbackDuration)

	// Get all installed packages with metadata including install dates
	cmd := exec.Command("brew", "info", "--json=v2", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return []Component{}, nil
	}

	var brewInfo struct {
		Formulae []struct {
			Name              string `json:"name"`
			FullName          string `json:"full_name"`
			Versions          map[string]string `json:"versions"`
			InstalledVersions []struct {
				Version       string `json:"version"`
				InstalledTime string `json:"installed_time"`
			} `json:"installed"`
		} `json:"formulae"`
		Casks []struct {
			Token         string `json:"token"`
			Version       string `json:"version"`
			InstalledTime string `json:"installed_time,omitempty"`
		} `json:"casks"`
	}

	if err := json.Unmarshal(output, &brewInfo); err != nil {
		return []Component{}, nil
	}

	var components []Component

	// Process formulae (command-line tools)
	for _, formula := range brewInfo.Formulae {
		for _, installed := range formula.InstalledVersions {
			// Parse install time
			installTime, err := time.Parse(time.RFC3339, installed.InstalledTime)
			if err != nil {
				continue
			}

			// Only include if within lookback period
			if !installTime.After(cutoffTime) {
				continue
			}

			components = append(components, scanners.Component{
				Type:           "library",
				Name:           formula.Name,
				Version:        installed.Version,
				PackageManager: "brew",
				Properties: map[string]string{
					"install_date": installTime.Format(time.RFC3339),
					"install_type": "historical", // Will be overridden if still installed
					"source":       "brew_info",
					"package_type": "formula",
				},
			})
		}
	}

	// Process casks (GUI applications)
	for _, cask := range brewInfo.Casks {
		if cask.InstalledTime == "" {
			continue
		}

		installTime, err := time.Parse(time.RFC3339, cask.InstalledTime)
		if err != nil {
			continue
		}

		if !installTime.After(cutoffTime) {
			continue
		}

		components = append(components, scanners.Component{
			Type:           "application",
			Name:           cask.Token,
			Version:        cask.Version,
			PackageManager: "brew",
			Properties: map[string]string{
				"install_date": installTime.Format(time.RFC3339),
				"install_type": "historical", // Will be overridden if still installed
				"source":       "brew_info",
				"package_type": "cask",
			},
		})
	}

	return components, nil
}

// GetLogFiles returns no logs (brew info is not a log file)
func (s *BrewHistoricalScanner) GetLogFiles(cfg *config.Config) ([]string, error) {
	return []string{}, nil
}

