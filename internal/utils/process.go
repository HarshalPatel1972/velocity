package utils

import (
	"errors"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ErrProcessNotFound is returned when the target process is not running.
var ErrProcessNotFound = errors.New("process not found")

// FindPID searches for a running process by its executable name and returns its PID.
// The exeName should be the base name of the executable (e.g., "WhatsApp.exe").
func FindPID(exeName string) (uint32, error) {
	// Create a snapshot of all processes
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	// Get the first process
	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		return 0, err
	}

	// Iterate through all processes
	for {
		// Convert the ExeFile array to a string
		processName := windows.UTF16ToString(entry.ExeFile[:])
		
		// Case-insensitive comparison
		if strings.EqualFold(processName, exeName) {
			return entry.ProcessID, nil
		}

		// Move to the next process
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				break
			}
			return 0, err
		}
	}

	return 0, ErrProcessNotFound
}

// FindProcessesByPrefix returns all PIDs whose executable name starts with the given prefix.
// This is useful for finding all child processes of an application like WhatsApp.
func FindProcessesByPrefix(prefix string) ([]uint32, error) {
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

	var pids []uint32
	prefixLower := strings.ToLower(prefix)

	for {
		processName := windows.UTF16ToString(entry.ExeFile[:])

		// Case-insensitive prefix match
		if strings.HasPrefix(strings.ToLower(processName), prefixLower) {
			pids = append(pids, entry.ProcessID)
		}

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				break
			}
			return pids, err
		}
	}

	return pids, nil
}

// FindProcessTree finds the root process by name and returns all its descendant PIDs.
// This is useful for finding WhatsApp.Root.exe and all its WebView2 child processes.
func FindProcessTree(rootExeName string) ([]uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	// First pass: build a map of PID -> parent PID and find root
	type procInfo struct {
		parentPID uint32
		name      string
	}
	processes := make(map[uint32]procInfo)
	var rootPID uint32

	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		return nil, err
	}

	for {
		processName := windows.UTF16ToString(entry.ExeFile[:])
		processes[entry.ProcessID] = procInfo{
			parentPID: entry.ParentProcessID,
			name:      processName,
		}

		if strings.EqualFold(processName, rootExeName) {
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
		return nil, ErrProcessNotFound
	}

	// Second pass: find all descendants using BFS
	var result []uint32
	queue := []uint32{rootPID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Find all children of current process
		for pid, info := range processes {
			if info.parentPID == current && pid != current {
				queue = append(queue, pid)
			}
		}
	}

	return result, nil
}
