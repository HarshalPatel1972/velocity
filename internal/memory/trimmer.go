package memory

import (
	"fmt"
	"sync"
	"velocity/internal/utils"

	"golang.org/x/sys/windows"
)

var (
	kernel32Trim               = windows.NewLazySystemDLL("kernel32.dll")
	setProcessWorkingSetSizeEx = kernel32Trim.NewProc("SetProcessWorkingSetSizeEx")
)

const (
	QUOTA_LIMITS_HARDWS_MIN_DISABLE = 0x00000002
	QUOTA_LIMITS_HARDWS_MAX_DISABLE = 0x00000008
)

var trimMutex sync.Mutex

// TrimAllProcesses reads PIDs from the shared process_tree cache
// and applies SetProcessWorkingSetSizeEx(-1, -1) to all PIDs to empty the working set.
// This is to be called only once during a foreground-to-background state transition.
func TrimAllProcesses() error {
	trimMutex.Lock()
	defer trimMutex.Unlock()

	pids, err := utils.GetWhatsAppProcessTree()
	if err != nil || len(pids) == 0 {
		return fmt.Errorf("no processes to trim: %w", err)
	}

	for _, pid := range pids {
		_ = TrimProcess(pid)
	}

	return nil
}

// TrimProcess reduces the working set (RAM usage) of the specified process.
func TrimProcess(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	ret, _, err := setProcessWorkingSetSizeEx.Call(
		uintptr(handle),
		^uintptr(0),
		^uintptr(0),
		0,
	)
	if ret == 0 {
		return fmt.Errorf("SetProcessWorkingSetSizeEx failed for %d: %w", pid, err)
	}

	return nil
}

