package memory

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var (
	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	setProcessWorkingSetSizeEx = kernel32.NewProc("SetProcessWorkingSetSizeEx")
)

const (
	QUOTA_LIMITS_HARDWS_MIN_DISABLE = 0x00000002
	QUOTA_LIMITS_HARDWS_MAX_DISABLE = 0x00000008
)

// TrimProcess reduces the working set (RAM usage) of the specified process.
// It uses SetProcessWorkingSetSizeEx with -1,-1 to empty the working set.
func TrimProcess(pid uint32) error {
	// Open the process with required permissions
	// PROCESS_SET_QUOTA is needed for SetProcessWorkingSetSizeEx
	// PROCESS_QUERY_LIMITED_INFORMATION works better with UWP apps
	handle, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	// Call SetProcessWorkingSetSizeEx with -1, -1 to empty the working set
	// This is equivalent to EmptyWorkingSet but more compatible
	ret, _, err := setProcessWorkingSetSizeEx.Call(
		uintptr(handle),
		^uintptr(0), // SIZE_T(-1) - minimum working set size
		^uintptr(0), // SIZE_T(-1) - maximum working set size  
		0,           // Flags
	)
	if ret == 0 {
		return fmt.Errorf("SetProcessWorkingSetSizeEx failed for process %d: %w", pid, err)
	}

	return nil
}
