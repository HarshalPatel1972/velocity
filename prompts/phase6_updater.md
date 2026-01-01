# Phase 6: Auto-Updater

Self-update using GitHub Releases and installer swap.

## Prompt

```
Create internal/updater package.

Requirements:
- CurrentVersion constant in main.go
- CheckForUpdates() queries GitHub Releases API
- Compare tag_name vs CurrentVersion
- Download installer to OS temp folder
- Launch installer + os.Exit(0) to unlock file
- Add "Check for Updates" to tray menu
- Silent check on startup
```

## API

```
GET https://api.github.com/repos/HarshalPatel1972/velocity/releases/latest
```

## Installer Config

```ini
[Setup]
CloseApplications=yes
RestartApplications=yes
```
