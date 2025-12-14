package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	encodedURL = ""
	encodedKey = ""
)

const (
	defaultURL = "http://localhost:8081"
	defaultKey = "YOURAPIKEYHERE"
)

const xorKey = 0x5A

// SBOM type configurations
var sbomTypes = map[string]SBOMTypeConfig{
	"package-managers": {
		Classifier:  "LIBRARY",
		Description: "Dependencies from package managers (npm, pip, etc.)",
	},
	"applications": {
		Classifier:  "APPLICATION",
		Description: "Installed desktop applications",
	},
	"ide-extensions": {
		Classifier:  "LIBRARY",
		Description: "IDE extensions and MCP servers",
	},
	"browser-extensions": {
		Classifier:  "LIBRARY",
		Description: "Browser extensions and plugins",
	},
}

type SBOMTypeConfig struct {
	Classifier  string
	Description string
}

type DependencyTrackClient struct {
	BaseURL string
	APIKey  string
	Headers map[string]string
}

type Project struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Classifier  string `json:"classifier"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Parent      *struct {
		UUID string `json:"uuid"`
	} `json:"parent,omitempty"`
}

type BOMMetadata struct {
	Hostname       string
	OS             string
	OSVersion      string
	ScanCategory   string
	LoggedInUser   string
	LocalIPs       []string
	PublicIP       string
	Timestamp      string
	ComponentCount int
}

func DecodeConfig() (url, apiKey string) {
	if encodedURL != "" {
		decoded, err := base64.StdEncoding.DecodeString(encodedURL)
		if err == nil {
			url = string(decoded)
		}
	}
	if url == "" {
		url = defaultURL
	}

	if encodedKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(encodedKey)
		if err == nil {
			apiKey = xorDecode(decoded)
		}
	}
	if apiKey == "" {
		apiKey = defaultKey
	}

	return url, apiKey
}

func xorEncode(input []byte) []byte {
	output := make([]byte, len(input))
	for i, b := range input {
		output[i] = b ^ xorKey
	}
	return output
}

func xorDecode(input []byte) string {
	output := make([]byte, len(input))
	for i, b := range input {
		output[i] = b ^ xorKey
	}
	return string(output)
}

// NewClient creates a new Dependency-Track client
func NewClient(baseURL, apiKey string) *DependencyTrackClient {
	return &DependencyTrackClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Headers: map[string]string{
			"X-Api-Key":    apiKey,
			"Content-Type": "application/json",
		},
	}
}

// CreateProject creates a new project in Dependency-Track
func (c *DependencyTrackClient) CreateProject(name, version, classifier, description string, parentUUID *string) (*Project, error) {
	fmt.Printf("\n📦 Creating project: %s v%s\n", name, version)
	fmt.Printf("   Classifier: %s\n", classifier)
	if parentUUID != nil {
		fmt.Printf("   Parent UUID: %s\n", *parentUUID)
	}

	payload := map[string]interface{}{
		"name":        name,
		"version":     version,
		"classifier":  classifier,
		"description": description,
		"active":      true,
	}

	if parentUUID != nil {
		payload["parent"] = map[string]string{"uuid": *parentUUID}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", c.BaseURL+"/api/v1/project", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		var project Project
		if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
			return nil, err
		}
		fmt.Printf("   ✅ Created project UUID: %s\n", project.UUID)
		return &project, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return nil, fmt.Errorf("failed to create project: %d - %s", resp.StatusCode, string(bodyBytes))
}

// GetProject looks up an existing project
func (c *DependencyTrackClient) GetProject(name, version string) (*Project, error) {
	url := fmt.Sprintf("%s/api/v1/project/lookup?name=%s&version=%s", c.BaseURL, name, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var project Project
		if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
			return nil, err
		}
		return &project, nil
	} else if resp.StatusCode == 404 {
		return nil, nil
	}

	return nil, fmt.Errorf("error looking up project: %d", resp.StatusCode)
}

// UploadBOM uploads a BOM file to a project
func (c *DependencyTrackClient) UploadBOM(projectUUID string, bomFile string) (string, error) {
	fmt.Printf("\n📤 Uploading BOM: %s\n", filepath.Base(bomFile))
	fmt.Printf("   Project UUID: %s\n", projectUUID)

	// Read BOM file
	bomContent, err := os.ReadFile(bomFile)
	if err != nil {
		return "", fmt.Errorf("failed to read BOM file: %w", err)
	}

	// Encode as base64
	bomBase64 := base64.StdEncoding.EncodeToString(bomContent)

	payload := map[string]interface{}{
		"project":    projectUUID,
		"bom":        bomBase64,
		"autoCreate": false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", c.BaseURL+"/api/v1/bom", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		token, ok := result["token"].(string)
		if !ok {
			return "", fmt.Errorf("no token in response")
		}
		fmt.Printf("   ✅ BOM uploaded successfully\n")
		fmt.Printf("   📋 Processing token: %s\n", token)
		return token, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("failed to upload BOM: %d - %s", resp.StatusCode, string(bodyBytes))
}

// CheckBOMProcessingStatus checks if a BOM has been processed
func (c *DependencyTrackClient) CheckBOMProcessingStatus(token string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/bom/token/%s", c.BaseURL, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, err
		}
		processing, ok := result["processing"].(bool)
		if !ok {
			return false, nil
		}
		return !processing, nil
	}

	return false, fmt.Errorf("failed to check status: %d", resp.StatusCode)
}

// ExtractMetadata extracts metadata from a BOM file
func ExtractMetadata(bomFile string) (*BOMMetadata, error) {
	content, err := os.ReadFile(bomFile)
	if err != nil {
		return nil, err
	}

	var bom map[string]interface{}
	if err := json.Unmarshal(content, &bom); err != nil {
		return nil, err
	}

	metadata := &BOMMetadata{}

	if meta, ok := bom["metadata"].(map[string]interface{}); ok {
		if comp, ok := meta["component"].(map[string]interface{}); ok {
			if name, ok := comp["name"].(string); ok {
				metadata.Hostname = name
			}

			if props, ok := comp["properties"].([]interface{}); ok {
				for _, p := range props {
					if prop, ok := p.(map[string]interface{}); ok {
						name := prop["name"].(string)
						value := prop["value"].(string)

						switch name {
						case "os":
							metadata.OS = value
						case "os_version":
							metadata.OSVersion = value
						case "scan_category":
							metadata.ScanCategory = value
						case "logged_in_user":
							metadata.LoggedInUser = value
						case "local_ip":
							metadata.LocalIPs = append(metadata.LocalIPs, value)
						case "public_ip":
							metadata.PublicIP = value
						}
					}
				}
			}
		}

		if timestamp, ok := meta["timestamp"].(string); ok {
			metadata.Timestamp = timestamp
		}
	}

	if components, ok := bom["components"].([]interface{}); ok {
		metadata.ComponentCount = len(components)
	}

	return metadata, nil
}

// FormatVersionFromTimestamp converts ISO timestamp to version format
func FormatVersionFromTimestamp(timestamp string) string {
	if timestamp == "" {
		return time.Now().Format("2006-01-02-1504")
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return time.Now().Format("2006-01-02-1504")
	}

	return t.Format("2006-01-02-1504")
}

// getMostRecentScanFiles finds the most recent scan files by timestamp
func getMostRecentScanFiles(scansDir string) ([]string, error) {
	allFiles, err := filepath.Glob(filepath.Join(scansDir, "*.cdx.json"))
	if err != nil {
		return nil, err
	}

	if len(allFiles) == 0 {
		return nil, nil
	}

	// Extract timestamps from filenames (format: hostname.YYYYMMDD-HHmmss.type.cdx.json)
	fileTimestamps := make(map[string]time.Time)
	for _, file := range allFiles {
		baseName := filepath.Base(file)
		parts := strings.Split(strings.TrimSuffix(baseName, ".cdx.json"), ".")
		
		if len(parts) >= 2 {
			timestampStr := parts[1]
			// Parse timestamp (format: 20251214-112345)
			timestamp, err := time.Parse("20060102-150405", timestampStr)
			if err != nil {
				fmt.Printf("⚠️  Skipping file with invalid timestamp format: %s\n", baseName)
				continue
			}
			fileTimestamps[file] = timestamp
		}
	}

	if len(fileTimestamps) == 0 {
		return nil, nil
	}

	// Find the most recent timestamp
	var mostRecentTime time.Time
	for _, ts := range fileTimestamps {
		if ts.After(mostRecentTime) {
			mostRecentTime = ts
		}
	}

	// Get all files matching the most recent timestamp
	var recentFiles []string
	for file, ts := range fileTimestamps {
		if ts.Equal(mostRecentTime) {
			recentFiles = append(recentFiles, file)
		}
	}

	return recentFiles, nil
}

// archiveFiles moves uploaded files to archive subdirectory
func archiveFiles(files []string, scansDir string) error {
	archiveDir := filepath.Join(scansDir, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	fmt.Printf("\n📦 Archiving %d uploaded files...\n", len(files))
	for _, file := range files {
		baseName := filepath.Base(file)
		dest := filepath.Join(archiveDir, baseName)
		if err := os.Rename(file, dest); err != nil {
			return err
		}
		fmt.Printf("   ✅ Moved to archive: %s\n", baseName)
	}

	return nil
}

// cleanupOldFiles removes files older than specified days from scans/ and scans/archive/
func cleanupOldFiles(scansDir string, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	fmt.Printf("\n🧹 Cleaning up files older than %d days...\n", days)
	
	// Check both main scans directory and archive
	directories := []string{scansDir, filepath.Join(scansDir, "archive")}
	
	totalRemoved := 0
	for _, dir := range directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		files, err := filepath.Glob(filepath.Join(dir, "*.cdx.json"))
		if err != nil {
			continue
		}

		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoffTime) {
				if err := os.Remove(file); err == nil {
					relPath, _ := filepath.Rel(filepath.Dir(scansDir), file)
					fmt.Printf("   🗑️  Removed old file: %s\n", relPath)
					totalRemoved++
				}
			}
		}
	}

	if totalRemoved == 0 {
		fmt.Println("   ✅ No old files to remove")
	} else {
		fmt.Printf("   ✅ Removed %d old file(s)\n", totalRemoved)
	}

	return nil
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("Dependency-Track SBOM Upload Tool")
	fmt.Println("=" + strings.Repeat("=", 79))

	// Decode configuration
	dtURL, apiKey := DecodeConfig()

	fmt.Printf("\n🔗 Dependency-Track URL: %s\n", dtURL)

	// Initialize client
	client := NewClient(dtURL, apiKey)

	// Find scans directory
	scansDir := "scans"
	if len(os.Args) > 1 {
		scansDir = os.Args[1]
	}

	if _, err := os.Stat(scansDir); os.IsNotExist(err) {
		fmt.Printf("❌ Scans directory not found: %s\n", scansDir)
		os.Exit(1)
	}

	// Get only the most recent scan files
	fmt.Printf("\n🔍 Looking for most recent scan files in %s...\n", scansDir)
	files, err := getMostRecentScanFiles(scansDir)
	if err != nil {
		fmt.Printf("❌ Error finding SBOM files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("❌ No valid SBOM files found in %s\n", scansDir)
		os.Exit(1)
	}

	// Extract timestamp from first file for logging
	firstFileBase := filepath.Base(files[0])
	parts := strings.Split(strings.TrimSuffix(firstFileBase, ".cdx.json"), ".")
	scanTimestamp := "unknown"
	if len(parts) >= 2 {
		scanTimestamp = parts[1]
	}

	fmt.Printf("\n📁 Found most recent scan: %s\n", scanTimestamp)
	fmt.Printf("   Files to upload (%d):\n", len(files))
	for _, file := range files {
		fmt.Printf("   - %s\n", filepath.Base(file))
	}

	// Group files by hostname
	hostnameGroups := make(map[string][]string)
	for _, file := range files {
		metadata, err := ExtractMetadata(file)
		if err != nil {
			continue
		}
		if metadata.Hostname != "" {
			hostnameGroups[metadata.Hostname] = append(hostnameGroups[metadata.Hostname], file)
		}
	}

	fmt.Printf("\n🖥️  Found %d unique hostnames:\n", len(hostnameGroups))
	for hostname, files := range hostnameGroups {
		fmt.Printf("   - %s (%d BOMs)\n", hostname, len(files))
	}

	// Process each hostname
	for hostname, bomFiles := range hostnameGroups {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Printf("Processing hostname: %s\n", hostname)
		fmt.Println(strings.Repeat("=", 80))

		// Get metadata from first BOM
		firstMetadata, _ := ExtractMetadata(bomFiles[0])

		fmt.Println("\n📊 Endpoint Metadata:")
		fmt.Printf("   Hostname: %s\n", firstMetadata.Hostname)
		fmt.Printf("   OS: %s %s\n", firstMetadata.OS, firstMetadata.OSVersion)
		fmt.Printf("   User: %s\n", firstMetadata.LoggedInUser)
		fmt.Printf("   Scan Time: %s\n", firstMetadata.Timestamp)

		// Create parent project
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("STEP 1: CREATE PARENT PROJECT")
		fmt.Println(strings.Repeat("-", 80))

		parentName := hostname
		parentVersion := "latest"

		parentProject, err := client.GetProject(parentName, parentVersion)
		if err != nil {
			fmt.Printf("❌ Error looking up parent: %v\n", err)
			continue
		}

		if parentProject == nil {
			parentProject, err = client.CreateProject(
				parentName,
				parentVersion,
				"DEVICE",
				fmt.Sprintf("Developer workstation: %s", hostname),
				nil,
			)
			if err != nil {
				fmt.Printf("❌ Failed to create parent: %v\n", err)
				continue
			}
		} else {
			fmt.Printf("\n📦 Parent project already exists: %s\n", parentProject.UUID)
		}

		// Process child projects
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("STEP 2: CREATE CHILD PROJECTS & UPLOAD BOMs")
		fmt.Println(strings.Repeat("-", 80))

		var uploadTokens []map[string]string

		for _, bomFile := range bomFiles {
			// Determine SBOM type
			var sbomType string
			baseName := filepath.Base(bomFile)
			for typeKey := range sbomTypes {
				if strings.Contains(baseName, typeKey) {
					sbomType = typeKey
					break
				}
			}

			if sbomType == "" {
				fmt.Printf("\n⚠️  Could not determine SBOM type for %s, skipping\n", baseName)
				continue
			}

			typeConfig := sbomTypes[sbomType]
			bomMetadata, _ := ExtractMetadata(bomFile)

			childName := fmt.Sprintf("%s - %s", hostname, sbomType)
			childVersion := FormatVersionFromTimestamp(bomMetadata.Timestamp)

			fmt.Printf("\n   Creating version: %s for %s\n", childVersion, sbomType)

			childProject, err := client.GetProject(childName, childVersion)
			if err != nil {
				fmt.Printf("❌ Error looking up child: %v\n", err)
				continue
			}

			if childProject == nil {
				childProject, err = client.CreateProject(
					childName,
					childVersion,
					typeConfig.Classifier,
					typeConfig.Description,
					&parentProject.UUID,
				)
				if err != nil {
					fmt.Printf("❌ Failed to create child: %v\n", err)
					continue
				}
			} else {
				fmt.Printf("\n📦 Child project already exists: %s\n", childProject.UUID)
			}

			// Upload BOM
			token, err := client.UploadBOM(childProject.UUID, bomFile)
			if err != nil {
				fmt.Printf("❌ Failed to upload BOM: %v\n", err)
				continue
			}

			uploadTokens = append(uploadTokens, map[string]string{
				"token":   token,
				"project": childName,
				"type":    sbomType,
			})
		}

		// Monitor processing
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("STEP 3: MONITORING BOM PROCESSING")
		fmt.Println(strings.Repeat("-", 80))

		fmt.Printf("\n⏳ Waiting for %d BOMs to process...\n", len(uploadTokens))
		fmt.Println("   (This may take a few moments)\n")

		for _, item := range uploadTokens {
			maxAttempts := 30
			for attempt := 0; attempt < maxAttempts; attempt++ {
				complete, err := client.CheckBOMProcessingStatus(item["token"])
				if err == nil && complete {
					fmt.Printf("   ✅ %s - Processing complete\n", item["project"])
					break
				}
				if attempt >= maxAttempts-1 {
					fmt.Printf("   ⏰ %s - Still processing (timeout)\n", item["project"])
				}
				time.Sleep(2 * time.Second)
			}
		}

		// Summary
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("STEP 4: UPLOAD SUMMARY")
		fmt.Println(strings.Repeat("-", 80))

		fmt.Printf("\n✅ Successfully uploaded %d BOMs for %s\n", len(uploadTokens), hostname)
		fmt.Printf("\n🔗 View in Dependency-Track:\n")
		fmt.Printf("   %s/projects/%s\n", dtURL, parentProject.UUID)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ All uploads complete!")
	fmt.Println(strings.Repeat("=", 80))

	// Archive uploaded files
	if err := archiveFiles(files, scansDir); err != nil {
		fmt.Printf("⚠️  Warning: Failed to archive files: %v\n", err)
	}

	// Cleanup old files (>60 days)
	if err := cleanupOldFiles(scansDir, 60); err != nil {
		fmt.Printf("⚠️  Warning: Failed to cleanup old files: %v\n", err)
	}
}

