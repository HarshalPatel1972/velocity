@echo off
setlocal enabledelayedexpansion

echo ============================================
echo   Velocity Release Builder v1.0
echo ============================================
echo.

:: Navigate to project root (parent of deploy folder)
cd /d "%~dp0.."

:: Step 1: Clean old binaries
echo [1/4] Cleaning old binaries...
if exist velocity.exe del velocity.exe
if exist Output rmdir /s /q Output
echo       Done.
echo.

:: Step 2: Build Go binary
echo [2/4] Building velocity.exe...
go build -ldflags "-s -w -H=windowsgui" -o velocity.exe ./cmd/velocity

if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] Go build failed!
    pause
    exit /b 1
)
echo       Done.
echo.

:: Step 3: Find Inno Setup Compiler
echo [3/4] Locating Inno Setup Compiler...

set "ISCC="

:: Check if ISCC is in PATH
where iscc >nul 2>&1
if %ERRORLEVEL% equ 0 (
    set "ISCC=iscc"
    goto :found_iscc
)

:: Check standard locations
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
    set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
    goto :found_iscc
)

if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
    set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
    goto :found_iscc
)

echo.
echo [ERROR] Inno Setup Compiler (ISCC.exe) not found!
echo         Please install Inno Setup 6 from: https://jrsoftware.org/isinfo.php
pause
exit /b 1

:found_iscc
echo       Found: %ISCC%
echo.

:: Step 4: Compile installer
echo [4/4] Compiling installer...
"%ISCC%" deploy\installer.iss

if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] Installer compilation failed!
    pause
    exit /b 1
)

echo.
echo ============================================
echo   BUILD SUCCESSFUL!
echo ============================================
echo.
echo   Output: Output\Velocity_Setup_v1.0.0.exe
echo.
echo   Ready for GitHub Releases!
echo ============================================
echo.

pause
