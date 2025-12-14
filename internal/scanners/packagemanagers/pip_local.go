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

// PipLocalScanner scans for locally installed pip packages in virtual environments
type PipLocalScanner struct{}

func (s *PipLocalScanner) Name() string {
	return "pip-local"
}

func (s *PipLocalScanner) Scan(cfg *config.Config) ([]scanners.Component, error) {
	if cfg.IsScannerDisabled("pip-local") {
		return nil, nil
	}

	var components []scanners.Component

	// Get list of potential project directories
	projectDirs := getProjectDirectories(cfg)

	for _, baseDir := range projectDirs {
		if cfg.Verbose {
			fmt.Printf("Scanning for Python virtual environments in: %s\n", baseDir)
		}

		// Find all virtual environment directories
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

			// Look for virtual environment directories
			if info.IsDir() && isVirtualEnv(path) {
				if cfg.Debug {
					fmt.Printf("Found Python virtual environment at: %s\n", path)
				}

				packages := scanVirtualEnv(path, cfg)
				components = append(components, packages...)

				// Don't descend into venv
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

// isVirtualEnv checks if a directory is a Python virtual environment
func isVirtualEnv(path string) bool {
	venvNames := []string{"venv", ".venv", "env", ".env", "virtualenv"}
	dirName := filepath.Base(path)

	for _, name := range venvNames {
		if dirName == name {
			// Verify it's actually a venv by checking for common structure
			binDir := filepath.Join(path, "bin")
			scriptsDir := filepath.Join(path, "Scripts") // Windows

			if info, err := os.Stat(binDir); err == nil && info.IsDir() {
				// Check for python executable
				pythonPath := filepath.Join(binDir, "python")
				if _, err := os.Stat(pythonPath); err == nil {
					return true
				}
			}

			if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
				// Windows venv
				pythonPath := filepath.Join(scriptsDir, "python.exe")
				if _, err := os.Stat(pythonPath); err == nil {
					return true
				}
			}
		}
	}

	return false
}

func scanVirtualEnv(venvPath string, cfg *config.Config) []scanners.Component {
	var components []scanners.Component
	projectPath := filepath.Dir(venvPath)

	// Determine the python executable path
	pythonExec := filepath.Join(venvPath, "bin", "python")
	if _, err := os.Stat(pythonExec); err != nil {
		// Try Windows path
		pythonExec = filepath.Join(venvPath, "Scripts", "python.exe")
		if _, err := os.Stat(pythonExec); err != nil {
			if cfg.Debug {
				fmt.Printf("Could not find python executable in %s\n", venvPath)
			}
			return nil
		}
	}

	// Run pip freeze to get installed packages
	cmd := exec.Command(pythonExec, "-m", "pip", "freeze")
	output, err := cmd.Output()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("pip freeze failed for %s: %v\n", venvPath, err)
		}
		return nil
	}

	// Build a map of all packages first
	packageMap := make(map[string]*scanners.Component)
	var packageNames []string

	// Parse pip freeze output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse package==version or package @ location
		parts := strings.Split(line, "==")
		if len(parts) != 2 {
			// Try @ syntax (editable installs)
			parts = strings.Split(line, " @ ")
			if len(parts) != 2 {
				continue
			}
		}

		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		packageNames = append(packageNames, name)

		comp := scanners.Component{
			Type:           "library",
			Name:           name,
			Version:        version,
			PackageManager: "pip",
			Location:       projectPath,
			Properties:     make(map[string]string),
			Dependencies:   []scanners.Component{},
		}

		comp.Properties["install_type"] = "local"
		comp.Properties["project_path"] = projectPath
		comp.Properties["venv_path"] = venvPath
		comp.Properties["source"] = "pip-local"

		packageMap[strings.ToLower(name)] = &comp
	}

	// Now get dependency information for each package
	for _, pkgName := range packageNames {
		deps := getLocalPipDependencies(pythonExec, pkgName, cfg)
		if comp, exists := packageMap[strings.ToLower(pkgName)]; exists {
			for _, depName := range deps {
				if depComp, depExists := packageMap[strings.ToLower(depName)]; depExists {
					comp.Dependencies = append(comp.Dependencies, *depComp)
				}
			}
		}
	}

	// Convert map to slice
	for _, comp := range packageMap {
		components = append(components, *comp)
	}

	return components
}

// getLocalPipDependencies gets the direct dependencies of a package in a venv
func getLocalPipDependencies(pythonExec, packageName string, cfg *config.Config) []string {
	var deps []string

	cmd := exec.Command(pythonExec, "-m", "pip", "show", packageName)
	output, err := cmd.Output()
	if err != nil {
		return deps
	}

	// Parse pip show output for "Requires:" line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Requires:") {
			// Format: "Requires: package1, package2, package3"
			requiresStr := strings.TrimPrefix(line, "Requires:")
			requiresStr = strings.TrimSpace(requiresStr)
			
			if requiresStr != "" && requiresStr != "None" {
				// Split by comma and clean up
				for _, dep := range strings.Split(requiresStr, ",") {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						deps = append(deps, dep)
					}
				}
			}
			break
		}
	}

	return deps
}

