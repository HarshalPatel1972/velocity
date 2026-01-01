# Phase 5: Packaging

Create professional Windows installer.

## Prompt

```
Create Inno Setup installer and build script.

Requirements:
- installer.iss with:
  - Admin privileges
  - LZMA2/ultra64 compression
  - Auto-start registry key
  - Desktop icon option
  
- build_release.bat:
  - go build with stripped symbols
  - Find and run ISCC.exe
  - Output ready installer
```

## Usage

```powershell
.\deploy\build_release.bat
```

Output: `Output\Velocity_Setup_v1.0.0.exe`
