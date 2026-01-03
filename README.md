# ⚡ Velocity

**The Ultimate WhatsApp Desktop Optimizer for Windows**

Velocity is a lightweight system tray utility that dramatically improves WhatsApp Desktop performance by optimizing memory usage, managing CPU priority, and preventing annoying focus stealing.

---

## ⚠️ Important: Read Before Installing

### Is This Safe?

**Yes, but please understand what you're installing:**

| Concern | Explanation |
|---------|-------------|
| **Admin Rights** | Required to access WhatsApp's memory. Without this, the app cannot work. |
| **SmartScreen Warning** | You'll see "Windows protected your PC" because this app is **not code-signed**. Code signing costs $200-400/year which I haven't purchased. |
| **Open Source** | 100% of the code is visible in this repository. You can audit every line. |
| **No Telemetry** | Zero data collection. No internet connection except for update checks to GitHub. |
| **No Background Mining** | This is not cryptomining or malware. Check the source code yourself. |

### What Does It Actually Do?

1. **Reads WhatsApp process list** - To find which processes to optimize
2. **Calls Windows APIs** - `SetProcessWorkingSetSizeEx` to trim RAM
3. **Sets process priority** - Lower priority when WhatsApp is in background
4. **Monitors focus changes** - To prevent WhatsApp stealing focus
5. **Checks GitHub for updates** - Only when you click "Check for Updates"

### Why Admin Rights?

Windows protects running applications' memory. To tell Windows to release unused RAM from WhatsApp, we need Administrator privileges. There's no way around this for memory optimization tools.

---

## 📥 Installation Guide

### Step 1: Download

1. Go to [**Releases**](https://github.com/HarshalPatel1972/velocity/releases)
2. Download the latest `Velocity_Setup_vX.X.X.exe`

### Step 2: Handle SmartScreen Warning

When you run the installer, Windows will show a warning:

```
Windows protected your PC
Microsoft Defender SmartScreen prevented an unrecognized app from starting.
```

**This is normal for unsigned applications.** To proceed:

1. Click **"More info"**
2. Click **"Run anyway"**

> 💡 **Why this happens:** Code signing certificates cost $200-400/year. As a free, open-source project, I haven't purchased one. The warning does NOT mean the app is dangerous.

### Step 3: Approve Administrator Access

You'll see a UAC (User Account Control) prompt:

```
Do you want to allow this app to make changes to your device?
```

Click **Yes**. This is required for memory optimization to work.

### Step 4: Install

1. Choose installation folder (default: `C:\Program Files\Velocity`)
2. ✅ Check **"Start Velocity when Windows boots"** (recommended)
3. Click **Install** → **Finish**

### Step 5: Verify It's Running

Look for the ⚡ icon in your **System Tray** (bottom-right, near the clock).

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
- Checks GitHub for new releases
- One-click update from the system tray
- Seamless installer-based updates

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

## 🖥️ System Requirements

| Requirement | Details |
|-------------|---------|
| **OS** | Windows 10 (version 1903+) or Windows 11 |
| **Privileges** | Administrator (required) |
| **Dependencies** | None - fully self-contained |
| **Disk Space** | ~10 MB |
| **RAM** | Minimal (~5 MB while running) |

> ⚠️ **Windows 7/8:** The app will run but EcoQoS (CPU optimization) won't work. Memory trimming will still function.

---

## 🗑️ Uninstalling

1. Go to **Settings → Apps → Installed Apps**
2. Find **Velocity**
3. Click **Uninstall**

This removes:
- The application files
- The auto-start registry entry
- All traces of the application

Your WhatsApp data is never touched.

---

## 🔧 Building from Source

If you don't trust pre-built binaries, build it yourself:

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)
- [Inno Setup 6](https://jrsoftware.org/isinfo.php) (optional, for installer)

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

## 🛡️ Safety Features

| Feature | How It Protects You |
|---------|---------------------|
| **Call Detection** | Never blocks incoming WhatsApp calls - checks window title for "Call" |
| **Alt-Tab Respect** | Always allows keyboard navigation |
| **Click-Through** | Intentional mouse clicks are never blocked |
| **No Data Access** | Cannot read your messages, contacts, or media |
| **No Network Access** | Only connects to GitHub API for update checks |
| **Open Source** | Full source code available for audit |

---

## ❓ FAQ

**Q: Will this break WhatsApp?**  
A: No. Velocity only interacts with Windows APIs, not WhatsApp itself. WhatsApp continues to function normally.

**Q: Is my data safe?**  
A: Velocity cannot access your messages, contacts, or files. It only sees process IDs and window titles.

**Q: Why does RAM usage go back up?**  
A: Windows reallocates memory as WhatsApp needs it. Velocity trims it again every 30 seconds.

**Q: Can I run this without admin rights?**  
A: No. Windows requires admin privileges to modify other applications' memory.

---

## 📜 License

MIT License - See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with:
- [systray](https://github.com/getlantern/systray) - System tray support
- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) - Windows API bindings
- [go-winres](https://github.com/tc-hib/go-winres) - Windows resource embedding

---

<p align="center">
  Made with ⚡ by <a href="https://github.com/HarshalPatel1972">Harshal Patel</a>
</p>

---

<p align="center">
  <sub><b>Reddit Verification:</b> This project is maintained by <a href="https://www.reddit.com/user/IllActive2550">u/IllActive2550</a></sub>
</p>

