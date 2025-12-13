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

// GoScanner scans for Go installed tools
type GoScanner struct{}

func (s *GoScanner) Name() string {
	return "go"
}

func (s *GoScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("go") {
		return nil, nil
	}

	if !isCommandAvailable("go") {
		if cfg.Debug {
			fmt.Println("go not found, skipping")
		}
		return nil, nil
	}

	var components []scanners.Component

	// List installed Go tools (go list will show installed binaries)
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("go scan error: %v\n", err)
		}
		return nil, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		version := parts[1]

		comp := scanners.Component{
			Type:           "library",
			Name:           name,
			Version:        version,
			PackageManager: "go",
			Properties:     make(map[string]string),
		}

		components = append(components, comp)
	}

	return components, nil
}

