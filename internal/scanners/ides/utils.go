package ides

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// getUserProfiles returns user home directories
func getUserProfiles() ([]string, error) {
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
		if entry.IsDir() && entry.Name() != "Shared" && entry.Name() != "Guest" {
			profiles = append(profiles, filepath.Join("/Users", entry.Name()))
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
			name := entry.Name()
			if name != "Public" && name != "Default" && name != "All Users" {
				profiles = append(profiles, filepath.Join("C:\\Users", name))
			}
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
			profiles = append(profiles, filepath.Join("/home", entry.Name()))
		}
	}
	return profiles, nil
}

