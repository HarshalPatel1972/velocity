package cpu

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	setProcessInformation = kernel32.NewProc("SetProcessInformation")
	psapi                 = windows.NewLazySystemDLL("psapi.dll")
	getProcessMemoryInfo  = psapi.NewProc("GetProcessMemoryInfo")
)

// Windows constants for process information
const (
	ProcessPowerThrottling = 4 // PROCESS_INFORMATION_CLASS

	// Priority classes
	IDLE_PRIORITY_CLASS   = 0x00000040
	NORMAL_PRIORITY_CLASS = 0x00000020
)

// PROCESS_POWER_THROTTLING_STATE structure
type processPowerThrottlingState struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

const (
	PROCESS_POWER_THROTTLING_CURRENT_VERSION    = 1
	PROCESS_POWER_THROTTLING_EXECUTION_SPEED    = 0x1
)

// PROCESS_MEMORY_COUNTERS structure
type processMemoryCounters struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// EnforceEfficiencyMode sets the process to use minimal CPU resources.
// - Enables EcoQoS (routes to efficiency cores)
// - Sets priority to IDLE_PRIORITY_CLASS
func EnforceEfficiencyMode(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	// Enable EcoQoS (Power Throttling)
	state := processPowerThrottlingState{
		Version:     PROCESS_POWER_THROTTLING_CURRENT_VERSION,
		ControlMask: PROCESS_POWER_THROTTLING_EXECUTION_SPEED,
		StateMask:   PROCESS_POWER_THROTTLING_EXECUTION_SPEED, // 1 = Enable throttling
	}

	ret, _, err := setProcessInformation.Call(
		uintptr(handle),
		uintptr(ProcessPowerThrottling),
		uintptr(unsafe.Pointer(&state)),
		uintptr(unsafe.Sizeof(state)),
	)
	if ret == 0 {
		// EcoQoS may not be supported on older Windows versions, continue anyway
	}

	// Set priority to IDLE
	err = windows.SetPriorityClass(handle, IDLE_PRIORITY_CLASS)
	if err != nil {
		return fmt.Errorf("failed to set IDLE priority for PID %d: %w", pid, err)
	}

	return nil
}

// EnforcePerformanceMode sets the process to normal CPU resources.
// - Disables EcoQoS (allows performance cores)
// - Sets priority to NORMAL_PRIORITY_CLASS
func EnforcePerformanceMode(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	// Disable EcoQoS (Power Throttling)
	state := processPowerThrottlingState{
		Version:     PROCESS_POWER_THROTTLING_CURRENT_VERSION,
		ControlMask: PROCESS_POWER_THROTTLING_EXECUTION_SPEED,
		StateMask:   0, // 0 = Disable throttling
	}

	ret, _, err := setProcessInformation.Call(
		uintptr(handle),
		uintptr(ProcessPowerThrottling),
		uintptr(unsafe.Pointer(&state)),
		uintptr(unsafe.Sizeof(state)),
	)
	if ret == 0 {
		// EcoQoS may not be supported, continue anyway
	}

	// Set priority to NORMAL
	err = windows.SetPriorityClass(handle, NORMAL_PRIORITY_CLASS)
	if err != nil {
		return fmt.Errorf("failed to set NORMAL priority for PID %d: %w", pid, err)
	}

	return nil
}

// GetMemoryUsageMB returns the current working set size of the process in MB.
func GetMemoryUsageMB(pid uint32) (uint64, error) {
	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	var memInfo processMemoryCounters
	memInfo.cb = uint32(unsafe.Sizeof(memInfo))

	ret, _, err := getProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&memInfo)),
		uintptr(memInfo.cb),
	)
	if ret == 0 {
		return 0, fmt.Errorf("GetProcessMemoryInfo failed for PID %d: %w", pid, err)
	}

	// Convert bytes to MB
	return uint64(memInfo.WorkingSetSize) / (1024 * 1024), nil
}

// GetTotalMemoryUsageMB returns the combined memory usage of multiple processes.
func GetTotalMemoryUsageMB(pids []uint32) uint64 {
	var total uint64
	for _, pid := range pids {
		mb, err := GetMemoryUsageMB(pid)
		if err == nil {
			total += mb
		}
	}
	return total
}
