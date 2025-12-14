package sbom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/eapolsniper/endpointbom/internal/scanners"
	"github.com/eapolsniper/endpointbom/internal/system"
)

// GenerateSBOMs creates CycloneDX SBOM files for different component categories
func GenerateSBOMs(result *scanners.ScanResult, sysInfo *system.Info, outputDir string) error {
	timestamp := time.Now().Format("20060102-150405")
	hostname := sysInfo.Hostname

	// Generate SBOM for package managers
	if len(result.PackageManagers) > 0 {
		filename := fmt.Sprintf("%s.%s.package-managers.cdx.json", hostname, timestamp)
		if err := generateSBOM(result.PackageManagers, sysInfo, filepath.Join(outputDir, filename), "package-managers"); err != nil {
			return fmt.Errorf("failed to generate package managers SBOM: %w", err)
		}
		fmt.Printf("Generated: %s\n", filename)
	}

	// Generate SBOM for applications
	if len(result.Applications) > 0 {
		filename := fmt.Sprintf("%s.%s.applications.cdx.json", hostname, timestamp)
		if err := generateSBOM(result.Applications, sysInfo, filepath.Join(outputDir, filename), "applications"); err != nil {
			return fmt.Errorf("failed to generate applications SBOM: %w", err)
		}
		fmt.Printf("Generated: %s\n", filename)
	}

	// Generate SBOM for IDE extensions
	if len(result.IDEExtensions) > 0 {
		filename := fmt.Sprintf("%s.%s.ide-extensions.cdx.json", hostname, timestamp)
		if err := generateSBOM(result.IDEExtensions, sysInfo, filepath.Join(outputDir, filename), "ide-extensions"); err != nil {
			return fmt.Errorf("failed to generate IDE extensions SBOM: %w", err)
		}
		fmt.Printf("Generated: %s\n", filename)
	}

	// Generate SBOM for browser extensions
	if len(result.BrowserExtensions) > 0 {
		filename := fmt.Sprintf("%s.%s.browser-extensions.cdx.json", hostname, timestamp)
		if err := generateSBOM(result.BrowserExtensions, sysInfo, filepath.Join(outputDir, filename), "browser-extensions"); err != nil {
			return fmt.Errorf("failed to generate browser extensions SBOM: %w", err)
		}
		fmt.Printf("Generated: %s\n", filename)
	}

	return nil
}

func generateSBOM(components []scanners.Component, sysInfo *system.Info, outputPath string, category string) error {
	// Create BOM
	bom := cdx.NewBOM()
	bom.SerialNumber = "urn:uuid:" + generateUUID()
	bom.Version = 1
	
	// Create root component with bom-ref
	rootBomRef := fmt.Sprintf("device:%s", sysInfo.Hostname)
	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Component: &cdx.Component{
			BOMRef:  rootBomRef,
			Type:    cdx.ComponentTypeDevice,
			Name:    sysInfo.Hostname,
			Version: sysInfo.OSVersion,
			Properties: &[]cdx.Property{
				{
					Name:  "os",
					Value: sysInfo.OSName,
				},
				{
					Name:  "os_version",
					Value: sysInfo.OSVersion,
				},
				{
					Name:  "scan_category",
					Value: category,
				},
			},
		},
	}

	// Add logged-in users to metadata
	if len(sysInfo.Users) > 0 {
		for _, user := range sysInfo.Users {
			*bom.Metadata.Component.Properties = append(*bom.Metadata.Component.Properties, cdx.Property{
				Name:  "logged_in_user",
				Value: user,
			})
		}
	}

	// Add network information to metadata
	if len(sysInfo.LocalIPs) > 0 {
		for _, ip := range sysInfo.LocalIPs {
			*bom.Metadata.Component.Properties = append(*bom.Metadata.Component.Properties, cdx.Property{
				Name:  "local_ip",
				Value: ip,
			})
		}
	}

	if sysInfo.PublicIP != "" && sysInfo.PublicIP != "unavailable" {
		*bom.Metadata.Component.Properties = append(*bom.Metadata.Component.Properties, cdx.Property{
			Name:  "public_ip",
			Value: sysInfo.PublicIP,
		})
	}

	// Convert components to CycloneDX components and build dependency graph
	var dependencies []cdx.Dependency
	rootDependsOnMap := make(map[string]bool) // Use map to deduplicate root dependencies
	
	// Use a map to deduplicate components by bom-ref
	componentMap := make(map[string]cdx.Component)
	
	for _, comp := range components {
		cdxComp, deps := convertToCycloneDXComponentWithDeps(comp, componentMap)
		
		// Add to map (will deduplicate automatically)
		componentMap[cdxComp.BOMRef] = cdxComp
		
		// Root device depends on all top-level components (deduplicated via map)
		rootDependsOnMap[cdxComp.BOMRef] = true
		
		// Add this component's dependencies to the dependencies array
		dependencies = append(dependencies, deps...)
	}
	
	// Convert map to slice for BOM
	var cdxComponents []cdx.Component
	for _, comp := range componentMap {
		cdxComponents = append(cdxComponents, comp)
	}

	if len(cdxComponents) > 0 {
		bom.Components = &cdxComponents
	}
	
	// Deduplicate dependencies by ref
	dependencyMap := make(map[string]cdx.Dependency)
	for _, dep := range dependencies {
		if existing, exists := dependencyMap[dep.Ref]; exists {
			// Merge dependsOn lists if both exist
			if dep.Dependencies != nil && existing.Dependencies != nil {
				// Deduplicate the dependsOn list
				mergedMap := make(map[string]bool)
				for _, ref := range *existing.Dependencies {
					mergedMap[ref] = true
				}
				for _, ref := range *dep.Dependencies {
					mergedMap[ref] = true
				}
				merged := make([]string, 0, len(mergedMap))
				for ref := range mergedMap {
					merged = append(merged, ref)
				}
				existing.Dependencies = &merged
				dependencyMap[dep.Ref] = existing
			}
		} else {
			dependencyMap[dep.Ref] = dep
		}
	}
	
	// Convert dependency map back to slice
	var deduplicatedDeps []cdx.Dependency
	for _, dep := range dependencyMap {
		deduplicatedDeps = append(deduplicatedDeps, dep)
	}
	
	// Add root dependency (device depends on all top-level components)
	if len(rootDependsOnMap) > 0 {
		rootDependsOn := make([]string, 0, len(rootDependsOnMap))
		for ref := range rootDependsOnMap {
			rootDependsOn = append(rootDependsOn, ref)
		}
		deduplicatedDeps = append([]cdx.Dependency{{
			Ref:          rootBomRef,
			Dependencies: &rootDependsOn,
		}}, deduplicatedDeps...)
	}
	
	if len(deduplicatedDeps) > 0 {
		bom.Dependencies = &deduplicatedDeps
	}

	// Write to file
	data, err := json.MarshalIndent(bom, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal BOM: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write BOM file: %w", err)
	}

	return nil
}

// convertToCycloneDXComponentWithDeps converts a component and returns both the component and its dependency relationships
func convertToCycloneDXComponentWithDeps(comp scanners.Component, componentMap map[string]cdx.Component) (cdx.Component, []cdx.Dependency) {
	var allDependencies []cdx.Dependency
	
	// Generate bom-ref using Package URL (purl) format when possible
	bomRef := generateBomRef(comp)
	
	// Check if we've already processed this component
	if existingComp, exists := componentMap[bomRef]; exists {
		// Component already exists, just return it (no need to process dependencies again)
		return existingComp, []cdx.Dependency{}
	}
	
	cdxComp := cdx.Component{
		BOMRef:  bomRef,
		Name:    comp.Name,
		Version: comp.Version,
	}

	// Map component type
	switch comp.Type {
	case "library":
		cdxComp.Type = cdx.ComponentTypeLibrary
	case "application":
		cdxComp.Type = cdx.ComponentTypeApplication
	case "ide-extension":
		cdxComp.Type = cdx.ComponentTypeLibrary // Extensions are library-like
	case "browser-extension":
		cdxComp.Type = cdx.ComponentTypeLibrary // Browser extensions are library-like
	case "mcp-server":
		cdxComp.Type = cdx.ComponentTypeApplication // MCP servers as applications
	default:
		cdxComp.Type = cdx.ComponentTypeLibrary
	}

	// Build enhanced description
	cdxComp.Description = buildDescription(comp)

	// Add group/namespace if available
	if comp.Group != "" {
		cdxComp.Group = comp.Group
	}

	// Add properties
	var props []cdx.Property
	
	// Add component type as a property for better filtering in Dependency-Track
	props = append(props, cdx.Property{
		Name:  "component_type",
		Value: comp.Type,
	})
	
	for key, value := range comp.Properties {
		props = append(props, cdx.Property{
			Name:  key,
			Value: value,
		})
	}
	
	// Add package manager info
	if comp.PackageManager != "" {
		props = append(props, cdx.Property{
			Name:  "package_manager",
			Value: comp.PackageManager,
		})
	}
	
	// Add location info
	if comp.Location != "" {
		props = append(props, cdx.Property{
			Name:  "location",
			Value: comp.Location,
		})
	}
	
	if len(props) > 0 {
		cdxComp.Properties = &props
	}

	// Add this component to the map early to prevent circular dependency issues
	componentMap[bomRef] = cdxComp
	
	// Process dependencies recursively
	var dependsOn []string
	if len(comp.Dependencies) > 0 {
		for _, dep := range comp.Dependencies {
			depComp, depDeps := convertToCycloneDXComponentWithDeps(dep, componentMap)
			
			// Add dependency bom-ref to this component's dependsOn list
			dependsOn = append(dependsOn, depComp.BOMRef)
			
			// Collect all nested dependencies
			allDependencies = append(allDependencies, depDeps...)
			
			// Add the nested component to the map
			componentMap[depComp.BOMRef] = depComp
		}
	}
	
	// Add this component's dependency relationship
	if len(dependsOn) > 0 {
		allDependencies = append(allDependencies, cdx.Dependency{
			Ref:          bomRef,
			Dependencies: &dependsOn,
		})
	} else {
		// Component with no dependencies still needs an entry
		allDependencies = append(allDependencies, cdx.Dependency{
			Ref: bomRef,
		})
	}

	return cdxComp, allDependencies
}

// generateBomRef creates a unique bom-ref for a component
func generateBomRef(comp scanners.Component) string {
	// Use Package URL (purl) format when we have package manager info
	if comp.PackageManager != "" {
		// Map package manager to purl type
		purlType := comp.PackageManager
		switch comp.PackageManager {
		case "npm":
			purlType = "npm"
		case "pip", "pip-local":
			purlType = "pypi"
		case "gem", "gem-local":
			purlType = "gem"
		case "brew":
			purlType = "brew"
		case "cargo":
			purlType = "cargo"
		case "go":
			purlType = "golang"
		case "composer":
			purlType = "composer"
		}
		
		// Build purl: pkg:type/name@version
		if comp.Version != "" {
			return fmt.Sprintf("pkg:%s/%s@%s", purlType, comp.Name, comp.Version)
		}
		return fmt.Sprintf("pkg:%s/%s", purlType, comp.Name)
	}
	
	// For applications and extensions, use a descriptive format
	if comp.Type == "application" {
		if comp.Version != "" {
			return fmt.Sprintf("app:%s@%s", comp.Name, comp.Version)
		}
		return fmt.Sprintf("app:%s", comp.Name)
	}
	
	if comp.Type == "browser-extension" {
		if comp.Version != "" {
			return fmt.Sprintf("browser-ext:%s@%s", comp.Name, comp.Version)
		}
		return fmt.Sprintf("browser-ext:%s", comp.Name)
	}
	
	if comp.Type == "ide-extension" {
		if comp.Version != "" {
			return fmt.Sprintf("ide-ext:%s@%s", comp.Name, comp.Version)
		}
		return fmt.Sprintf("ide-ext:%s", comp.Name)
	}
	
	// Fallback: use name@version or just name
	if comp.Version != "" {
		return fmt.Sprintf("%s@%s", comp.Name, comp.Version)
	}
	return comp.Name
}

// buildDescription creates an enhanced description based on component type and properties
func buildDescription(comp scanners.Component) string {
	var description string
	
	// Start with original description if available
	if comp.Description != "" {
		description = comp.Description
	}
	
	// Enhance based on component type
	switch comp.Type {
	case "ide-extension":
		// Add IDE information
		if ide := comp.Properties["ide"]; ide != "" {
			if description != "" {
				description += fmt.Sprintf(" | IDE: %s", ide)
			} else {
				description = fmt.Sprintf("IDE Extension for %s", ide)
			}
		} else {
			if description == "" {
				description = "IDE Extension"
			}
		}
		
	case "browser-extension":
		// Add browser and profile information
		browser := comp.Properties["browser"]
		profile := comp.Properties["profile"]
		
		parts := []string{}
		if browser != "" {
			parts = append(parts, fmt.Sprintf("Browser: %s", browser))
		}
		if profile != "" {
			parts = append(parts, fmt.Sprintf("Profile: %s", profile))
		}
		
		if len(parts) > 0 {
			browserInfo := fmt.Sprintf(" | %s", fmt.Sprintf("%s", parts[0]))
			if len(parts) > 1 {
				browserInfo += fmt.Sprintf(", %s", parts[1])
			}
			if description != "" {
				description += browserInfo
			} else {
				description = "Browser Extension" + browserInfo
			}
		} else if description == "" {
			description = "Browser Extension"
		}
		
	case "mcp-server":
		if description == "" {
			description = "Model Context Protocol (MCP) Server"
		} else {
			description += " | MCP Server"
		}
		
	case "library":
		// Add local project context if applicable
		if projectPath := comp.Properties["project_path"]; projectPath != "" {
			if description != "" {
				description += fmt.Sprintf(" | Local project: %s", projectPath)
			} else {
				description = fmt.Sprintf("Library from local project: %s", projectPath)
			}
		}
	}
	
	return description
}

// generateUUID generates a proper UUID v4
func generateUUID() string {
	return uuid.New().String()
}

