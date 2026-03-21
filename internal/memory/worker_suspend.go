package memory

import (
	"strings"
	"sync"
	"velocity/internal/utils"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

var (
	modntdll             = windows.NewLazySystemDLL("ntdll.dll")
	procNtSuspendProcess = modntdll.NewProc("NtSuspendProcess")
	procNtResumeProcess  = modntdll.NewProc("NtResumeProcess")
)

var (
	suspendedPIDs []uint32
	suspendMutex  sync.Mutex
)

type ProcessType string

const (
	TypeRenderer ProcessType = "renderer"
	TypeGPU      ProcessType = "gpu-process"
	TypeUtility  ProcessType = "utility"
	TypeUnknown  ProcessType = "unknown"
)

// fetchWebView2CmdLines queries WMI using go-ole to get the command line arguments for WebView2 processes.
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

	// Connect to WMI root CIMV2
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, "ROOT\\CIMV2")
	if err != nil {
		return cmdLines
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// Execute WMI Query
	query := "SELECT ProcessId, CommandLine FROM Win32_Process WHERE Name = 'msedgewebview2.exe'"
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return cmdLines
	}
	result := resultRaw.ToIDispatch()
	defer result.Release()

	// wmi collection uses _NewEnum
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

	// Iterate through the results
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

func ClassifyProcess(cmd string) ProcessType {
	if strings.Contains(cmd, "--type=gpu-process") {
		return TypeGPU
	}
	if strings.Contains(cmd, "--type=utility") {
		return TypeUtility
	}
	if strings.Contains(cmd, "--type=renderer") {
		return TypeRenderer
	}
	return TypeUnknown
}

func SuspendWorkers() {
	suspendMutex.Lock()
	defer suspendMutex.Unlock()

	treePIDs, err := utils.GetWhatsAppProcessTree()
	if err != nil || len(treePIDs) == 0 {
		return
	}

	treeMap := make(map[uint32]bool)
	for _, p := range treePIDs {
		treeMap[p] = true
	}

	cmdMap := fetchWebView2CmdLines()
	var toSuspend []uint32

	for pid, cmd := range cmdMap {
		if !treeMap[pid] {
			continue
		}

		class := ClassifyProcess(cmd)
		if class == TypeGPU || class == TypeUtility {
			toSuspend = append(toSuspend, pid)
		}
	}

	for _, pid := range toSuspend {
		hProcess, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, pid)
		if err == nil {
			procNtSuspendProcess.Call(uintptr(hProcess))
			windows.CloseHandle(hProcess)
			suspendedPIDs = append(suspendedPIDs, pid)
		}
	}
}

func ResumeWorkers() {
	suspendMutex.Lock()
	defer suspendMutex.Unlock()

	for _, pid := range suspendedPIDs {
		hProcess, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, pid)
		if err == nil {
			procNtResumeProcess.Call(uintptr(hProcess))
			windows.CloseHandle(hProcess)
		}
	}
	suspendedPIDs = nil
}

func ResumeWorkersSafe() {
	ResumeWorkers()
}


