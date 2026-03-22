package memory

import (
	"log"
	"time"

	"velocity/internal/cdp"
	"velocity/internal/cpu"
	"velocity/internal/utils"
)

var (
	sharedClient *cdp.Client
	isConnecting bool
)

// AttemptCDPConnection tries to connect with exponential backoff
func AttemptCDPConnectionAsync() {
	if sharedClient != nil || isConnecting {
		return
	}

	isConnecting = true
	go func() {
		backoffs := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
		attempt := 0

		for {
			// Fast exit if process isn't running
			pids, _ := utils.GetWhatsAppProcessTree()
			if len(pids) == 0 {
				isConnecting = false
				return
			}

			err := InitCDPClient()
			if err == nil {
				log.Println("[CDP] Successfully connected to DevTools endpoint")
				isConnecting = false
				return
			}

			var wait time.Duration
			if attempt < len(backoffs) {
				wait = backoffs[attempt]
			} else {
				wait = 30 * time.Second
			}
			
			log.Printf("[CDP] Connection failed, retrying in %v...", wait)
			time.Sleep(wait)
			attempt++
		}
	}()
}

// InitCDPClient initialized or reconnects the persistent CDP client
func InitCDPClient() error {
	if sharedClient != nil {
		return nil
	}

	port, err := cdp.FindDevToolsPort()
	if err != nil {
		return err
	}

	client, err := cdp.Connect(port)
	if err != nil {
		return err
	}

	sharedClient = client
	return nil
}

// PurgeNativeCache executes the Phase 2 CDP purge sequence to release native Chromium buffers
func PurgeNativeCache() error {
	if sharedClient == nil {
		return nil // Client not connected, ignore
	}

	pids, _ := utils.GetWhatsAppProcessTree()
	beforeMB := cpu.GetTotalMemoryUsageMB(pids)
	
	loc, _ := time.LoadLocation("Local")
	log.Printf("[PURGE] %s — before: %d MB", time.Now().In(loc).Format("15:04:05"), beforeMB)
	
	start := time.Now()

	// Phase 1: Clear Network Caches
	// 1. Clear Browser Cache explicitly
	_, err := sharedClient.Send("Network.clearBrowserCache", nil)
	if err != nil {
		return handleCDPError(err)
	}

	// 2. Enable Network domain, toggle cache to force flush, then restore and disable.
	sharedClient.Send("Network.enable", nil)
	sharedClient.Send("Network.setCacheDisabled", map[string]interface{}{"cacheDisabled": true})
	sharedClient.Send("Network.setCacheDisabled", map[string]interface{}{"cacheDisabled": false})
	sharedClient.Send("Network.disable", nil)

	// Phase 2: Release Decoded Image & Native Rendering Buffers
	sharedClient.Send("DOM.enable", nil)
	sharedClient.Send("CSS.enable", nil)
	sharedClient.Send("CSS.forcePseudoState", map[string]interface{}{
		"nodeId":              1,
		"forcedPseudoClasses": []string{},
	})

	// Phase 3: Purge JS Memory directly
	sharedClient.Send("Memory.forciblyPurgeJavaScriptMemory", nil)

	// Release DOM/CSS snapshot buffers safely
	sharedClient.Send("DOM.disable", nil)
	sharedClient.Send("CSS.disable", nil)

	// Phase 4: Final V8 Garbage Collection Sweep
	sharedClient.Send("HeapProfiler.enable", nil)
	sharedClient.Send("HeapProfiler.collectGarbage", nil)
	sharedClient.Send("HeapProfiler.disable", nil)

	// Let GC settle
	time.Sleep(500 * time.Millisecond)

	duration := time.Since(start)
	
	pidsAfter, _ := utils.GetWhatsAppProcessTree()
	afterMB := cpu.GetTotalMemoryUsageMB(pidsAfter)
	delta := int64(afterMB) - int64(beforeMB)
	
	log.Printf("[PURGE] %s — after:  %d MB  (%d MB)", time.Now().In(loc).Format("15:04:05"), afterMB, delta)
	log.Printf("[PURGE] %s — duration: %.1fs", time.Now().In(loc).Format("15:04:05"), duration.Seconds())

	return nil
}

func handleCDPError(err error) error {
	log.Printf("[PURGE] CDP Error: %v\n", err)
	// If the connection drops, clear the shared client so it can be reinitialized
	if sharedClient != nil {
		sharedClient.Close()
		sharedClient = nil
	}
	
	AttemptCDPConnectionAsync()
	
	return err
}
