package packagemanagers

import (
	"os/exec"
	"strings"
)

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}

// splitAtLastChar splits a string at the last occurrence of a character
func splitAtLastChar(s string, sep rune) []string {
	idx := strings.LastIndexFunc(s, func(r rune) bool {
		return r == sep
	})
	
	if idx == -1 {
		return []string{s}
	}
	
	return []string{s[:idx], s[idx+1:]}
}

