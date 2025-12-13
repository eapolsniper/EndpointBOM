package packagemanagers

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// ChocolateyScanner scans for Chocolatey installed packages (Windows)
type ChocolateyScanner struct{}

func (s *ChocolateyScanner) Name() string {
	return "chocolatey"
}

func (s *ChocolateyScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("chocolatey") {
		return nil, nil
	}

	if runtime.GOOS != "windows" {
		return nil, nil
	}

	if !isCommandAvailable("choco") {
		if cfg.Debug {
			fmt.Println("choco not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	cmd := exec.Command("choco", "list", "--local-only")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("chocolatey scan error: %v\n", err)
		}
		return nil, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip header and footer lines
		if strings.Contains(line, "packages installed") {
			continue
		}

		// Format: "package-name version"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			comp := scanners.Component{
				Type:           "application",
				Name:           parts[0],
				Version:        parts[1],
				PackageManager: "chocolatey",
				Properties:     make(map[string]string),
			}
			components = append(components, comp)
		}
	}

	return components, nil
}

