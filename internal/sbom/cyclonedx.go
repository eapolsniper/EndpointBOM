package sbom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
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
	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Component: &cdx.Component{
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

	// Convert components to CycloneDX components
	var cdxComponents []cdx.Component
	for _, comp := range components {
		cdxComp := convertToCycloneDXComponent(comp)
		cdxComponents = append(cdxComponents, cdxComp)
	}

	if len(cdxComponents) > 0 {
		bom.Components = &cdxComponents
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

func convertToCycloneDXComponent(comp scanners.Component) cdx.Component {
	cdxComp := cdx.Component{
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

	// Add description
	if comp.Description != "" {
		cdxComp.Description = comp.Description
	}

	// Add group/namespace if available
	if comp.Group != "" {
		cdxComp.Group = comp.Group
	}

	// Add properties
	if len(comp.Properties) > 0 {
		var props []cdx.Property
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
		
		cdxComp.Properties = &props
	}

	// Add dependencies
	if len(comp.Dependencies) > 0 {
		var deps []cdx.Component
		for _, dep := range comp.Dependencies {
			deps = append(deps, convertToCycloneDXComponent(dep))
		}
		cdxComp.Components = &deps
	}

	return cdxComp
}

// Simple UUID generation (for demo purposes - in production use proper UUID library)
func generateUUID() string {
	return fmt.Sprintf("%d-%d-%d-%d-%d",
		time.Now().UnixNano()%10000,
		time.Now().UnixNano()%10000,
		time.Now().UnixNano()%10000,
		time.Now().UnixNano()%10000,
		time.Now().UnixNano()%10000,
	)
}

