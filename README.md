# ⚡ Velocity

> ## ⚠️ Project status: discontinued — not recommended for use
>
> Velocity was built to fix WhatsApp Desktop lag on Windows. After measuring and testing it properly, **I found that it does not fix the lag**, and that some of its features can make WhatsApp worse. The claims this README used to make were not accurate, so I've replaced them with what I actually found.
>
> **If you have Velocity installed, please uninstall it** (see [Uninstalling](#-uninstalling)).

---

## Why the lag happens

In late 2025, Meta replaced the native WhatsApp app for Windows with a **WebView2 wrapper**: WhatsApp Web running inside an embedded Microsoft Edge (Chromium) window. Since then, many users have reported high RAM use (1 GB+), slow chat switching, choppy scrolling and typing lag.

I measured WhatsApp `2.2638.102.0` (WebView2 `154.0.4258.53`) on a Windows 11 laptop with an Intel i3-1220P and 16 GB of RAM:

| Observation | What it means |
|---|---|
| During lag, WhatsApp's **renderer process ran at 100–150% of a CPU core for ~20 seconds straight** | The slowness is WhatsApp's own JavaScript doing a lot of work. No outside tool can make that code faster — only Meta can. |
| The GPU process averaged ~10% CPU, with hardware acceleration on | Graphics is not the problem. |
| WhatsApp used **~2.2 GB of RAM** right after login and sync | It's one of the heaviest apps on a typical PC. |
| Lag was worst when the system had **~1.2 GB of RAM free**, and WhatsApp felt smooth with **~2.2 GB free** | Low system memory seems to matter. (Not tested in isolation.) |
| `WhatsApp.Root.exe` always launches in Windows efficiency mode (Idle priority + EcoQoS) | Real, but a **blind A/B test showed no difference you could feel** when it was turned off (scored 4/5 both ways, 4 rounds). |

## What Velocity did, and why it doesn't help

| Feature | Problem |
|---|---|
| **Memory trimmer** (`SetProcessWorkingSetSizeEx`) | Only lowers the number shown in Task Manager. The memory is pushed to the page file and has to be read back from disk when you use WhatsApp again, which adds lag. The "370 MB → 90 MB" figure this README used to show measured the working set, not real memory savings. |
| **Background throttling** (Idle priority + EcoQoS) | Windows already throttles background apps. Going further can delay message syncing. |
| **Focus Bouncer** | Can send focus back to your previous window when you open WhatsApp from the taskbar, a notification or a keyboard shortcut, because only Alt-Tab and clicks inside the window count as intentional. The "call" check only works on English window titles. |
| **Worker suspension** *(unreleased, `main` branch only)* | Suspends WebView2 "utility" processes, which include the **network, storage and audio services**. This can stop messages arriving and break calls while WhatsApp is in the background. If Velocity quits or crashes, those processes stay suspended. |
| **DevTools cache purge** *(unreleased, `main` branch only)* | Needs WhatsApp's remote-debugging port, which is off by default; turning it on is a security risk. Running code inside WhatsApp may also break WhatsApp's [Terms of Service](https://www.whatsapp.com/legal/terms-of-service), which forbid modifying the app. |
| **"Start with Windows"** | Doesn't work. Windows blocks programs that require admin rights from starting via the `Run` registry key. |
| **Auto-updater** | Downloads and silently installs new releases from GitHub **on every startup** with admin rights, without checking a signature or hash. (The old README said it only checked when you clicked the button; that was wrong.) |
| **Admin rights** | Not actually needed for the Windows calls Velocity makes on a process owned by the same user. |

## What actually helps (no software needed)

- **Free up RAM.** Close apps you aren't using, and turn off startup programs you don't need: Task Manager → **Startup apps**.
- **Keep WebView2 and Windows updated.**
- **Restart WhatsApp** if it's been open a long time.
- **Report the lag to WhatsApp** (in the app: Settings → Help → Contact us). Only Meta can fix the underlying cause.

---

## 🗑️ Uninstalling

1. Right-click the ⚡ tray icon → **Quit**
2. Go to **Settings → Apps → Installed apps**
3. Find **Velocity** → **Uninstall**

This removes the program and its startup entry. Velocity never touched your WhatsApp data.

If WhatsApp seems frozen or doesn't receive messages after uninstalling, fully quit it (tray icon → Quit, or end `WhatsApp.Root.exe` in Task Manager) and open it again.

---

## For developers

The code is left here for reference and learning. `cmd/memtest` and `cmd/heapinspect` are investigation tools; `cmd/heapinspect` currently doesn't compile.

```
velocity/
├── cmd/velocity/          # Tray app entry point
├── cmd/memtest/           # Live memory dashboard (test tool)
├── cmd/heapinspect/       # CDP heap inspector (doesn't compile)
├── internal/
│   ├── memory/            # Trimmer, memory priority, worker suspend, CDP purge
│   ├── cpu/               # EcoQoS / priority governor
│   ├── cdp/               # Chrome DevTools Protocol client
│   ├── watcher/           # Focus bouncer
│   ├── updater/           # Auto-updater
│   ├── tray/              # Tray icon
│   ├── utils/             # Process tree helpers
│   └── window/            # Foreground detection
├── deploy/                # Inno Setup installer
└── prompts/               # AI prompt documentation
```

---

## 📜 License

MIT License. See [LICENSE](LICENSE).

---

<p align="center">
  Made with ⚡ by <a href="https://github.com/HarshalPatel1972">Harshal Patel</a>
</p>

---

<p align="center">
  <sub><b>Reddit Verification:</b> This project is maintained by <a href="https://www.reddit.com/user/IllActive2550">u/IllActive2550</a></sub>
</p>
