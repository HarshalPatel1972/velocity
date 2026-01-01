# Phase 1: Memory Trimmer

Build the Smart Memory Trimmer module using Windows APIs.

## Prompt

```
Act as a Senior Go Systems Engineer. We are building a Windows utility called 'Velocity' to optimize WhatsApp Desktop.

Task: Initialize the project and build the 'Smart Memory Trimmer' module.

Requirements:
- Initialize a Go module named velocity
- Create folder structure: cmd/velocity and internal/memory
- Use golang.org/x/sys/windows for Windows APIs
- Implement EmptyWorkingSet/SetProcessWorkingSetSizeEx
- Find WhatsApp process tree and trim all child processes
- 30-second interval loop with timestamped logging
```

## Key APIs
- `CreateToolhelp32Snapshot` - Enumerate processes
- `SetProcessWorkingSetSizeEx` - Trim working set
- `Process32First/Next` - Iterate process list
