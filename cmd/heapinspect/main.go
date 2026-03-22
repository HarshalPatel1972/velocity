package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

type Target struct {
	Id                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	Url                  string `json:"url"`
	WebSocketDebuggerUrl string `json:"webSocketDebuggerUrl"`
}

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

	// Fetch targets
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/json/list", port))
	if err != nil {
		log.Fatalf("Error fetching targets: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading targets response: %v", err)
	}

	var targets []Target
	if err := json.Unmarshal(body, &targets); err != nil {
		log.Fatalf("Error parsing targets JSON: %v", err)
	}

	var pageWsUrl string
	var swWsUrl string

	fmt.Println("Targets found:")
	for _, t := range targets {
		fmt.Printf("- [%s] %s\n", t.Type, t.Url)
		if t.Type == "page" && strings.Contains(t.Url, "web.whatsapp.com") {
			pageWsUrl = t.WebSocketDebuggerUrl
		}
		if t.Type == "service_worker" {
			swWsUrl = t.WebSocketDebuggerUrl
		}
	}

	if pageWsUrl == "" {
		log.Fatalf("Could not find page target for WhatsApp Web")
	}

	fmt.Println("Page WS URL:", pageWsUrl)
	fmt.Println("Service Worker WS URL:", swWsUrl)

	// Connect to CDP
	conn, _, err := websocket.DefaultDialer.Dial(pageWsUrl, nil)
	if err != nil {
		log.Fatalf("Error connecting to CDP: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to CDP WebSocket")
}

