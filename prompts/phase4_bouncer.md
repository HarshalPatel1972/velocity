# Phase 4: Focus Bouncer

Prevent focus stealing with safety filters.

## Prompt

```
Create internal/watcher with SetWinEventHook.

Requirements:
- Hook EVENT_SYSTEM_FOREGROUND (no polling)
- Track lastGoodWindow
- Safety filters:
  1. Title contains "Call"/"Video" → ALLOW
  2. Alt key pressed → ALLOW
  3. Mouse inside window → ALLOW
  4. None of above → BLOCK + revert focus
- FlashWindow on blocked steal
```

## Key APIs
- `SetWinEventHook` - Focus event hook
- `GetWindowText` - Window title
- `GetAsyncKeyState` - Alt key check
- `GetCursorPos` / `GetWindowRect` - Mouse check
- `SetForegroundWindow` - Revert focus
