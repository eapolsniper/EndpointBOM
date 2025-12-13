package packagemanagers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// PipScanner scans for pip installed packages
type PipScanner struct{}

func (s *PipScanner) Name() string {
	return "pip"
}

func (s *PipScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("pip") {
		return nil, nil
	}

	var components []scanners.Component

	// Try pip, pip3, and python -m pip
	commands := [][]string{
		{"pip", "list", "--format=json"},
		{"pip3", "list", "--format=json"},
		{"python", "-m", "pip", "list", "--format=json"},
		{"python3", "-m", "pip", "list", "--format=json"},
	}

	seen := make(map[string]bool)

	for _, cmdArgs := range commands {
		if !isCommandAvailable(cmdArgs[0]) {
			continue
		}

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.Output()
		if err != nil {
			if cfg.Debug {
				fmt.Printf("pip scan error with %v: %v\n", cmdArgs, err)
			}
			continue
		}

		var packages []pipPackage
		if err := json.Unmarshal(output, &packages); err != nil {
			if cfg.Debug {
				fmt.Printf("pip parse error: %v\n", err)
			}
			continue
		}

		for _, pkg := range packages {
			key := pkg.Name + "@" + pkg.Version
			if seen[key] {
				continue
			}
			seen[key] = true

			comp := scanners.Component{
				Type:           "library",
				Name:           pkg.Name,
				Version:        pkg.Version,
				PackageManager: "pip",
				Properties:     make(map[string]string),
			}

			// Try to get dependencies using pip show
			deps := getPipDependencies(cmdArgs[0], pkg.Name, cfg)
			if len(deps) > 0 {
				comp.Properties["requires"] = strings.Join(deps, ", ")
			}

			components = append(components, comp)
		}
	}

	return components, nil
}

type pipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func getPipDependencies(pipCmd, packageName string, cfg *config.Config) []string {
	cmd := exec.Command(pipCmd, "show", packageName)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Requires:") {
			depsStr := strings.TrimPrefix(line, "Requires:")
			depsStr = strings.TrimSpace(depsStr)
			if depsStr == "" {
				return nil
			}
			deps := strings.Split(depsStr, ",")
			for i, dep := range deps {
				deps[i] = strings.TrimSpace(dep)
			}
			return deps
		}
	}

	return nil
}

