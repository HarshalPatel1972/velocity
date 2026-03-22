package cdp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Target struct {
	Id                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	Url                  string `json:"url"`
	WebSocketDebuggerUrl string `json:"webSocketDebuggerUrl"`
}

type Client struct {
	conn  *websocket.Conn
	msgID int
	mu    sync.Mutex
}

func FindDevToolsPort() (string, error) {
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
		return "", fmt.Errorf("DevToolsActivePort not found")
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

func Connect(port string) (*Client, error) {
	// Let's use 9222 temporarily if dynamic discovery is down, or whatever is passed.
	url := fmt.Sprintf("http://localhost:%s/json/list", port)
	
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch targets: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read targets: %w", err)
	}

	var targets []Target
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, fmt.Errorf("failed to parse targets JSON: %w", err)
	}

	var pageWsUrl string
	for _, t := range targets {
		if t.Type == "page" && strings.Contains(t.Url, "web.whatsapp.com") {
			pageWsUrl = t.WebSocketDebuggerUrl
			break
		}
	}

	if pageWsUrl == "" {
		return nil, fmt.Errorf("could not find page target for WhatsApp Web")
	}

	conn, _, err := websocket.DefaultDialer.Dial(pageWsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}

	return &Client{
		conn:  conn,
		msgID: 1,
	}, nil
}

func (c *Client) Send(method string, params map[string]interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	id := c.msgID
	c.msgID++
	c.mu.Unlock()

	cmd := map[string]interface{}{
		"id":     id,
		"method": method,
	}
	if params != nil {
		cmd["params"] = params
	}

	if err := c.conn.WriteJSON(cmd); err != nil {
		return nil, err
	}

	// Warning: Concurrent reads from a single WebSocket should be handled locally if 
	// there are multiple listeners. Here we assume sequential operations for simplicity.
	// Production may need a read loop routing to channels, but for purge/inspect this is enough.
	for {
		var resp map[string]interface{}
		if err := c.conn.ReadJSON(&resp); err != nil {
			return nil, err
		}
		
		if rid, ok := resp["id"].(float64); ok && int(rid) == id {
			return resp, nil
		}
		// If we receive an event rather than a response, we just continue reading.
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Expose raw connection for HeapInspect chunk reading logic
func (c *Client) Conn() *websocket.Conn {
	return c.conn
}
