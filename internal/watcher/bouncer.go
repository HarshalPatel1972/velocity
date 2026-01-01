package watcher

import (
	"log"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	setWinEventHook           = user32.NewProc("SetWinEventHook")
	unhookWinEvent            = user32.NewProc("UnhookWinEvent")
	getWindowTextW            = user32.NewProc("GetWindowTextW")
	getWindowTextLengthW      = user32.NewProc("GetWindowTextLengthW")
	getAsyncKeyState          = user32.NewProc("GetAsyncKeyState")
	getCursorPos              = user32.NewProc("GetCursorPos")
	getWindowRect             = user32.NewProc("GetWindowRect")
	setForegroundWindow       = user32.NewProc("SetForegroundWindow")
	flashWindow               = user32.NewProc("FlashWindow")
	getWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	getMessage                = user32.NewProc("GetMessageW")
	translateMessage          = user32.NewProc("TranslateMessage")
	dispatchMessage           = user32.NewProc("DispatchMessageW")
)

const (
	EVENT_SYSTEM_FOREGROUND = 0x0003
	WINEVENT_OUTOFCONTEXT   = 0x0000
	VK_MENU                 = 0x12 // Alt key
)

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// Watcher manages focus steal prevention
type Watcher struct {
	mu             sync.RWMutex
	whatsappPIDs   map[uint32]bool
	lastGoodWindow uintptr
	running        bool
	hookHandle     uintptr
}

var globalWatcher *Watcher

// NewWatcher creates a new focus watcher
func NewWatcher() *Watcher {
	w := &Watcher{
		whatsappPIDs: make(map[uint32]bool),
	}
	globalWatcher = w
	return w
}

// UpdatePIDs updates the list of WhatsApp process IDs
func (w *Watcher) UpdatePIDs(pids []uint32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.whatsappPIDs = make(map[uint32]bool)
	for _, pid := range pids {
		w.whatsappPIDs[pid] = true
	}
}

// Start begins watching for focus steals
func (w *Watcher) Start() {
	if w.running {
		return
	}
	w.running = true

	go func() {
		runtime.LockOSThread() // Required for Windows hooks
		defer runtime.UnlockOSThread()

		// Set up the hook
		callback := syscall.NewCallback(winEventProc)
		hook, _, _ := setWinEventHook.Call(
			EVENT_SYSTEM_FOREGROUND,
			EVENT_SYSTEM_FOREGROUND,
			0,
			callback,
			0,
			0,
			WINEVENT_OUTOFCONTEXT,
		)

		if hook == 0 {
			log.Println("[BOUNCER] Failed to set hook")
			return
		}

		w.hookHandle = hook
		log.Println("[BOUNCER] Focus watcher started")

		// Message pump
		var msg MSG
		for {
			ret, _, _ := getMessage.Call(
				uintptr(unsafe.Pointer(&msg)),
				0, 0, 0,
			)
			if ret == 0 || ret == uintptr(0xFFFFFFFF) {
				break
			}
			translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}

		// Cleanup
		unhookWinEvent.Call(hook)
	}()
}

// winEventProc is the callback for focus changes
func winEventProc(hWinEventHook, event, hwnd, idObject, idChild uintptr, dwEventThread, dwmsEventTime uint32) uintptr {
	if globalWatcher == nil || hwnd == 0 {
		return 0
	}

	// Get PID of the new foreground window
	var pid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	globalWatcher.mu.RLock()
	isWhatsApp := globalWatcher.whatsappPIDs[pid]
	lastGood := globalWatcher.lastGoodWindow
	globalWatcher.mu.RUnlock()

	if !isWhatsApp {
		// Not WhatsApp - update lastGoodWindow
		globalWatcher.mu.Lock()
		globalWatcher.lastGoodWindow = hwnd
		globalWatcher.mu.Unlock()
		return 0
	}

	// It's WhatsApp - check safety filters

	// Safety Filter 1: Check window title for calls
	title := getWindowTitle(hwnd)
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, "call") || strings.Contains(titleLower, "video") {
		log.Printf("[BOUNCER] Allowed: Incoming call/video (%s)", title)
		return 0
	}

	// Safety Filter 2: Check if Alt key is pressed (Alt-Tab)
	if isAltPressed() {
		log.Println("[BOUNCER] Allowed: Alt-Tab detected")
		return 0
	}

	// Safety Filter 3: Check if mouse is inside the window (intentional click)
	if isMouseInsideWindow(hwnd) {
		log.Println("[BOUNCER] Allowed: Mouse click detected")
		return 0
	}

	// BLOCK: This is a focus steal!
	if lastGood != 0 {
		log.Printf("[BOUNCER] BLOCKED focus steal! Reverting to previous window")
		
		// Revert focus
		setForegroundWindow.Call(lastGood)
		
		// Flash the WhatsApp window to alert user
		flashWindow.Call(hwnd, 1)
	}

	return 0
}

// getWindowTitle retrieves the window title
func getWindowTitle(hwnd uintptr) string {
	length, _, _ := getWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}

	buf := make([]uint16, length+1)
	getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	return windows.UTF16ToString(buf)
}

// isAltPressed checks if the Alt key is currently pressed
func isAltPressed() bool {
	state, _, _ := getAsyncKeyState.Call(VK_MENU)
	return (state & 0x8000) != 0
}

// isMouseInsideWindow checks if the mouse cursor is inside the window
func isMouseInsideWindow(hwnd uintptr) bool {
	var cursor POINT
	getCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))

	var rect RECT
	getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

	return cursor.X >= rect.Left && cursor.X <= rect.Right &&
		cursor.Y >= rect.Top && cursor.Y <= rect.Bottom
}
