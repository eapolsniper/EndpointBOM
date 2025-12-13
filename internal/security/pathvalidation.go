package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SensitivePaths contains paths that should never be accessed
var SensitivePaths = []string{
	// Unix password/authentication files
	"/etc/shadow",
	"/etc/gshadow",
	"/etc/passwd-",
	"/etc/shadow-",
	"/etc/security/opasswd",
	
	// SSH keys
	"/root/.ssh",
	".ssh/id_rsa",
	".ssh/id_ecdsa",
	".ssh/id_ed25519",
	".ssh/id_dsa",
	
	// GPG/PGP keys
	".gnupg/",
	".gnupg/secring.gpg",
	".gnupg/private-keys-v1.d/",
	
	// AWS credentials
	".aws/credentials",
	".aws/config",
	
	// Kubernetes configs
	".kube/config",
	
	// Docker credentials
	".docker/config.json",
	
	// Database files
	".mysql_history",
	".psql_history",
	
	// Shell history (may contain secrets)
	".bash_history",
	".zsh_history",
	".sh_history",
	
	// Windows sensitive files
	"C:\\Windows\\System32\\config\\SAM",
	"C:\\Windows\\System32\\config\\SECURITY",
	"C:\\Windows\\System32\\config\\SYSTEM",
	"C:\\Windows\\repair\\SAM",
	"C:\\Windows\\repair\\SECURITY",
	
	// Credential stores
	".netrc",
	".git-credentials",
	
	// Certificate private keys
	".pem",
	".key",
	"*.key",
	"*-key.pem",
	"privkey.pem",
}

// SensitiveDirectories are directories that should not be written to
var SensitiveDirectories = []string{
	// System directories
	"/etc",
	"/bin",
	"/sbin",
	"/usr/bin",
	"/usr/sbin",
	"/boot",
	"/sys",
	"/proc",
	"C:\\Windows",
	"C:\\Windows\\System32",
	"C:\\Program Files",
	
	// Root directories (dangerous to write to)
	"/",
	"C:\\",
	
	// System Library directories
	"/usr/lib",
	"/usr/local/lib",
	"/Library/",
	"C:\\Windows\\System",
}

// ValidatePath validates and sanitizes a file path
func ValidatePath(path string, purpose string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path provided")
	}
	
	// Clean the path to remove .., ./, etc.
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path '%s': %w", path, err)
	}
	
	// Check if path is sensitive based on purpose
	if purpose == "read" || purpose == "config" {
		if err := checkSensitivePathRead(absPath); err != nil {
			return "", err
		}
	}
	
	if purpose == "write" || purpose == "output" {
		if err := checkSensitivePathWrite(absPath); err != nil {
			return "", err
		}
	}
	
	return absPath, nil
}

// checkSensitivePathRead checks if reading from a path would expose sensitive data
func checkSensitivePathRead(path string) error {
	// Check against sensitive file patterns
	for _, sensitive := range SensitivePaths {
		if matchPath(path, sensitive) {
			return fmt.Errorf("access denied: '%s' is a sensitive file/path", path)
		}
	}
	
	// Check if trying to read from sensitive directories
	for _, sensitiveDir := range SensitiveDirectories {
		absDir, err := filepath.Abs(sensitiveDir)
		if err != nil {
			continue
		}
		
		// Normalize for comparison
		if runtime.GOOS == "windows" {
			path = strings.ToLower(path)
			absDir = strings.ToLower(absDir)
		}
		
		// Check if path is within sensitive directory
		if strings.HasPrefix(path, absDir+string(filepath.Separator)) || path == absDir {
			// Only block if it's a direct match, allow subdirectories for config files
			if strings.Contains(filepath.Base(path), "credentials") ||
			   strings.Contains(filepath.Base(path), "secret") ||
			   strings.Contains(filepath.Base(path), "password") ||
			   strings.Contains(filepath.Base(path), "private") {
				return fmt.Errorf("access denied: '%s' may contain sensitive data", path)
			}
		}
	}
	
	return nil
}

// checkSensitivePathWrite checks if writing to a path could overwrite sensitive files
func checkSensitivePathWrite(path string) error {
	// Check if trying to write to system directories
	for _, sensitiveDir := range SensitiveDirectories {
		absDir, err := filepath.Abs(sensitiveDir)
		if err != nil {
			continue
		}
		
		// Normalize for comparison
		normalizedPath := path
		normalizedDir := absDir
		if runtime.GOOS == "windows" {
			normalizedPath = strings.ToLower(path)
			normalizedDir = strings.ToLower(absDir)
		}
		
		// Check if path is within sensitive directory
		if strings.HasPrefix(normalizedPath, normalizedDir+string(filepath.Separator)) || normalizedPath == normalizedDir {
			return fmt.Errorf("access denied: cannot write to system directory '%s'", sensitiveDir)
		}
	}
	
	// Check if trying to overwrite critical files
	for _, sensitive := range SensitivePaths {
		if matchPath(path, sensitive) {
			return fmt.Errorf("access denied: cannot write to sensitive location '%s'", path)
		}
	}
	
	return nil
}

// matchPath checks if a path matches a pattern (supports wildcards)
func matchPath(path, pattern string) bool {
	// Normalize paths
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
		pattern = strings.ToLower(pattern)
	}
	
	// Direct match
	if strings.Contains(path, pattern) {
		return true
	}
	
	// Filename match
	if strings.Contains(filepath.Base(path), strings.TrimPrefix(pattern, "*")) {
		return true
	}
	
	// Try glob match
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}
	
	return false
}

// ValidateOutputDirectory ensures the output directory is safe to write to
func ValidateOutputDirectory(path string) (string, error) {
	cleanPath, err := ValidatePath(path, "output")
	if err != nil {
		return "", err
	}
	
	// Ensure directory exists or can be created
	if err := os.MkdirAll(cleanPath, 0755); err != nil {
		return "", fmt.Errorf("cannot create output directory '%s': %w", path, err)
	}
	
	// Verify we can write to the directory
	testFile := filepath.Join(cleanPath, ".write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return "", fmt.Errorf("cannot write to output directory '%s': %w", path, err)
	}
	os.Remove(testFile) // Clean up test file
	
	return cleanPath, nil
}

// ValidateConfigPath ensures the config file path is safe to read
func ValidateConfigPath(path string) (string, error) {
	if path == "" {
		return "", nil // Empty path is OK (will use defaults)
	}
	
	cleanPath, err := ValidatePath(path, "config")
	if err != nil {
		return "", fmt.Errorf("invalid config file path: %w", err)
	}
	
	// Check if file exists
	if _, err := os.Stat(cleanPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - this is OK, will use defaults
			return cleanPath, nil
		}
		return "", fmt.Errorf("cannot access config file '%s': %w", path, err)
	}
	
	// Verify it's a regular file, not a directory or special file
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot stat config file '%s': %w", path, err)
	}
	
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("config path '%s' is not a regular file", path)
	}
	
	return cleanPath, nil
}

// GetExecutableDir returns the directory containing the executable
func GetExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}
	
	return filepath.Dir(exePath), nil
}

// GetDefaultOutputDir returns the default output directory (scans/ next to executable)
func GetDefaultOutputDir() (string, error) {
	execDir, err := GetExecutableDir()
	if err != nil {
		// Fallback to current directory if we can't determine executable location
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return "", fmt.Errorf("failed to determine output directory: %w", err)
		}
		return filepath.Join(cwd, "scans"), nil
	}
	
	return filepath.Join(execDir, "scans"), nil
}

