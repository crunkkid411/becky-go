package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	sessionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type session struct {
	name   string
	pid    int
	active bool
}

type model struct {
	sessions []session
	cursor   int
	output   string
}

type execMsg struct {
	output string
	err    error
}

func initialModel() model {
	return model{
		sessions: []session{
			{name: "shell-1", pid: 1234, active: true},
			{name: "shell-2", pid: 1235, active: false},
		},
		cursor: 0,
		output: "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.sessions[m.cursor].active = true
				m.sessions[m.cursor+1].active = false
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
				m.sessions[m.cursor].active = true
				m.sessions[m.cursor-1].active = false
			}
		case "n":
			// Create new session
			newPid := 1000 + len(m.sessions) + 1
			newSession := session{
				name:   fmt.Sprintf("shell-%d", len(m.sessions)+1),
				pid:    newPid,
				active: true,
			}
			for i := range m.sessions {
				m.sessions[i].active = false
			}
			m.sessions = append(m.sessions, newSession)
			m.cursor = len(m.sessions) - 1
			m.output = fmt.Sprintf("Created new session: %s (PID: %d)", newSession.name, newPid)
		case "x":
			// Close current session
			if len(m.sessions) > 1 {
				closed := m.sessions[m.cursor]
				m.sessions = append(m.sessions[:m.cursor], m.sessions[m.cursor+1:]...)
				if m.cursor >= len(m.sessions) {
					m.cursor = len(m.sessions) - 1
				}
				if len(m.sessions) > 0 {
					m.sessions[m.cursor].active = true
				}
				m.output = fmt.Sprintf("Closed session: %s", closed.name)
			}
		case "r":
			// Run command in current session
			return m, runCommand("echo 'Becky PTY running'")
		}
	case execMsg:
		if msg.err != nil {
			m.output = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.output = msg.output
		}
	case tea.WindowSizeMsg:
		// Handle resize
	}
	return m, nil
}

func runCommand(cmdStr string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", cmdStr)
		}
		output, err := cmd.Output()
		if err != nil {
			return execMsg{err: err}
		}
		return execMsg{output: strings.TrimSpace(string(output))}
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky PTY") + "\n\n")
	b.WriteString(sessionStyle.Render("Terminal Sessions") + "\n\n")

	for i, s := range m.sessions {
		cursor := "  "
		style := sessionStyle
		if s.active {
			cursor = "> "
			style = activeStyle
		}
		if i == m.cursor {
			style = activeStyle
		}
		b.WriteString(fmt.Sprintf("%s%s (PID: %d)\n", cursor, style.Render(s.name), s.pid))
	}

	b.WriteString("\n")
	if m.output != "" {
		b.WriteString(sessionStyle.Render("Output: ") + m.output + "\n\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate | n: new | x: close | r: run | q: quit\n"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
