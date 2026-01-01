# Phase 3: System Tray (Ghost Mode)

Turn the app into a background service with tray icon.

## Prompt

```
Implement GUI using github.com/getlantern/systray.

Requirements:
- No console window (-ldflags -H=windowsgui)
- Embedded icon as byte array
- Menu: Status, Force Trim Now, Quit
- Main loop in goroutine (don't block UI thread)
```

## Key Components
- `systray.Run()` - Main entry point
- `systray.SetIcon()` - Tray icon
- `systray.AddMenuItem()` - Menu items
