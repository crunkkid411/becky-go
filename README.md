## becky-go

A suite of Becky-like tools and cli commands to help a computer work like a computer so a human can think like a human.
Specifically with ADHD users in mind: humans who need tools that work immediately, not after 14 hours of debugging.
Tools and workflows customized to control the machine with user preferences built in to the workflows.
Able to be used immediately by any AI agent on the machine, and also plugged into various applications.

## Structure

```
becky-go/
├── README.md                  ← You are here
├── .gitignore                 ← bin/ and go.sum excluded from repo
├── build-first-tool.bat       ← Automated build script (Windows)
├── go.mod                     ← Go module definition (versioned)
├── cmd/                       ← Source code (versioned)
│   └── status/
│       └── main.go            ← becky-status source (Tool #1)
└── bin/                       ← Compiled binaries (LOCAL, .gitignore'd)
    └── becky-status.exe       ← Built on your machine, not in repo
```

**Key distinction:**
- `cmd/` = source code. Versioned in the repo. Every agent reads from here.
- `bin/` = compiled output. Built locally on YOUR machine. Added to your PATH. Never committed.

## Architecture

### Design Principles

- **Each tool is a standalone Go binary** — independent, no shared runtime
- **Go for UI/CLI** — bubbletea TUI, lipgloss styling, instant startup
- **Rust for compute** — STT, intent classification, SQLite (called via `os/exec`)
- **No webview, no JavaScript, no Python** — compiled binaries only
- **WezTerm is the display layer** — Go TUI adapts to whatever terminal size exists

### Go Toolbelt Stack

| Library | Version | Purpose |
|---|---|---|
| [bubbletea](https://github.com/charmbracelet/bubbletea) | v1.3.10 | TUI framework (Model-Update-View) |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | v1.1.0 | Styling (colors, borders, layouts) |
| [bubbles](https://github.com/charmbracelet/bubbles) | v1.0.0 | Pre-built TUI components (viewport, textinput, etc.) |

### Planned Go Tools

| Tool | Status | Description | Key Libraries |
|---|---|---|---|
| `becky-status` | ✅ Source ready | Status dashboard (time, system info, quick actions) | bubbletea, lipgloss, bubbles/viewport |
| `becky-ask` | Planned | Pipe text to LLM | mods/fantasy |
| `becky-forms` | Planned | Interactive prompts and menus | huh |
| `becky-pty` | Planned | Manage terminal sessions | creack/pty |
| `becky-config` | Planned | Settings management | cobra + viper |
| `becky-voice` | Planned | Voice input → STT → intent → action | Rust STT binary bridge |
| `becky-db` | Planned | Habit/state queries | Rust SQLite binary bridge |
| `becky-notify` | Planned | Desktop + Telegram notifications | winrt-notification |

### Rust Compute Binaries (separate project)

Go tools call these via `os/exec`. They live on PATH.

| Binary | Purpose | Crate | Model |
|---|---|---|---|
| `stt` | Speech → text | cpal + transcribe-rs | Parakeet v2 (ONNX) |
| `type-text` | Simulate keyboard input | enigo 0.6 | — |
| `intent` | NL command → JSON action | llama-cpp-rs | Qwen 3.5 4B GGUF |
| `db` | Habit/state SQLite queries | rusqlite 0.37 (bundled) | — |
| `notify` | Windows desktop notifications | notify-rust | — |

### IPC Pattern

```
Go tool (becky-voice)
    │
    ├── exec("stt", "--mic")
    │   └── returns: "open chrome"
    │
    ├── exec("intent", "open chrome")
    │   └── returns: {"action": "launch", "target": "chrome"}
    │
    └── exec("type-text", "chrome.exe")
        └── types into active window
```

**Rules:**
- stdin/stdout for simple data passing
- Named pipes on Windows for low-latency IPC (go-winio)
- No HTTP servers for local communication
- No WebSockets

# Installation Instructions

## Prerequisites

### 1. Go (Required)

- **Version:** Go 1.26+ (tested with 1.26.1)
- **Download:** https://go.dev/dl/
- **Verify:** `go version` should output `go1.26.x windows/amd64`
- **PATH:** Installer adds Go to PATH automatically. Restart terminal after install.

### 2. Rust (Optional — only for compute binaries)

- **Download:** https://rustup.rs/
- **Required only if** building `stt`, `intent`, `db`, or other Rust binaries
- **Not required** for Go tools alone

### 3. WezTerm (Recommended — display layer)

- **Download:** https://wezfurlong.org/wezterm/install/index.html
- Go tools render to any terminal, but WezTerm provides pane management and font control

## Quick Build (Windows)

Run from the `becky-go/` directory:

```cmd
build-first-tool.bat
```

This script will:
1. Verify Go is installed
2. Initialize the Go module (`github.com/crunkkid411/becky-go`)
3. Install pinned dependencies
4. Build `bin/becky-status.exe`

## Manual Build

```bash
# Step 1: Initialize module (if not already done)
go mod init github.com/crunkkid411/becky-go

# Step 2: Install dependencies (exact versions)
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v1.0.0
go mod tidy

# Step 3: Build
go build -o bin/becky-status.exe ./cmd/status

# Or run without building
go run ./cmd/status
```

## Cross-Compile

```bash
# Windows (from Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o bin/becky-status.exe ./cmd/status

# Linux (from Windows)
GOOS=linux GOARCH=amd64 go build -o bin/becky-status ./cmd/status

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/becky-status ./cmd/status
```

## Adding to PATH (Critical)

After building, add the `bin/` directory to your system PATH so any agent or terminal can call tools directly:

### Option 1: Via build script (recommended)

The build script prints the exact command. Run it in an **Administrator** terminal:

```cmd
setx PATH "%PATH%;C:\path\to\becky-go\bin"
```

### Option 2: PowerShell (one-liner)

```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\becky-go\bin", "Machine")
```

### Option 3: GUI

1. Win+R → `sysdm.cpl` → Advanced → Environment Variables
2. Under **System variables**, find `Path` → Edit → New
3. Add the full path to `becky-go\bin\`
4. OK out, restart all terminals

### Verify PATH

```cmd
becky-status
```

If the dashboard opens, PATH is configured correctly.

# Instructions for use

## Running the Status Dashboard

Once on PATH, from any terminal:

```cmd
becky-status
```

Or from the project directory:

```cmd
bin\becky-status.exe
```

**Keybindings:**
| Key | Action |
|---|---|
| `q` or `Ctrl+C` | Quit |
| `r` | Refresh |

The dashboard displays:
- Current time (live, updates every second)
- Date
- Go runtime info (OS/arch/version)
- Quick action hints

## Adding a New Tool

Every tool follows the same pattern. To add `becky-ask`:

### 1. Create the source directory

```bash
mkdir -p cmd/ask
```

### 2. Create `cmd/ask/main.go`

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type model struct {
    // your state
}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" || msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
    }
    return m, nil
}
func (m model) View() string {
    return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B8")).Render("becky-ask")
}

func main() {
    p := tea.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

### 3. Build it

```bash
go build -o bin/becky-ask.exe ./cmd/ask
```

### 4. Run it

```cmd
becky-ask
```

## Calling Rust Binaries from Go

```go
package main

import (
    "fmt"
    "os/exec"
)

func main() {
    // Call STT binary
    cmd := exec.Command("stt", "--input", "raw.pcm")
    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("STT error: %v", err)
        return
    }
    // output = transcribed text
    fmt.Printf("Transcribed: %s", string(output))
}
```

**Important:** Rust binaries must be on PATH or referenced by full path.

## Performance

| Metric | Value |
|---|---|
| Memory (Go TUI) | ~10-30MB |
| Startup time | Instant (compiled binary) |
| Binary size | ~4-8MB per tool (static, no CGO) |
| Rust STT latency | ~100ms |
| Rust intent latency | ~seconds (GPU-dependent) |

## Verified Build

| Item | Status |
|---|---|
| Go version | 1.26.1 windows/amd64 |
| bubbletea | v1.3.10 |
| lipgloss | v1.1.0 |
| bubbles | v1.0.0 |
| Build | ✅ Successful |
| Binary | `bin/becky-status.exe` (~4.4MB, static) |

## Anti-Patterns (Forbidden)

- ❌ Tauri — tools are standalone CLI
- ❌ Python — no runtime dependency
- ❌ HTTP server for local IPC — use stdin/stdout or named pipes
- ❌ Ollama — use llama-cpp-rs directly
- ❌ WebSockets — use named pipes on Windows
- ❌ `as any` / `@ts-ignore` equivalents — no error suppression in Go
- ❌ Empty catch blocks — handle all errors explicitly

## Agent Usage

Any AI agent on the system can use these tools (once on PATH):

```bash
# Agent checks status
becky-status

# Agent queries habits
db query "SELECT * FROM habits WHERE date = 'today'"

# Agent sends notification
notify --title "Task Complete" --body "Build finished successfully"

# Agent types text into active window
type-text "Hello from agent"
```

## Roadmap

1. ✅ `becky-status` — source ready, build verified
2. ⏳ `becky-ask` — LLM pipe tool
3. ⏳ `becky-forms` — interactive prompts
4. ⏳ `becky-pty` — terminal session management
5. ⏳ `becky-config` — settings management
6. ⏳ `becky-voice` — voice input pipeline
7. ⏳ `becky-db` — habit/state queries
8. ⏳ `becky-notify` — notifications

## To-Do

Update this section every time an agent works on this project. Mark completed items with `[x]` and add new ones at the bottom.

### Completed
- [x] Project structure created at `becky-go/`
- [x] Module initialized: `github.com/crunkkid411/becky-go`
- [x] `becky-status` source written and build verified
- [x] Exact dependency versions pinned (bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0)
- [x] Build script (`build-first-tool.bat`) working
- [x] README.md complete with architecture, install, usage, PATH instructions
- [x] All "Jarvis" references replaced with "Becky"
- [x] Project renamed from `go-toolbelt` to `becky-go`
- [x] `.gitignore` added — `bin/` and `go.sum` excluded from repo
- [x] `bin/` rebuilt with Becky naming (was showing "Jarvis Status")
- [x] `hairs-toolbelt/` cleaned of leftover Go files

### Active
- [ ] Next Step - Create GitHub repo for `becky-go/` (separate from `hairs-toolbelt/`), push all files
- [ ] Next Step - Add `becky-go/bin/` to system PATH so `becky-status` works from any terminal
- [ ] Next Step - Build `becky-ask` (LLM pipe tool using mods/fantasy)
- [ ] Next Step - Update `hairs-toolbelt/` architecture docs to reference `becky-go/` naming
