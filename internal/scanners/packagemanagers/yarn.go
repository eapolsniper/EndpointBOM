package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// YarnScanner scans for globally installed yarn packages
type YarnScanner struct{}

func (s *YarnScanner) Name() string {
	return "yarn"
}

func (s *YarnScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("yarn") {
		return nil, nil
	}

	if !isCommandAvailable("yarn") {
		if cfg.Debug {
			fmt.Println("yarn not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get global packages
	cmd := exec.Command("yarn", "global", "list", "--json")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("yarn scan error: %v\n", err)
		}
		return nil, nil
	}

	// Yarn outputs multiple JSON objects per line
	lines := string(output)
	for _, line := range splitLines(lines) {
		var result yarnListResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		if result.Type == "tree" && result.Data.Trees != nil {
			for _, tree := range result.Data.Trees {
				comp := parseYarnPackage(tree)
				if comp.Name != "" {
					comp.PackageManager = "yarn"
					components = append(components, comp)
				}
			}
		}
	}

	return components, nil
}

type yarnListResult struct {
	Type string `json:"type"`
	Data struct {
		Trees []yarnTree `json:"trees"`
	} `json:"data"`
}

type yarnTree struct {
	Name     string     `json:"name"`
	Children []yarnTree `json:"children"`
}

func parseYarnPackage(tree yarnTree) scanners.Component {
	// Package format is usually "name@version"
	parts := splitAtLastChar(tree.Name, '@')
	name := tree.Name
	version := ""
	
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	}

	comp := scanners.Component{
		Type:       "library",
		Name:       name,
		Version:    version,
		Properties: make(map[string]string),
	}

	if tree.Children != nil {
		for _, child := range tree.Children {
			childComp := parseYarnPackage(child)
			if childComp.Name != "" {
				comp.Dependencies = append(comp.Dependencies, childComp)
			}
		}
	}

	return comp
}

