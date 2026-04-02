package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	timeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type tickMsg time.Time

type model struct {
	ready    bool
	viewport viewport.Model
	width    int
	height   int
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}
	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	now := time.Now()

	content := fmt.Sprintf(
		"%s\n\n%s\n%s\n\n%s\n%s\n%s\n\n%s\n%s\n\n%s",
		titleStyle.Render("Becky Status"),
		infoStyle.Render("Status: ")+"Online",
		infoStyle.Render("Time: ")+timeStyle.Render(now.Format("15:04:05")),
		infoStyle.Render("Date: ")+now.Format("Monday, January 2, 2006"),
		infoStyle.Render("Go: ")+runtimeInfo(),
		infoStyle.Render("OS: ")+runtime.GOOS,
		infoStyle.Render("Actions:"),
		"  [s] System info  [t] Tasks  [n] Notes",
		helpStyle.Render("q: quit | r: refresh"),
	)

	return content
}

func runtimeInfo() string {
	return fmt.Sprintf("%s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
