package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/eapolsniper/endpointbom/internal/archive"
	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/sbom"
	"github.com/eapolsniper/endpointbom/internal/scanners"
	"github.com/eapolsniper/endpointbom/internal/scanners/applications"
	"github.com/eapolsniper/endpointbom/internal/scanners/browsers"
	"github.com/eapolsniper/endpointbom/internal/scanners/historical"
	"github.com/eapolsniper/endpointbom/internal/scanners/ides"
	"github.com/eapolsniper/endpointbom/internal/scanners/packagemanagers"
	"github.com/eapolsniper/endpointbom/internal/security"
	"github.com/eapolsniper/endpointbom/internal/system"
	"github.com/eapolsniper/endpointbom/internal/version"
)

var (
	cfgFile            string
	outputDir          string
	debug              bool
	verbose            bool
	requireAdmin       bool
	scanAllUsers       bool
	excludePaths       []string
	disabledScanners   []string
	enabledScanners    []string
	disablePublicIP    bool
	fetchPublicIP      bool
	enableAll          bool // Enable all optional features (browser extensions, public IP)
	historicalDays     int
	noHistorical       bool
	noRawLogs          bool
	noZip              bool
	showVersion        bool
)

var rootCmd = &cobra.Command{
	Use:   "endpointbom",
	Short: "EndpointBOM - Endpoint Bill of Materials Scanner",
	Long: `EndpointBOM scans developer endpoints for installed software, package managers,
IDE extensions, and MCP servers to generate comprehensive CycloneDX SBOMs.

The tool scans:
  - Package managers (npm, pip, yarn, brew, gem, cargo, composer, chocolatey, etc.)
  - Installed applications (all non-OS applications)
  - IDE extensions and plugins (VSCode, Cursor, JetBrains, Sublime, etc.)
  - MCP servers configured in supported IDEs

Security features:
  - No secrets or environment variables are collected
  - Uses minimal, trusted dependencies
  - Dependency pinning for security

Copyright © Tim Jensen (EapolSniper)
GitHub: https://github.com/eapolsniper/endpointbom`,
	RunE: runScan,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./endpointbom.yaml)")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output", "", "output directory for SBOM files (default is ./scans relative to executable)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&requireAdmin, "require-admin", false, "require admin/root privileges (fail if not admin)")
	rootCmd.PersistentFlags().BoolVar(&scanAllUsers, "scan-all-users", true, "scan all user profiles (requires admin, default: true)")
	rootCmd.PersistentFlags().StringSliceVar(&excludePaths, "exclude", []string{}, "paths to exclude from scanning")
	rootCmd.PersistentFlags().StringSliceVar(&disabledScanners, "disable", []string{}, "scanners to disable (e.g., npm,pip,vscode)")
	rootCmd.PersistentFlags().StringSliceVar(&enabledScanners, "enable", []string{}, "scanners to enable (e.g., browser-extensions, chrome-extensions, local-projects)")
	rootCmd.PersistentFlags().BoolVar(&fetchPublicIP, "fetch-public-ip", false, "enable public IP address gathering from external services (disabled by default)")
	rootCmd.PersistentFlags().BoolVar(&enableAll, "all", false, "enable all optional features (browser extensions, public IP lookup)")
	rootCmd.PersistentFlags().IntVar(&historicalDays, "historical-days", 30, "days to look back for historical package installations")
	rootCmd.PersistentFlags().BoolVar(&noHistorical, "no-historical", false, "disable historical package tracking")
	
	// Deprecated flag - hidden from help but still functional for backward compatibility
	rootCmd.PersistentFlags().BoolVar(&disablePublicIP, "disable-public-ip", false, "deprecated: use --fetch-public-ip instead")
	rootCmd.PersistentFlags().MarkHidden("disable-public-ip")
	rootCmd.PersistentFlags().BoolVar(&noRawLogs, "no-raw-logs", false, "don't include raw log files in zip archive")
	rootCmd.PersistentFlags().BoolVar(&noZip, "no-zip", false, "don't create zip archive")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version information")
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = "endpointbom.yaml"
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	// Handle version flag
	if showVersion {
		fmt.Println(version.Info())
		return nil
	}

	// Validate and sanitize config file path
	validatedCfgFile := cfgFile
	if cfgFile != "" {
		var err error
		validatedCfgFile, err = security.ValidateConfigPath(cfgFile)
		if err != nil {
			return fmt.Errorf("invalid config file path: %w", err)
		}
	}
	
	// Load configuration
	cfg, err := config.LoadFromFile(validatedCfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine output directory
	if cmd.Flags().Changed("output") {
		cfg.OutputDir = outputDir
	} else if cfg.OutputDir == "" {
		// Use default: scans/ directory next to executable
		defaultOutput, err := security.GetDefaultOutputDir()
		if err != nil {
			return fmt.Errorf("failed to determine default output directory: %w", err)
		}
		cfg.OutputDir = defaultOutput
	}
	
	// Validate output directory
	validatedOutput, err := security.ValidateOutputDirectory(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}
	cfg.OutputDir = validatedOutput
	
	// Override other config settings with CLI flags
	if cmd.Flags().Changed("debug") {
		cfg.Debug = debug
	}
	if cmd.Flags().Changed("verbose") {
		cfg.Verbose = verbose
	}
	if cmd.Flags().Changed("require-admin") {
		cfg.RequireAdmin = requireAdmin
	}
	if cmd.Flags().Changed("scan-all-users") {
		cfg.ScanAllUsers = scanAllUsers
	}
	if cmd.Flags().Changed("exclude") {
		cfg.ExcludePaths = append(cfg.ExcludePaths, excludePaths...)
	}
	if cmd.Flags().Changed("disable") {
		cfg.DisabledScanners = append(cfg.DisabledScanners, disabledScanners...)
	}
	// Handle --all flag (enables all optional features)
	if cmd.Flags().Changed("all") && enableAll {
		// Enable browser extensions
		browserExtensions := []string{"chrome-extensions", "firefox-extensions", "edge-extensions", "safari-extensions"}
		for _, scanner := range browserExtensions {
			cfg.DisabledScanners = removeFromSlice(cfg.DisabledScanners, scanner)
		}
		// Enable public IP lookup
		cfg.DisablePublicIP = false
	}

	if cmd.Flags().Changed("enable") {
		// Expand shorthand groups
		expandedScanners := expandScannerGroups(enabledScanners)
		
		// Remove enabled scanners from the disabled list
		for _, enableScanner := range expandedScanners {
			cfg.DisabledScanners = removeFromSlice(cfg.DisabledScanners, enableScanner)
		}
	}
	// Handle public IP flags (--fetch-public-ip takes precedence over --disable-public-ip)
	if cmd.Flags().Changed("fetch-public-ip") {
		cfg.DisablePublicIP = !fetchPublicIP
	} else if cmd.Flags().Changed("disable-public-ip") {
		cfg.DisablePublicIP = disablePublicIP
	}
	if cmd.Flags().Changed("historical-days") {
		cfg.HistoricalLookbackDays = historicalDays
	}
	if cmd.Flags().Changed("no-historical") {
		cfg.IncludeHistorical = !noHistorical
	}
	if cmd.Flags().Changed("no-raw-logs") {
		cfg.IncludeRawLogs = !noRawLogs
	}
	if cmd.Flags().Changed("no-zip") {
		cfg.CreateZipArchive = !noZip
	}

	// Check admin privileges and adjust behavior accordingly
	isAdmin := system.IsAdmin()
	
	if cfg.RequireAdmin && !isAdmin {
		return fmt.Errorf("this tool requires administrator/root privileges. Please run with sudo or as administrator")
	}

	// Auto-adjust scan scope based on privileges
	if !isAdmin {
		if cfg.ScanAllUsers {
			fmt.Println("⚠️  WARNING: Not running as administrator/root")
			fmt.Println("    Scan will be limited to current user profile only")
			fmt.Println("    For complete endpoint inventory, run with sudo (macOS/Linux) or as Administrator (Windows)")
			fmt.Println()
			cfg.ScanAllUsers = false
		}
	} else {
		// Running as admin
		if cfg.ScanAllUsers {
			fmt.Println("✓ Running with administrator privileges - scanning all user profiles")
		} else {
			fmt.Println("ℹ️  Scanning current user only (use --scan-all-users=true for all users)")
		}
		fmt.Println()
	}

	// Get system information
	fmt.Println("Gathering system information...")
	sysInfo, err := system.GetSystemInfo(cfg.DisablePublicIP)
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("Hostname: %s\n", sysInfo.Hostname)
		fmt.Printf("OS: %s %s\n", sysInfo.OSName, sysInfo.OSVersion)
		fmt.Printf("Logged in users: %v\n", sysInfo.Users)
		if len(sysInfo.LocalIPs) > 0 {
			fmt.Printf("Local IP(s): %v\n", sysInfo.LocalIPs)
		}
		if sysInfo.PublicIP != "unavailable" {
			fmt.Printf("Public IP: %s\n", sysInfo.PublicIP)
		}
		fmt.Printf("Output directory: %s\n", cfg.OutputDir)
	}

	// Initialize scanners
	allScanners := []scanners.Scanner{
		// Package managers (global)
		&packagemanagers.NPMScanner{},
		&packagemanagers.PipScanner{},
		&packagemanagers.YarnScanner{},
		&packagemanagers.PnpmScanner{},
		&packagemanagers.BrewScanner{},
		&packagemanagers.GemScanner{},
		&packagemanagers.CargoScanner{},
		&packagemanagers.ComposerScanner{},
		&packagemanagers.ChocolateyScanner{},
		&packagemanagers.GoScanner{},

		// Package managers (local projects)
		&packagemanagers.NPMLocalScanner{},
		&packagemanagers.PipLocalScanner{},
		&packagemanagers.GemLocalScanner{},

		// Applications
		&applications.ApplicationScanner{},

		// IDEs
		&ides.VSCodeScanner{},
		&ides.CursorScanner{},
		&ides.JetBrainsScanner{},
		&ides.SublimeScanner{},

		// Browser Extensions
		&browsers.ChromeScanner{},
		&browsers.FirefoxScanner{},
		&browsers.EdgeScanner{},
		&browsers.SafariScanner{},

		// Historical tracking (best-effort from logs)
		&historical.NPMHistoricalScanner{},
		&historical.BrewHistoricalScanner{},
	}

	// Run all scanners - separate current from historical
	result := &scanners.ScanResult{
		Applications:      []scanners.Component{},
		PackageManagers:   []scanners.Component{},
		IDEExtensions:     []scanners.Component{},
		BrowserExtensions: []scanners.Component{},
	}

	var historicalComponents []scanners.Component

	for _, scanner := range allScanners {
		if cfg.IsScannerDisabled(scanner.Name()) {
			if cfg.Verbose {
				fmt.Printf("Skipping disabled scanner: %s\n", scanner.Name())
			}
			continue
		}

		fmt.Printf("Running scanner: %s\n", scanner.Name())
		components, err := scanner.Scan(cfg)
		if err != nil {
			if cfg.Debug {
				fmt.Printf("Scanner %s error: %v\n", scanner.Name(), err)
			}
			continue
		}

		if cfg.Verbose {
			fmt.Printf("  Found %d components\n", len(components))
		}

		// Separate historical scanners from current scanners
		isHistorical := strings.HasSuffix(scanner.Name(), "-historical")
		
		if isHistorical {
			// Store historical components for later deduplication
			historicalComponents = append(historicalComponents, components...)
		} else {
			// Categorize current components
			for _, comp := range components {
				// Mark as current (actively installed)
				if comp.Properties == nil {
					comp.Properties = make(map[string]string)
				}
				// Only set install_type if not already set (preserve local scanner properties)
				if _, exists := comp.Properties["install_type"]; !exists {
					comp.Properties["install_type"] = "current"
				}
				// Only set source if not already set
				if _, exists := comp.Properties["source"]; !exists {
					comp.Properties["source"] = scanner.Name()
				}

				switch comp.Type {
				case "application":
					// Check if it's from a package manager
					if comp.PackageManager != "" {
						result.PackageManagers = append(result.PackageManagers, comp)
					} else {
						result.Applications = append(result.Applications, comp)
					}
				case "library":
					result.PackageManagers = append(result.PackageManagers, comp)
				case "browser-extension":
					result.BrowserExtensions = append(result.BrowserExtensions, comp)
				case "ide-extension", "mcp-server":
					result.IDEExtensions = append(result.IDEExtensions, comp)
				default:
					result.Applications = append(result.Applications, comp)
				}
			}
		}
	}

	// Deduplicate historical components
	// Only add historical components that are NOT currently installed
	currentPackages := buildPackageSet(result.PackageManagers)
	
	for _, histComp := range historicalComponents {
		key := fmt.Sprintf("%s:%s:%s", histComp.Name, histComp.Version, histComp.PackageManager)
		
		// If this package@version is NOT currently installed, add it as historical
		if !currentPackages[key] {
			if histComp.Properties == nil {
				histComp.Properties = make(map[string]string)
			}
			// Override install_type to "historical" (not currently installed)
			histComp.Properties["install_type"] = "historical"
			
			result.PackageManagers = append(result.PackageManagers, histComp)
		}
		// If it IS currently installed, we already have it from the current scan
	}

	// Print summary
	fmt.Println("\n=== Scan Summary ===")
	fmt.Printf("Package Manager Components: %d\n", len(result.PackageManagers))
	fmt.Printf("Applications: %d\n", len(result.Applications))
	fmt.Printf("IDE Extensions/Plugins: %d\n", len(result.IDEExtensions))
	fmt.Printf("Browser Extensions: %d\n", len(result.BrowserExtensions))
	fmt.Printf("Output Directory: %s\n", cfg.OutputDir)

	// Generate SBOMs
	fmt.Println("\n=== Generating SBOMs ===")
	if err := sbom.GenerateSBOMs(result, sysInfo, cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to generate SBOMs: %w", err)
	}

	// Collect log files and create zip archive
	if cfg.CreateZipArchive {
		fmt.Println("\n=== Creating Archive ===")
		
		var allLogFiles []string
		
		// Collect logs from historical scanners
		npmHistorical := &historical.NPMHistoricalScanner{}
		if logs, err := npmHistorical.GetLogFiles(cfg); err == nil {
			allLogFiles = append(allLogFiles, logs...)
		}
		
		brewHistorical := &historical.BrewHistoricalScanner{}
		if logs, err := brewHistorical.GetLogFiles(cfg); err == nil {
			allLogFiles = append(allLogFiles, logs...)
		}

		zipFilename, err := archive.CreateScanArchive(cfg.OutputDir, sysInfo, cfg, allLogFiles)
		if err != nil {
			fmt.Printf("Warning: Failed to create zip archive: %v\n", err)
		} else if zipFilename != "" {
			fmt.Printf("Created archive: %s\n", zipFilename)
		}
	}

	fmt.Println("\n✓ Scan complete!")
	return nil
}

// buildPackageSet creates a set of currently installed packages for deduplication
func buildPackageSet(components []scanners.Component) map[string]bool {
	set := make(map[string]bool)
	for _, comp := range components {
		key := fmt.Sprintf("%s:%s:%s", comp.Name, comp.Version, comp.PackageManager)
		set[key] = true
	}
	return set
}

// expandScannerGroups expands shorthand scanner groups into individual scanners
func expandScannerGroups(scanners []string) []string {
	var expanded []string
	
	for _, scanner := range scanners {
		switch scanner {
		case "browser-extensions":
			// Expand to all browser extension scanners
			expanded = append(expanded, "chrome-extensions")
			expanded = append(expanded, "firefox-extensions")
			expanded = append(expanded, "edge-extensions")
			expanded = append(expanded, "safari-extensions")
		case "local-projects":
			// Expand to all local project scanners
			expanded = append(expanded, "npm-local")
			expanded = append(expanded, "pip-local")
			expanded = append(expanded, "gem-local")
		default:
			// Not a group, add as-is
			expanded = append(expanded, scanner)
		}
	}
	
	return expanded
}

// removeFromSlice removes all occurrences of a value from a string slice
func removeFromSlice(slice []string, value string) []string {
	result := []string{}
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

