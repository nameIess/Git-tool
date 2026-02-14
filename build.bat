@echo off
echo ============================================
echo   Git Tool - Build Script
echo ============================================
echo.

:: Step 1: Generate resource file with embedded icon
echo [1/2] Embedding icon.ico into resource...
go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -icon=internal/icon.ico -manifest="" -o resource_windows.syso
if %errorlevel% neq 0 (
    echo ERROR: Failed to generate resource file.
    pause
    exit /b 1
)
echo       resource_windows.syso created.
echo.

:: Step 2: Build the executable
echo [2/2] Building git-tool.exe...
go build -o git-tool.exe .
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
