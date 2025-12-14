package packagemanagers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/scanners"
)

// NPMLocalScanner scans for locally installed npm packages in project directories
type NPMLocalScanner struct{}

func (s *NPMLocalScanner) Name() string {
	return "npm-local"
}

func (s *NPMLocalScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("npm-local") {
		return nil, nil
	}

	// Check if npm is installed
	if !isCommandAvailable("npm") {
		if cfg.Debug {
			fmt.Println("npm not found, skipping local npm scan")
		}
		return nil, nil
	}

	var components []scanners.Component

	// Get list of potential project directories
	projectDirs := getProjectDirectories(cfg)

	for _, baseDir := range projectDirs {
		if cfg.Verbose {
			fmt.Printf("Scanning for npm projects in: %s\n", baseDir)
		}

		// Find all node_modules directories
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

			// Skip node_modules within node_modules (nested)
			if info.IsDir() && info.Name() == "node_modules" {
				// Check if parent has package.json
				projectPath := filepath.Dir(path)
				packageJSONPath := filepath.Join(projectPath, "package.json")
				
				if _, err := os.Stat(packageJSONPath); err == nil {
					// Found a valid node project
					if cfg.Debug {
						fmt.Printf("Found npm project at: %s\n", projectPath)
					}
					
					packages := scanNPMProject(projectPath, cfg)
					components = append(components, packages...)
				}

				// Don't descend into node_modules
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

func scanNPMProject(projectPath string, cfg *config.Config) []scanners.Component {
	var components []scanners.Component

	// Run npm list with full dependency tree
	cmd := exec.Command("npm", "list", "--json", "--all")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		// npm list returns non-zero even on success sometimes
		if len(output) == 0 {
			if cfg.Debug {
				fmt.Printf("npm list failed for %s: %v\n", projectPath, err)
			}
			return nil
		}
	}

	var result npmListResult
	if err := json.Unmarshal(output, &result); err != nil {
		if cfg.Debug {
			fmt.Printf("npm parse error for %s: %v\n", projectPath, err)
		}
		return nil
	}

	// Parse dependencies with full tree
	if result.Dependencies != nil {
		for name, pkg := range result.Dependencies {
			comp := scanners.Component{
				Type:           "library",
				Name:           name,
				Version:        pkg.Version,
				PackageManager: "npm",
				Location:       projectPath,
				Properties:     make(map[string]string),
			}

			comp.Properties["install_type"] = "local"
			comp.Properties["project_path"] = projectPath
			comp.Properties["source"] = "npm-local"
			comp.Properties["dependency_depth"] = "0" // Direct dependency

			if pkg.Resolved != "" {
				comp.Properties["resolved"] = pkg.Resolved
			}

			// Parse transitive dependencies recursively
			if pkg.Dependencies != nil {
				comp.Dependencies = parseLocalNPMDependencies(pkg.Dependencies, projectPath, 1)
			}

			components = append(components, comp)
		}
	}

	return components
}

// parseLocalNPMDependencies recursively parses npm dependency trees
func parseLocalNPMDependencies(deps map[string]npmPackage, projectPath string, depth int) []scanners.Component {
	var components []scanners.Component

	for name, pkg := range deps {
		comp := scanners.Component{
			Type:           "library",
			Name:           name,
			Version:        pkg.Version,
			PackageManager: "npm",
			Location:       projectPath,
			Properties:     make(map[string]string),
		}

		comp.Properties["dependency_depth"] = fmt.Sprintf("%d", depth)
		
		if pkg.Resolved != "" {
			comp.Properties["resolved"] = pkg.Resolved
		}

		// Recursively parse nested dependencies
		if pkg.Dependencies != nil {
			comp.Dependencies = parseLocalNPMDependencies(pkg.Dependencies, projectPath, depth+1)
		}

		components = append(components, comp)
	}

	return components
}

// getProjectDirectories returns common directories where projects might be located
func getProjectDirectories(cfg *config.Config) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	// Common project directory names
	commonDirs := []string{
		"projects",
		"code",
		"dev",
		"development",
		"workspace",
		"repos",
		"git",
		"src",
		"work",
	}

	var projectDirs []string

	// Check common project locations
	for _, dir := range commonDirs {
		fullPath := filepath.Join(homeDir, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			projectDirs = append(projectDirs, fullPath)
		}
	}

	// Also check if home directory itself has projects
	// but limit depth to avoid scanning everything
	entries, err := os.ReadDir(homeDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			
			// Skip hidden directories, system directories
			if strings.HasPrefix(name, ".") || 
			   name == "Library" || 
			   name == "Documents" || 
			   name == "Downloads" || 
			   name == "Desktop" ||
			   name == "Pictures" ||
			   name == "Movies" ||
			   name == "Music" {
				continue
			}

			fullPath := filepath.Join(homeDir, name)
			
			// Check if this looks like it might contain projects
			// (has node_modules, package.json, .git, etc.)
			if looksLikeProjectDir(fullPath) {
				projectDirs = append(projectDirs, fullPath)
			}
		}
	}

	return projectDirs
}

func looksLikeProjectDir(path string) bool {
	// Check for common project indicators
	indicators := []string{
		"package.json",
		"go.mod",
		"requirements.txt",
		"Gemfile",
		"Cargo.toml",
		"composer.json",
		".git",
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		for _, indicator := range indicators {
			if entry.Name() == indicator {
				return true
			}
		}
	}

	return false
}

