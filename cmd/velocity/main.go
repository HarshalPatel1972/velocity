package main

import (
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"

	"velocity/internal/cpu"
	"velocity/internal/memory"
	"velocity/internal/tray"
	"velocity/internal/updater"
	"velocity/internal/utils"
	"velocity/internal/watcher"
	"velocity/internal/window"
)

const (
	CurrentVersion     = "v1.0.2"
	rootProcess        = "WhatsApp.Root.exe"
	memoryTrimInterval = 30 * time.Second
	qosCheckInterval   = 500 * time.Millisecond
	maxMemoryMB        = 2048
)

type AppState string

const (
	StateUnknown    AppState = "unknown"
	StateFocused    AppState = "focused"
	StateBackground AppState = "background"
)

var (
	forceTrimChan  = make(chan struct{})
	logFile        *os.File
	focusWatcher   *watcher.Watcher
)

func main() {
	// Setup logging to file
	var err error
	logFile, err = os.OpenFile("velocity.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime)

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(tray.Icon)
	systray.SetTitle("Velocity")
	systray.SetTooltip("WhatsApp Memory & CPU Optimizer")

	log.Println("Velocity " + CurrentVersion + " - With Auto-Updater")
	log.Printf("Monitoring: %s | Trim: %v | QoS: %v\n", rootProcess, memoryTrimInterval, qosCheckInterval)

	// Silent update check on startup
	go checkForUpdatesAsync(false)

	// Initialize focus watcher
	focusWatcher = watcher.NewWatcher()
	focusWatcher.Start()

	// Menu items
	mStatus := systray.AddMenuItem("Status: Active", "Velocity is running")
	mStatus.Disable()

	systray.AddSeparator()

	mTrim := systray.AddMenuItem("Force Trim Now", "Immediately trim WhatsApp memory")
	mUpdate := systray.AddMenuItem("Check for Updates", "Check for new version")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit Velocity")

	// Start background worker
	go backgroundLoop()

	// Handle menu events
	go func() {
		for {
			select {
			case <-mTrim.ClickedCh:
				log.Println("[USER] Force Trim requested")
				forceTrimChan <- struct{}{}
			case <-mUpdate.ClickedCh:
				log.Println("[USER] Check for Updates requested")
				go checkForUpdatesAsync(true)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Println("Velocity shutting down...")
	if logFile != nil {
		logFile.Close()
	}
}

func backgroundLoop() {
	memoryTicker := time.NewTicker(memoryTrimInterval)
	qosTicker := time.NewTicker(qosCheckInterval)
	defer memoryTicker.Stop()
	defer qosTicker.Stop()

	lastState := StateUnknown

	// Initial trim
	trimWhatsApp()

	for {
		select {
		case <-memoryTicker.C:
			trimWhatsApp()

		case <-forceTrimChan:
			trimWhatsApp()

		case <-qosTicker.C:
			pids, err := utils.FindProcessTree(rootProcess)
			if err != nil {
				if lastState != StateUnknown {
					lastState = StateUnknown
				}
				continue
			}

			// Update watcher with current PIDs
			if focusWatcher != nil {
				focusWatcher.UpdatePIDs(pids)
			}

			var currentState AppState
			if window.IsAnyProcessFocused(pids) {
				currentState = StateFocused
			} else {
				currentState = StateBackground
			}

			if currentState != lastState {
				switch currentState {
				case StateFocused:
					totalMB := cpu.GetTotalMemoryUsageMB(pids)
					if totalMB < maxMemoryMB {
						applyPerformanceMode(pids)
						log.Printf("[QoS] Focused → Performance Mode (RAM: %d MB)", totalMB)
					} else {
						log.Printf("[QoS] Focused but RAM high (%d MB) → Staying Normal", totalMB)
					}
				case StateBackground:
					applyEfficiencyMode(pids)
					log.Printf("[QoS] Background → Efficiency Mode")
				}
				lastState = currentState
			}
		}
	}
}

func trimWhatsApp() {
	pids, err := utils.FindProcessTree(rootProcess)
	if err != nil {
		if err == utils.ErrProcessNotFound {
			log.Printf("[INFO] %s not running", rootProcess)
		} else {
			log.Printf("[ERROR] %v", err)
		}
		return
	}

	trimmed := 0
	for _, pid := range pids {
		if memory.TrimProcess(pid) == nil {
			trimmed++
		}
	}
	log.Printf("[MEM] Trimmed %d/%d processes", trimmed, len(pids))
}

func applyEfficiencyMode(pids []uint32) {
	for _, pid := range pids {
		cpu.EnforceEfficiencyMode(pid)
	}
}

func applyPerformanceMode(pids []uint32) {
	for _, pid := range pids {
		cpu.EnforcePerformanceMode(pid)
	}
}

func checkForUpdatesAsync(showNotification bool) {
	info, err := updater.CheckForUpdates(CurrentVersion)
	if err != nil {
		log.Printf("[UPDATE] Error checking for updates: %v", err)
		return
	}

	if !info.Available {
		if showNotification {
			log.Println("[UPDATE] You are running the latest version")
		}
		return
	}

	log.Printf("[UPDATE] New version available: %s", info.NewVersion)

	// Download the installer
	log.Printf("[UPDATE] Downloading %s...", info.InstallerName)
	installerPath, err := updater.DownloadInstaller(info.DownloadURL, info.InstallerName)
	if err != nil {
		log.Printf("[UPDATE] Download failed: %v", err)
		return
	}

	log.Printf("[UPDATE] Downloaded to: %s", installerPath)
	log.Println("[UPDATE] Launching installer...")

	// Launch installer and exit
	if err := updater.LaunchInstallerAndExit(installerPath); err != nil {
		log.Printf("[UPDATE] Failed to launch installer: %v", err)
	}
}

