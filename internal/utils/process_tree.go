package utils

import (
	"errors"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	cachedPIDs []uint32
	cacheMutex sync.RWMutex
	treeErr    error
)

func init() {
	// Start the 5-second ticker to cache the process tree
	go func() {
		updateCache() // Initial update
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateCache()
		}
	}()
}

func updateCache() {
	pids, err := buildWhatsAppProcessTree()
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cachedPIDs = pids
	treeErr = err
}

// GetWhatsAppProcessTree returns all PIDs in the WhatsApp process tree.
// Returns an empty slice if WhatsApp is not running.
func GetWhatsAppProcessTree() ([]uint32, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if len(cachedPIDs) == 0 {
		return nil, treeErr
	}

	// Return a copy to avoid data races with the caller
	res := make([]uint32, len(cachedPIDs))
	copy(res, cachedPIDs)
	return res, treeErr
}

func buildWhatsAppProcessTree() ([]uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		return nil, err
	}

	parentMap := make(map[uint32][]uint32)
	var rootPID uint32

	for {
		processName := windows.UTF16ToString(entry.ExeFile[:])
		parentMap[entry.ParentProcessID] = append(parentMap[entry.ParentProcessID], entry.ProcessID)

		if strings.EqualFold(processName, "WhatsApp.exe") || strings.EqualFold(processName, "WhatsApp.Root.exe") {
			rootPID = entry.ProcessID
		}

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				break
			}
			return nil, err
		}
	}

	if rootPID == 0 {
		return nil, errors.New("WhatsApp process not found")
	}

	// BFS to collect all descendants
	var result []uint32
	queue := []uint32{rootPID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		children := parentMap[current]
		for _, child := range children {
			// Avoid self-references just in case
			if child != current {
				queue = append(queue, child)
			}
		}
	}

	return result, nil
}
