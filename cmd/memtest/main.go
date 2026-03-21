package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"velocity/internal/utils"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	modntdll    = windows.NewLazySystemDLL("ntdll.dll")
	modpsapi    = windows.NewLazySystemDLL("psapi.dll")

	procGetProcessInformation    = modkernel32.NewProc("GetProcessInformation")
	procGetProcessMemoryInfo     = modpsapi.NewProc("GetProcessMemoryInfo")
	procCreateToolhelp32Snapshot = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procThread32First            = modkernel32.NewProc("Thread32First")
	procThread32Next             = modkernel32.NewProc("Thread32Next")
	procSuspendThread            = modkernel32.NewProc("SuspendThread")
	procResumeThread             = modkernel32.NewProc("ResumeThread")
	procOpenThread               = modkernel32.NewProc("OpenThread")
)

const ProcessMemoryPriority = 0
const TH32CS_SNAPTHREAD = 0x00000004
const THREAD_SUSPEND_RESUME = 0x0002

type MEMORY_PRIORITY_INFORMATION struct {
	MemoryPriority uint32
}

type THREADENTRY32 struct {
	dwSize             uint32
	cntUsage           uint32
	th32ThreadID       uint32
	th32OwnerProcessID uint32
	tpBasePri          int32
	tpDeltaPri         int32
	dwFlags            uint32
}

type PROCESS_MEMORY_COUNTERS struct {
	CB                         uint32
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

func main() {
	fmt.Println("Starting Velocity Memory Monitor Test Harness...")

	// Start the process tree updater from utils
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		printDashboard()
	}
}

func printDashboard() {
	cmdMap := fetchWebView2CmdLines()

	pids, _ := utils.GetWhatsAppProcessTree()

	if len(pids) == 0 {
		fmt.Print("\033[H\033[2J")
		fmt.Println("=== Velocity Memory Monitor ===")
		fmt.Println("WhatsApp is NOT running.")
		return
	}

	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println("=== Velocity Memory Monitor ===")

	var stateStr = "UNKNOWN"
	if isForeground(pids) {
		stateStr = "FOREGROUNDED"
	} else {
		stateStr = "BACKGROUNDED"
	}

	fmt.Printf("WhatsApp state : %s\n", stateStr)
	fmt.Printf("PIDs tracked   : %v   (root + %d children)\n\n", pids, len(pids)-1)
	fmt.Printf("%-8s %-15s %-15s %-15s %-10s\n", "PID", "Type", "Working Set", "Mem Priority", "Suspended?")

	var totalMB uint64 = 0

	for _, pid := range pids {
		ptype := "child"
		if cmd, ok := cmdMap[pid]; ok {
			if strings.Contains(cmd, "--type=gpu-process") {
				ptype = "gpu-process"
			} else if strings.Contains(cmd, "--type=utility") {
				ptype = "utility"
			} else if strings.Contains(cmd, "--type=renderer") {
				ptype = "renderer"
			} else {
				ptype = "webview2-other"
			}
		} else {
			// If it's the root process from tracking
			if isRootProcess(pid) {
				ptype = "root"
			}
		}

		wsSize, _ := getWorkingSet(pid)
		totalMB += (wsSize / 1024 / 1024)
		wsStr := fmt.Sprintf("%d MB", wsSize/1024/1024)

		memPri := "NORMAL"
		if getMemPriorityStr(pid) == 4 {
			memPri = "BELOW_NORMAL"
		}

		suspendStr := "No"
		if isProcessSuspended(pid) {
			suspendStr = "Yes"
		}

		fmt.Printf("%-8d %-15s %-15s %-15s %-10s\n", pid, ptype, wsStr, memPri, suspendStr)
	}

	fmt.Printf("\nTotal  : %d MB\n", totalMB)
}

func isForeground(pids []uint32) bool {
	// Let's rely on GetForegroundWindow to see if any matches
	hwnd := windows.GetForegroundWindow()
	var fgPid uint32
	windows.GetWindowThreadProcessId(hwnd, &fgPid)
	for _, p := range pids {
		if p == fgPid {
			return true
		}
	}
	return false
}

func isRootProcess(pid uint32) bool {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(snapshot)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		if entry.ProcessID == pid {
			name := windows.UTF16ToString(entry.ExeFile[:])
			if strings.EqualFold(name, "whatsapp.exe") || strings.EqualFold(name, "whatsapp.root.exe") {
				return true
			}
		}
		err = windows.Process32Next(snapshot, &entry)
	}
	return false
}

func getWorkingSet(pid uint32) (uint64, error) {
	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(hProcess)

	var pmc PROCESS_MEMORY_COUNTERS
	pmc.CB = uint32(unsafe.Sizeof(pmc))

	ret, _, _ := procGetProcessMemoryInfo.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&pmc)),
		uintptr(pmc.CB),
	)
	if ret == 0 {
		return 0, fmt.Errorf("GetProcessMemoryInfo failed")
	}

	return uint64(pmc.WorkingSetSize), nil
}

func getMemPriorityStr(pid uint32) uint32 {
	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 5 // default fallback
	}
	defer windows.CloseHandle(hProcess)

	var pmi MEMORY_PRIORITY_INFORMATION
	ret, _, _ := procGetProcessInformation.Call(
		uintptr(hProcess),
		uintptr(ProcessMemoryPriority),
		uintptr(unsafe.Pointer(&pmi)),
		unsafe.Sizeof(pmi),
	)
	if ret == 0 {
		return 5
	}
	return pmi.MemoryPriority
}

func isProcessSuspended(pid uint32) bool {
	snapshot, _, err := procCreateToolhelp32Snapshot.Call(uintptr(TH32CS_SNAPTHREAD), 0)
	if snapshot == ^uintptr(0) || err != syscall.Errno(0) {
		return false
	}
	defer windows.CloseHandle(windows.Handle(snapshot))

	var threadEntry THREADENTRY32
	threadEntry.dwSize = uint32(unsafe.Sizeof(threadEntry))

	ret, _, _ := procThread32First.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
	if ret == 0 {
		return false
	}

	var suspendedCount = 0

	for {
		if threadEntry.th32OwnerProcessID == pid {
			hThread, _, _ := procOpenThread.Call(uintptr(THREAD_SUSPEND_RESUME), 0, uintptr(threadEntry.th32ThreadID))
			if hThread != 0 {
				suspendRet, _, _ := procSuspendThread.Call(hThread)
				if suspendRet != ^uintptr(0) && suspendRet > 0 {
					// Thread was already suspended
					suspendedCount++
				}
				// Always resume if we suspended it
				procResumeThread.Call(hThread)
				windows.CloseHandle(windows.Handle(hThread))
			}

			if suspendedCount > 0 {
				return true
			}
		}

		ret, _, _ := procThread32Next.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
		if ret == 0 {
			break
		}
	}

	return suspendedCount > 0
}

func fetchWebView2CmdLines() map[uint32]string {
	cmdLines := make(map[uint32]string)

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return cmdLines
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return cmdLines
	}
	defer wmi.Release()

	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, "ROOT\\CIMV2")
	if err != nil {
		return cmdLines
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := "SELECT ProcessId, CommandLine FROM Win32_Process WHERE Name = 'msedgewebview2.exe'"
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return cmdLines
	}
	result := resultRaw.ToIDispatch()
	defer result.Release()

	enumProp, err := result.GetProperty("_NewEnum")
	if err != nil {
		return cmdLines
	}
	enumRaw := enumProp.ToIUnknown()
	defer enumRaw.Release()

	enum, err := enumRaw.IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return cmdLines
	}
	defer enum.Release()

	for {
		itemVariant, fetched, err := enum.Next(1)
		if err != nil || fetched == 0 {
			break
		}

		itemDetails := itemVariant.ToIDispatch()

		pidProp, err1 := oleutil.GetProperty(itemDetails, "ProcessId")
		cmdProp, err2 := oleutil.GetProperty(itemDetails, "CommandLine")

		if err1 == nil && err2 == nil && pidProp.Value() != nil && cmdProp.Value() != nil {
			var pid uint32
			switch val := pidProp.Value().(type) {
			case int32:
				pid = uint32(val)
			case float64:
				pid = uint32(val)
			case int64:
				pid = uint32(val)
			}
			if pid > 0 {
				cmdLines[pid] = cmdProp.ToString()
			}
		}

		if pidProp != nil {
			pidProp.Clear()
		}
		if cmdProp != nil {
			cmdProp.Clear()
		}
		itemDetails.Release()
		itemVariant.Clear()
	}

	return cmdLines
}
