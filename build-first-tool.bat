@echo off
REM becky-go build script
REM Run from the becky-go directory — builds the first tool

echo === Becky Go - First Build ===
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go is not installed or not in PATH.
    echo Download from: https://go.dev/dl/
    pause
    exit /b 1
)

echo [1/4] Go found:
go version
echo.

REM Initialize Go module if it doesn't exist
if not exist "go.mod" (
    echo [2/4] Initializing Go module...
    go mod init github.com/crunkkid411/becky-go
) else (
    echo [2/4] Module already exists.
)
echo.

REM Ensure source file exists (versioned in repo)
if not exist "cmd\status\main.go" (
    echo [ERROR] cmd\status\main.go not found.
    echo Make sure you cloned the repo and are running from the root.
    pause
    exit /b 1
)

echo [3/4] Installing dependencies...
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v1.0.0
go mod tidy
echo.

echo [4/4] Building becky-status...
if not exist "bin" mkdir bin
go build -o bin/becky-status.exe ./cmd/status
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

echo.
echo === BUILD SUCCESS ===
echo Binary: bin/becky-status.exe
echo.
echo Run it: bin\becky-status.exe
echo.
echo To add to PATH (run once as Administrator):
echo   setx PATH "%%PATH%%;%CD%\bin"
echo.
pause
