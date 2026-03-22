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

	// We'll store results here later
	_ = swWsUrl

	// Connect to CDP
	conn, _, err := websocket.DefaultDialer.Dial(pageWsUrl, nil)
	if err != nil {
		log.Fatalf("Error connecting to CDP: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to CDP WebSocket")

	runHeapsQueries(conn)
}

func sendCommand(conn *websocket.Conn, id int, method string, params map[string]interface{}) map[string]interface{} {
	cmd := map[string]interface{}{
		"id":     id,
		"method": method,
	}
	if params != nil {
		cmd["params"] = params
	}
	conn.WriteJSON(cmd)

	for {
		var resp map[string]interface{}
		if err := conn.ReadJSON(&resp); err != nil {
			return nil
		}
		if rid, ok := resp["id"].(float64); ok && int(rid) == id {
			return resp
		}
	}
}

var (
	largeObjectsData string
	reactFiberData   string
	waStoreData      string
	moduleSystemData string
	webpackChunkData string
	domCountersData  map[string]interface{}
)

func runHeapsQueries(conn *websocket.Conn) {
	fmt.Println("Running heap queries...")

	// 4a
	resp := sendCommand(conn, 3, "Runtime.evaluate", map[string]interface{}{
		"expression":    "(() => { const results = []; const keys = Object.keys(window); keys.forEach(k => { try { const s = JSON.stringify(window[k]); if(s && s.length > 100000) results.push({key: k, size: s.length}); } catch(e){} }); return JSON.stringify(results); })()",
		"returnByValue": true,
	})
	if res, ok := resp["result"].(map[string]interface{})["result"].(map[string]interface{}); ok {
		largeObjectsData = fmt.Sprintf("%v", res["value"])
	}

	// 4b
	resp = sendCommand(conn, 4, "Runtime.evaluate", map[string]interface{}{
		"expression":    "(() => { const roots = []; document.querySelectorAll('*').forEach(el => { const keys = Object.keys(el).filter(k => k.startsWith('__reactFiber') || k.startsWith('__reactInternalInstance')); if(keys.length) roots.push({tag: el.tagName, id: el.id, key: keys[0]}); }); return JSON.stringify(roots.slice(0,20)); })()",
		"returnByValue": true,
	})
	if res, ok := resp["result"].(map[string]interface{})["result"].(map[string]interface{}); ok {
		reactFiberData = fmt.Sprintf("%v", res["value"])
	}

	// 4c
	resp = sendCommand(conn, 5, "Runtime.evaluate", map[string]interface{}{
		"expression":    "(() => { const wa = window.WA || window.Store || window.whatsapp || window.require; return typeof wa + ' : ' + JSON.stringify(Object.keys(wa || {}).slice(0, 50)); })()",
		"returnByValue": true,
	})
	if res, ok := resp["result"].(map[string]interface{})["result"].(map[string]interface{}); ok {
		waStoreData = fmt.Sprintf("%v", res["value"])
	}

	// 4d
	resp = sendCommand(conn, 6, "Runtime.evaluate", map[string]interface{}{
		"expression":    "typeof window.require + ' | ' + typeof window.webpackChunk_whatsapp_web_client + ' | ' + typeof window.__webpack_modules__",
		"returnByValue": true,
	})
	if res, ok := resp["result"].(map[string]interface{})["result"].(map[string]interface{}); ok {
		moduleSystemData = fmt.Sprintf("%v", res["value"])
	}

	// 4e
	resp = sendCommand(conn, 7, "Runtime.evaluate", map[string]interface{}{
		"expression":    "(() => { try { const chunks = window.webpackChunk_whatsapp_web_client; if(!chunks) return 'no chunks'; return 'chunk count: ' + chunks.length; } catch(e) { return e.toString(); } })()",
		"returnByValue": true,
	})
	if res, ok := resp["result"].(map[string]interface{})["result"].(map[string]interface{}); ok {
		webpackChunkData = fmt.Sprintf("%v", res["value"])
	}

	// DOM counters
	resp = sendCommand(conn, 9, "Memory.getDOMCounters", nil)
	if res, ok := resp["result"].(map[string]interface{}); ok {
		domCountersData = res
	}

	fmt.Println("Queries completed.")
}

