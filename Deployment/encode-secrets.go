package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

const xorKey = 0x5A

func xorEncode(input []byte) []byte {
	output := make([]byte, len(input))
	for i, b := range input {
		output[i] = b ^ xorKey
	}
	return output
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Secret Encoder for Dependency-Track Upload Tool")
		fmt.Println("=================================================")
		fmt.Println("\nUsage: go run encode-secrets.go <URL> <API_KEY>")
		fmt.Println("\nExample:")
		fmt.Println("  go run encode-secrets.go https://dtrack.company.com odt_AbCdEfGh123456789")
		fmt.Println("\nThis will output encoded values to use in your build command.")
		os.Exit(1)
	}

	url := os.Args[1]
	apiKey := os.Args[2]

	encodedURL := base64.StdEncoding.EncodeToString([]byte(url))
	xorEncoded := xorEncode([]byte(apiKey))
	encodedKey := base64.StdEncoding.EncodeToString(xorEncoded)

	fmt.Println("=" + "=")
	fmt.Println("Build Variables")
	fmt.Println("=" + "=")
	fmt.Println()
	fmt.Println("Copy these values into your build command:")
	fmt.Println()
	fmt.Printf("URL Value: %s\n", encodedURL)
	fmt.Printf("Key Value: %s\n", encodedKey)
	fmt.Println()
	fmt.Println("=" + "=")
	fmt.Println("Build Commands")
	fmt.Println("=" + "=")
	fmt.Println()
	fmt.Println("For Linux (amd64):")
	fmt.Printf("GOOS=linux GOARCH=amd64 go build -ldflags \"-X main.encodedURL=%s -X main.encodedKey=%s\" -o dt-upload-linux-amd64 upload-to-dependency-track.go\n", encodedURL, encodedKey)
	fmt.Println()
	fmt.Println("For macOS (Apple Silicon):")
	fmt.Printf("GOOS=darwin GOARCH=arm64 go build -ldflags \"-X main.encodedURL=%s -X main.encodedKey=%s\" -o dt-upload-macos-arm64 upload-to-dependency-track.go\n", encodedURL, encodedKey)
	fmt.Println()
	fmt.Println("For macOS (Intel):")
	fmt.Printf("GOOS=darwin GOARCH=amd64 go build -ldflags \"-X main.encodedURL=%s -X main.encodedKey=%s\" -o dt-upload-macos-amd64 upload-to-dependency-track.go\n", encodedURL, encodedKey)
	fmt.Println()
	fmt.Println("For Windows (amd64):")
	fmt.Printf("GOOS=windows GOARCH=amd64 go build -ldflags \"-X main.encodedURL=%s -X main.encodedKey=%s\" -o dt-upload-windows-amd64.exe upload-to-dependency-track.go\n", encodedURL, encodedKey)
	fmt.Println()
	fmt.Println("=" + "=")
	fmt.Println("Verification")
	fmt.Println("=" + "=")
	fmt.Println()
	fmt.Println("After building, distribute the binaries to your team.")
	fmt.Println()
}

