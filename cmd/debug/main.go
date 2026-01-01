package main

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {
	fmt.Println("Searching for WhatsApp processes...")
	fmt.Println()

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	found := false
	for {
		processName := windows.UTF16ToString(entry.ExeFile[:])

		if strings.Contains(strings.ToLower(processName), "whats") {
			fmt.Printf("Found: %s (PID: %d)\n", processName, entry.ProcessID)
			found = true
		}

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	if !found {
		fmt.Println("No WhatsApp processes found.")
	}
}
