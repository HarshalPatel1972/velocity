package memory

import (
	"sync"
	"time"
	"unsafe"

	"velocity/internal/utils"

	"golang.org/x/sys/windows"
)

var (
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procSetProcessInformation = kernel32.NewProc("SetProcessInformation")
)

const (
	ProcessMemoryPriority = 0

	MEMORY_PRIORITY_BELOW_NORMAL = 4
	MEMORY_PRIORITY_NORMAL       = 5
)

type MEMORY_PRIORITY_INFORMATION struct {
	MemoryPriority uint32
}

var (
	currentPriority uint32 = MEMORY_PRIORITY_NORMAL
	priorityMutex   sync.RWMutex
)

func init() {
	// Continuously re-apply memory priority to ensure new child processes get the right priority
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			applyCurrentPriority()
		}
	}()
}

func applyCurrentPriority() {
	priorityMutex.RLock()
	priority := currentPriority
	priorityMutex.RUnlock()

	pids, err := utils.GetWhatsAppProcessTree()
	if err != nil || len(pids) == 0 {
		return
	}

	for _, pid := range pids {
		setProcessMemoryPriority(pid, priority)
	}
}

// SetForegroundPriority restores process memory priority to normal.
func SetForegroundPriority() {
	priorityMutex.Lock()
	currentPriority = MEMORY_PRIORITY_NORMAL
	priorityMutex.Unlock()
	applyCurrentPriority()
}

// SetBackgroundPriority demotes process memory priority to below normal.
func SetBackgroundPriority() {
	priorityMutex.Lock()
	currentPriority = MEMORY_PRIORITY_BELOW_NORMAL
	priorityMutex.Unlock()
	applyCurrentPriority()
}

func setProcessMemoryPriority(pid uint32, priority uint32) {
	// Required process access rights: PROCESS_SET_INFORMATION | PROCESS_QUERY_LIMITED_INFORMATION
	handle, err := windows.OpenProcess(windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return
	}
	defer windows.CloseHandle(handle)

	memPriInfo := MEMORY_PRIORITY_INFORMATION{
		MemoryPriority: priority,
	}

	procSetProcessInformation.Call(
		uintptr(handle),
		ProcessMemoryPriority,
		uintptr(unsafe.Pointer(&memPriInfo)),
		unsafe.Sizeof(memPriInfo),
	)
}
