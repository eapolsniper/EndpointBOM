package archive

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eapolsniper/endpointbom/internal/config"
	"github.com/eapolsniper/endpointbom/internal/system"
)

// ScanMetadata contains metadata about the scan for the archive
type ScanMetadata struct {
	Hostname              string    `json:"hostname"`
	Timestamp             time.Time `json:"timestamp"`
	OSName                string    `json:"os_name"`
	OSVersion             string    `json:"os_version"`
	LocalIPs              []string  `json:"local_ips"`
	PublicIP              string    `json:"public_ip"`
	HistoricalLookbackDays int       `json:"historical_lookback_days"`
	IncludeHistorical     bool      `json:"include_historical"`
	IncludeRawLogs        bool      `json:"include_raw_logs"`
}

// CreateScanArchive creates a zip file containing SBOMs and optional logs
func CreateScanArchive(outputDir string, sysInfo *system.Info, cfg *config.Config, logFiles []string) (string, error) {
	if !cfg.CreateZipArchive {
		return "", nil
	}

	timestamp := time.Now().Format("20060102-150405")
	zipFilename := fmt.Sprintf("%s.%s.scan.zip", sysInfo.Hostname, timestamp)
	zipPath := filepath.Join(outputDir, zipFilename)

	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add metadata.json
	metadata := ScanMetadata{
		Hostname:              sysInfo.Hostname,
		Timestamp:             sysInfo.Timestamp,
		OSName:                sysInfo.OSName,
		OSVersion:             sysInfo.OSVersion,
		LocalIPs:              sysInfo.LocalIPs,
		PublicIP:              sysInfo.PublicIP,
		HistoricalLookbackDays: cfg.HistoricalLookbackDays,
		IncludeHistorical:     cfg.IncludeHistorical,
		IncludeRawLogs:        cfg.IncludeRawLogs,
	}

	if err := addJSONToZip(zipWriter, "metadata.json", metadata); err != nil {
		return "", err
	}

	// Add all CycloneDX SBOM files
	sbomFiles, err := filepath.Glob(filepath.Join(outputDir, fmt.Sprintf("%s.%s.*.cdx.json", sysInfo.Hostname, timestamp)))
	if err == nil {
		for _, sbomFile := range sbomFiles {
			if err := addFileToZip(zipWriter, sbomFile, "sboms/"+filepath.Base(sbomFile)); err != nil {
				return "", err
			}
		}
	}

	// Add log files if enabled
	if cfg.IncludeRawLogs && len(logFiles) > 0 {
		for _, logFile := range logFiles {
			// Organize logs by package manager
			pm := detectPackageManager(logFile)
			archivePath := filepath.Join("logs", pm, filepath.Base(logFile))
			if err := addFileToZip(zipWriter, logFile, archivePath); err != nil {
				// Non-fatal - skip this log and continue
				continue
			}
		}
	}

	return zipFilename, nil
}

// addFileToZip adds a file to the zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, archivePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	header.Name = archivePath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// addJSONToZip adds JSON data to the zip archive
func addJSONToZip(zipWriter *zip.Writer, archivePath string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	writer, err := zipWriter.Create(archivePath)
	if err != nil {
		return err
	}

	_, err = writer.Write(jsonData)
	return err
}

// detectPackageManager detects package manager from log file path
func detectPackageManager(logPath string) string {
	logPath = filepath.ToSlash(strings.ToLower(logPath))

	if strings.Contains(logPath, "/.npm/") {
		return "npm"
	} else if strings.Contains(logPath, "/homebrew/") || strings.Contains(logPath, "/usr/local/var/log/") {
		return "brew"
	} else if strings.Contains(logPath, "chocolatey/logs/") {
		return "chocolatey"
	} else if strings.Contains(logPath, "/.pip/") {
		return "pip"
	}

	return "other"
}

