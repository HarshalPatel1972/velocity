package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func findDevToolsPort() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	packagesDir := filepath.Join(localAppData, "Packages")

	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return "", err
	}

	var waPackageDir string
	for _, e := range entries {
		if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), "whatsapp") {
			waPackageDir = filepath.Join(packagesDir, e.Name())
			break
		}
	}

	if waPackageDir == "" {
		return "", fmt.Errorf("whatsapp package directory not found")
	}

	var activePortPath string
	err = filepath.Walk(waPackageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "DevToolsActivePort" {
			activePortPath = path
			return fmt.Errorf("found") // Break walk early
		}
		return nil
	})

	if activePortPath == "" {
		return "", fmt.Errorf("DevToolsActivePort not found. WhatsApp may need to be restarted with --remote-debugging-port=9222")
	}

	portBytes, err := os.ReadFile(activePortPath)
	if err != nil {
		return "", fmt.Errorf("could not read port file: %w", err)
	}

	lines := strings.Split(string(portBytes), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", fmt.Errorf("port file is empty")
}

func main() {
	fmt.Println("Starting WhatsApp heap inspection...")

	port, err := findDevToolsPort()
	if err != nil {
		log.Fatalf("Error finding CDP port: %v", err)
	}
	fmt.Println("Found CDP port:", port)
}

