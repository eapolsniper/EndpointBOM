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

// GemScanner scans for Ruby gems
type GemScanner struct{}

func (s *GemScanner) Name() string {
	return "gem"
}

func (s *GemScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("gem") {
		return nil, nil
	}

	if !isCommandAvailable("gem") {
		if cfg.Debug {
			fmt.Println("gem not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	cmd := exec.Command("gem", "list", "--local")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("gem scan error: %v\n", err)
		}
		return nil, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: "gem-name (version1, version2)"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		versionsStr := strings.Trim(parts[1], "()")
		versions := strings.Split(versionsStr, ",")

		for _, ver := range versions {
			comp := scanners.Component{
				Type:           "library",
				Name:           name,
				Version:        strings.TrimSpace(ver),
				PackageManager: "gem",
				Properties:     make(map[string]string),
			}
			components = append(components, comp)
		}
	}

	return components, nil
}

