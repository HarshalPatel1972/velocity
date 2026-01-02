# ⚡ Velocity

**The Ultimate WhatsApp Desktop Optimizer for Windows**

Velocity is a lightweight system tray utility that dramatically improves WhatsApp Desktop performance by optimizing memory usage, managing CPU priority, and preventing annoying focus stealing.

---

## ✨ Features

### 🧠 Smart Memory Trimmer
- Reduces WhatsApp RAM usage by **~280 MB** (from ~370 MB to ~90 MB)
- Automatically trims all WhatsApp processes including WebView2 components
- Runs every 30 seconds in the background

### ⚡ Adaptive CPU Governor (EcoQoS)
- Uses Windows EcoQoS to route background WhatsApp to efficiency cores
- Saves battery when WhatsApp is not in focus
- Instantly switches to performance mode when you open WhatsApp

### 🛡️ Focus Bouncer
- Prevents WhatsApp from stealing focus while you're typing
- Smart safety filters: **Never blocks incoming calls or video chats**
- Respects Alt-Tab and intentional clicks

### 🔄 Auto-Updater
- Checks GitHub for new releases automatically
- One-click update from the system tray
- Seamless installer-based updates

---

## 📥 Installation

### Download
1. Go to [**Releases**](https://github.com/HarshalPatel1972/velocity/releases)
2. Download `Velocity_Setup_vX.X.X.exe`
3. Run the installer (requires Administrator)

### Options
- ✅ **Auto-start with Windows** - Recommended
- ✅ **Desktop shortcut** - Optional

---

## 🖥️ Usage

Velocity runs silently in your **System Tray** (bottom-right, near the clock).

**Right-click the tray icon for options:**
| Menu Item | Action |
|-----------|--------|
| Status: Active | Shows Velocity is running |
| Force Trim Now | Immediately trim WhatsApp memory |
| Check for Updates | Check GitHub for new versions |
| Quit | Exit Velocity |

---

## 📊 Performance Impact

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| RAM Usage | ~370 MB | ~90 MB | **-76%** |
| Background CPU | Normal | Efficiency Mode | 🔋 Battery Saver |
| Focus Interruptions | Frequent | Blocked | ✅ No more stealing |

---

## 🔧 Building from Source

### Prerequisites
- Go 1.21+
- [Inno Setup 6](https://jrsoftware.org/isinfo.php) (for installer)

### Build
```powershell
# Clone the repo
git clone https://github.com/HarshalPatel1972/velocity.git
cd velocity

# Build executable
go build -ldflags "-s -w -H=windowsgui" -o velocity.exe ./cmd/velocity

# Build installer (optional)
.\deploy\build_release.bat
```

---

## 📁 Project Structure

```
velocity/
├── cmd/velocity/          # Entry point
├── internal/
│   ├── memory/            # RAM trimmer (Phase 1)
│   ├── cpu/               # EcoQoS governor (Phase 2)
│   ├── tray/              # System tray icon (Phase 3)
│   ├── watcher/           # Focus bouncer (Phase 4)
│   ├── updater/           # Auto-updater (Phase 6)
│   ├── utils/             # Process helpers
│   └── window/            # Focus detection
├── deploy/                # Installer scripts (Phase 5)
└── prompts/               # AI prompt documentation
```

---

## 🛡️ Safety

- **Incoming calls are never blocked** - Smart title detection for "Call" and "Video"
- **Alt-Tab always works** - Keyboard shortcuts respected
- **Click-through** - Intentional mouse clicks are allowed
- **Admin required** - Needed for process memory operations

---

## 📜 License

MIT License - See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with:
- [systray](https://github.com/getlantern/systray) - System tray support
- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) - Windows API bindings

---

<p align="center">
  Made with ⚡ by <a href="https://github.com/HarshalPatel1972">Harshal Patel</a>
</p>
