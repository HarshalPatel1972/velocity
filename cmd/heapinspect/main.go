package main

import (
	"fmt"
	"log"
	"os"

	"velocity/internal/cdp"
)

func generateReport() {
	fmt.Printf("[*] Generating heap_report.txt...\n")
	file, err := os.Create("heap_report.txt")
	if err != nil {
		fmt.Printf("[-] Failed to create report file: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("==========================================\n")
	file.WriteString("       Velocity CDP Heap Report\n")
	file.WriteString("==========================================\n\n")

	file.WriteString("--- Memory Counters ---\n")
	file.WriteString(fmt.Sprintf("%v\n\n", domCountersData))

	file.WriteString("--- Store/React Metrics ---\n")
	if reactFiberData != "" {
		file.WriteString(fmt.Sprintf("Store: %v\n", reactFiberData))
	} else {
		file.WriteString("Store: Not available (or not evaluated)\n")
	}
	file.WriteString("\n")

	file.WriteString("--- Largest 10 Objects (by retained size) ---\n")
	file.WriteString(fmt.Sprintf("%v\n\n", largeObjectsData))

	file.WriteString("==========================================\n")
	fmt.Printf("[+] Report saved to heap_report.txt\n")
}

func main() {
	fmt.Println("Starting WhatsApp heap inspection...")

	port, err := cdp.FindDevToolsPort()
	if err != nil {
		fmt.Printf("Error finding CDP port dynamically: %v\nFalling back to 9222...\n", err)
		port = "9222"
	}
	fmt.Println("Using CDP port:", port)

	client, err := cdp.Connect(port)
	if err != nil {
		log.Fatalf("Error connecting to CDP: %v", err)
	}
	defer client.Close()

	fmt.Println("Successfully connected to CDP WebSocket")

	runHeapsQueries(client)
	takeHeapSnapshot(client)

	generateReport()
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

func takeHeapSnapshot(conn *websocket.Conn) {
	fmt.Println("Taking full heap snapshot... This will take a while.")

	// Enable HeapProfiler
	conn.WriteJSON(map[string]interface{}{
		"id":     1,
		"method": "HeapProfiler.enable",
	})

	// Start snapshot
	conn.WriteJSON(map[string]interface{}{
		"id":     2,
		"method": "HeapProfiler.takeHeapSnapshot",
		"params": map[string]interface{}{
			"reportProgress":      true,
			"captureNumericValue": true,
		},
	})

	file, err := os.Create("whatsapp_heap.json")
	if err != nil {
		log.Fatalf("Could not create heap file: %v", err)
	}
	defer file.Close()

	for {
		var resp map[string]interface{}
		if err := conn.ReadJSON(&resp); err != nil {
			break
		}

		if method, ok := resp["method"].(string); ok {
			if method == "HeapProfiler.addHeapSnapshotChunk" {
				params := resp["params"].(map[string]interface{})
				chunk := params["chunk"].(string)
				file.WriteString(chunk)
			}
			if method == "HeapProfiler.reportHeapSnapshotProgress" {
				// Just printing progress might be too noisy. Keep silent.
			}
		}

		if id, ok := resp["id"].(float64); ok && int(id) == 2 {
			fmt.Println("Snapshot complete. Saved to whatsapp_heap.json")
			break
		}
	}
}

