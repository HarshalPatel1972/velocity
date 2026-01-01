package window

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	getForegroundWindow       = user32.NewProc("GetForegroundWindow")
	getWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
)

// IsAnyProcessFocused checks if any of the given PIDs owns the foreground window.
func IsAnyProcessFocused(pids []uint32) bool {
	// Get the foreground window handle
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}

	// Get the PID of the foreground window
	var foregroundPID uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&foregroundPID)))

	// Check if any of our target PIDs match
	for _, pid := range pids {
		if pid == foregroundPID {
			return true
		}
	}

	return false
}
