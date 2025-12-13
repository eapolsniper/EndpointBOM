package system

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"
)

// Info contains system information
type Info struct {
	Hostname  string
	OSName    string
	OSVersion string
	Users     []string
	LocalIPs  []string
	PublicIP  string
	Timestamp time.Time
}

// GetSystemInfo gathers system information
func GetSystemInfo(disablePublicIP bool) (*Info, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	info := &Info{
		Hostname:  hostname,
		OSName:    runtime.GOOS,
		OSVersion: getOSVersion(),
		Timestamp: time.Now(),
	}

	// Get logged in users
	users, err := getLoggedInUsers()
	if err == nil {
		info.Users = users
	}

	// Get local IP addresses
	localIPs, err := getLocalIPs()
	if err == nil {
		info.LocalIPs = localIPs
	}

	// Get public IP (non-blocking, may fail if no internet)
	// Can be disabled to avoid external service calls
	if disablePublicIP {
		info.PublicIP = "disabled"
	} else {
		publicIP, err := getPublicIP()
		if err == nil {
			info.PublicIP = publicIP
		} else {
			info.PublicIP = "unavailable"
		}
	}

	return info, nil
}

// getOSVersion returns the OS version string
func getOSVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return getMacOSVersion()
	case "windows":
		return getWindowsVersion()
	case "linux":
		return getLinuxVersion()
	default:
		return "unknown"
	}
}

func getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getWindowsVersion() string {
	cmd := exec.Command("cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getLinuxVersion() string {
	// Try to read /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}
	}
	return "unknown"
}

// getLoggedInUsers returns a list of currently logged in users
func getLoggedInUsers() ([]string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getUnixLoggedInUsers()
	case "windows":
		return getWindowsLoggedInUsers()
	default:
		return nil, fmt.Errorf("unsupported OS")
	}
}

func getUnixLoggedInUsers() ([]string, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	users := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			users[fields[0]] = true
		}
	}

	var result []string
	for u := range users {
		result = append(result, u)
	}
	return result, nil
}

func getWindowsLoggedInUsers() ([]string, error) {
	cmd := exec.Command("query", "user")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	users := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] != "" {
			users[fields[0]] = true
		}
	}

	var result []string
	for u := range users {
		result = append(result, u)
	}
	return result, nil
}

// GetAllUserProfiles returns all user home directories
func GetAllUserProfiles() ([]string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacUserProfiles()
	case "windows":
		return getWindowsUserProfiles()
	case "linux":
		return getLinuxUserProfiles()
	default:
		return nil, fmt.Errorf("unsupported OS")
	}
}

func getMacUserProfiles() ([]string, error) {
	entries, err := os.ReadDir("/Users")
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			// Exclude system users
			if entry.Name() == "Shared" || entry.Name() == "Guest" {
				continue
			}
			profiles = append(profiles, "/Users/"+entry.Name())
		}
	}
	return profiles, nil
}

func getWindowsUserProfiles() ([]string, error) {
	entries, err := os.ReadDir("C:\\Users")
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Exclude system users
			name := entry.Name()
			if name == "Public" || name == "Default" || name == "All Users" {
				continue
			}
			profiles = append(profiles, "C:\\Users\\"+name)
		}
	}
	return profiles, nil
}

func getLinuxUserProfiles() ([]string, error) {
	entries, err := os.ReadDir("/home")
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			profiles = append(profiles, "/home/"+entry.Name())
		}
	}
	return profiles, nil
}

// GetCurrentUser returns the current user information
func GetCurrentUser() (*user.User, error) {
	return user.Current()
}

// IsAdmin checks if the current process is running with admin/root privileges
func IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		return isWindowsAdmin()
	default:
		return os.Geteuid() == 0
	}
}

func isWindowsAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// getLocalIPs returns all local IP addresses (excluding loopback)
func getLocalIPs() ([]string, error) {
	var ips []string
	
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// Check if it's an IP address (not just a network interface)
		if ipNet, ok := addr.(*net.IPNet); ok {
			// Skip loopback addresses
			if ipNet.IP.IsLoopback() {
				continue
			}
			
			// Get IPv4 addresses
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
			
			// Also include IPv6 addresses (excluding link-local)
			if ipNet.IP.To16() != nil && !ipNet.IP.IsLinkLocalUnicast() && ipNet.IP.To4() == nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	return ips, nil
}

// getPublicIP attempts to determine the public IP address
// Returns error if unable to reach external service (no internet, firewall, etc.)
// SECURITY: Uses strong input validation on untrusted external services
func getPublicIP() (string, error) {
	// Try multiple services for reliability
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second, // 5 second timeout
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue // Try next service
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		// SECURITY: Limit response size to prevent memory exhaustion
		// An IP address should be max 45 chars (IPv6 with colons)
		// We allow 100 bytes to be safe, but prevent large responses
		limitedReader := io.LimitReader(resp.Body, 100)
		body, err := io.ReadAll(limitedReader)
		if err != nil {
			continue
		}
		
		ip := strings.TrimSpace(string(body))
		
		// SECURITY: Strict validation of IP address format
		if !isValidPublicIP(ip) {
			continue // Invalid IP format, try next service
		}
		
		return ip, nil
	}

	return "", fmt.Errorf("unable to determine public IP")
}

// isValidPublicIP performs strict validation on IP addresses from untrusted sources
func isValidPublicIP(ip string) bool {
	// Must be a valid IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// SECURITY: Reject private/internal IP addresses
	// Public IP services should only return public IPs
	// This prevents potential SSRF or manipulation attacks
	if parsedIP.IsLoopback() || parsedIP.IsPrivate() || parsedIP.IsLinkLocalUnicast() || 
	   parsedIP.IsLinkLocalMulticast() || parsedIP.IsMulticast() {
		return false
	}
	
	// Additional check: IP string should be reasonable length
	// IPv4: max 15 chars (xxx.xxx.xxx.xxx)
	// IPv6: max 39 chars (8 groups of 4 hex digits with colons)
	if len(ip) > 45 {
		return false
	}
	
	// Ensure the IP string doesn't contain any unexpected characters
	// Valid characters: 0-9, a-f, A-F, ., :
	for _, char := range ip {
		if !((char >= '0' && char <= '9') || 
		     (char >= 'a' && char <= 'f') || 
		     (char >= 'A' && char <= 'F') || 
		     char == '.' || char == ':') {
			return false
		}
	}
	
	return true
}

