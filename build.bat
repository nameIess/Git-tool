@echo off
echo ============================================
echo   Git Tool - Build Script
echo ============================================
echo.

:: Step 1: Generate resource file with embedded icon
echo 1. Generating resource file (icon and version info)...
go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -icon=icon.ico -manifest="" -o resource_windows.syso
if %errorlevel% neq 0 (
    echo ERROR: Failed to generate resource file.
    pause
    exit /b 1
)
echo       resource_windows.syso created.
echo.

:: Step 2: Build the executable
echo 2. Compiling executable...
go build -ldflags "-s -w" -o git-tool.exe .
if %errorlevel% neq 0 (
    echo ERROR: Build failed.
    pause
    exit /b 1
)
echo       Build successful!
echo.

echo ============================================
echo   Done! git-tool.exe is ready.
echo ============================================
exit /b 3
