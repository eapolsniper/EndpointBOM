package packagemanagers

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// CargoScanner scans for Rust cargo installed packages
type CargoScanner struct{}

func (s *CargoScanner) Name() string {
	return "cargo"
}

func (s *CargoScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("cargo") {
		return nil, nil
	}

	if !isCommandAvailable("cargo") {
		if cfg.Debug {
			fmt.Println("cargo not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	cmd := exec.Command("cargo", "install", "--list")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("cargo scan error: %v\n", err)
		}
		return nil, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var currentPkg string
	var currentVersion string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Lines starting without whitespace are package names
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// Format: "package-name v0.1.0:"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentPkg = parts[0]
				currentVersion = strings.TrimPrefix(strings.TrimSuffix(parts[1], ":"), "v")
				
				comp := scanners.Component{
					Type:           "library",
					Name:           currentPkg,
					Version:        currentVersion,
					PackageManager: "cargo",
					Properties:     make(map[string]string),
				}
				components = append(components, comp)
			}
		}
	}

	return components, nil
}

