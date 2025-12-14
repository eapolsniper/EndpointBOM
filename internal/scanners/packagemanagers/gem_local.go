package packagemanagers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// GemLocalScanner scans for locally installed Ruby gems in Bundler projects
type GemLocalScanner struct{}

func (s *GemLocalScanner) Name() string {
	return "gem-local"
}

func (s *GemLocalScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("gem-local") {
		return nil, nil
	}

	// Check if bundle is installed
	if !isCommandAvailable("bundle") {
		if cfg.Debug {
			fmt.Println("bundle not found, skipping local gem scan")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get list of potential project directories
	projectDirs := getProjectDirectories(cfg)

	for _, baseDir := range projectDirs {
		if cfg.Verbose {
			fmt.Printf("Scanning for Ruby/Bundler projects in: %s\n", baseDir)
		}

		// Find all Gemfile directories
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip directories we can't access
			}

			// Skip if this path is excluded
			if cfg.IsPathExcluded(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Look for Gemfile
			if !info.IsDir() && info.Name() == "Gemfile" {
				projectPath := filepath.Dir(path)
				if cfg.Debug {
					fmt.Printf("Found Ruby project at: %s\n", projectPath)
				}

				packages := scanBundlerProject(projectPath, cfg)
				components = append(components, packages...)
			}

			// Skip vendor directories
			if info.IsDir() && info.Name() == "vendor" {
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil && cfg.Debug {
			fmt.Printf("Error walking directory %s: %v\n", baseDir, err)
		}
	}

	return components, nil
}

func scanBundlerProject(projectPath string, cfg *config.Config) []scanners.Component {
	var components []scanners.Component

	// Check if Gemfile.lock exists
	lockfilePath := filepath.Join(projectPath, "Gemfile.lock")
	if _, err := os.Stat(lockfilePath); err != nil {
		// No lockfile, gems might not be installed
		if cfg.Debug {
			fmt.Printf("No Gemfile.lock found in %s, skipping\n", projectPath)
		}
		return nil
	}

	// Build a map of all gems first
	gemMap := make(map[string]*scanners.Component)

	// Run bundle list
	cmd := exec.Command("bundle", "list")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("bundle list failed for %s: %v\n", projectPath, err)
		}
		return nil
	}

	// Parse bundle list output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Gems included") {
			continue
		}

		// Format: "  * gem_name (version)"
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		// Extract name and version
		if idx := strings.Index(line, " ("); idx > 0 {
			name := line[:idx]
			versionPart := line[idx+2:]
			version := strings.TrimSuffix(versionPart, ")")

			comp := scanners.Component{
				Type:           "library",
				Name:           name,
				Version:        version,
				PackageManager: "gem",
				Location:       projectPath,
				Properties:     make(map[string]string),
				Dependencies:   []scanners.Component{},
			}

			comp.Properties["install_type"] = "local"
			comp.Properties["project_path"] = projectPath
			comp.Properties["source"] = "gem-local"

			gemMap[name] = &comp
		}
	}

	// Parse Gemfile.lock for dependency information
	lockfileData, err := os.ReadFile(lockfilePath)
	if err == nil {
		parseGemfileLockDependencies(string(lockfileData), gemMap)
	}

	// Convert map to slice
	for _, comp := range gemMap {
		components = append(components, *comp)
	}

	return components
}

// parseGemfileLockDependencies parses Gemfile.lock to extract dependency relationships
func parseGemfileLockDependencies(lockfileContent string, gemMap map[string]*scanners.Component) {
	scanner := bufio.NewScanner(strings.NewReader(lockfileContent))
	
	var currentGem string
	inSpecsSection := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Look for the specs section
		if strings.Contains(line, "specs:") {
			inSpecsSection = true
			continue
		}
		
		if !inSpecsSection {
			continue
		}
		
		// End of specs section
		if line != "" && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "  ") {
			break
		}
		
		// Gem declaration: "    gem_name (version)"
		if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 1 {
				currentGem = parts[0]
			}
			continue
		}
		
		// Dependency: "      dep_gem (>= version)"
		if strings.HasPrefix(line, "      ") && currentGem != "" {
			depLine := strings.TrimSpace(line)
			parts := strings.Fields(depLine)
			if len(parts) >= 1 {
				depName := parts[0]
				
				// Add this dependency to the current gem
				if gemComp, exists := gemMap[currentGem]; exists {
					if depComp, depExists := gemMap[depName]; depExists {
						gemComp.Dependencies = append(gemComp.Dependencies, *depComp)
					}
				}
			}
		}
	}
}

