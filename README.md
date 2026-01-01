# Velocity

A Windows utility to optimize the WhatsApp Desktop application.

## Features

- **Memory Trimmer** - Reduces WhatsApp RAM usage by ~280 MB
- **Adaptive QoS** - Uses Windows EcoQoS to save battery when backgrounded
- **Focus Bouncer** - Prevents WhatsApp from stealing focus while typing
- **System Tray** - Runs silently in the background

## Installation

1. Download `Velocity_Setup_v1.0.0.exe` from [Releases](../../releases)
2. Run the installer (requires Administrator)
3. Velocity starts automatically with Windows

## Building from Source

### Prerequisites
- Go 1.21+
- Inno Setup 6 (for installer)

### Build
```powershell
# Build executable only
go build -ldflags "-s -w -H=windowsgui" -o velocity.exe ./cmd/velocity

# Build installer (requires Inno Setup)
.\deploy\build_release.bat
```

## Project Structure

```
velocity/
├── cmd/velocity/           # Entry point
├── internal/
│   ├── memory/trimmer.go   # RAM optimization
│   ├── cpu/governor.go     # Priority & EcoQoS
│   ├── watcher/bouncer.go  # Focus protection
│   ├── tray/               # System tray icon
│   ├── window/             # Focus detection
│   └── utils/process.go    # Process helpers
├── deploy/                 # Installer scripts
└── prompts/                # AI prompt documentation
```

## License

MIT
