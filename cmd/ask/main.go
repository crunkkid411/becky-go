package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	responseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type model struct {
	prompt     string
	input      string
	response   string
	cursorBlink bool
	done       bool
}

type tickMsg struct{}
type responseMsg string

func initialModel() model {
	return model{
		prompt: "Ask Becky:",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.input != "" && !m.done {
				m.done = true
				return m, getResponse(m.input)
			}
		default:
			if !m.done && len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		// Handle resize if needed
	case tickMsg:
		m.cursorBlink = !m.cursorBlink
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg{}
		})
	case responseMsg:
		m.response = string(msg)
	}
	return m, nil
}

func getResponse(input string) tea.Cmd {
	return func() tea.Msg {
		// Simulate LLM response - in production, pipe to actual LLM
		response := fmt.Sprintf("Processing: %s\n[LLM integration pending]", input)
		return responseMsg(response)
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(promptStyle.Render(m.prompt) + "\n\n")
	b.WriteString(inputStyle.Render("> " + m.input))
	if m.cursorBlink && !m.done {
		b.WriteString(inputStyle.Render("█"))
	}
	b.WriteString("\n\n")

	if m.response != "" {
		b.WriteString(responseStyle.Render(m.response) + "\n\n")
	}

	if !m.done {
		b.WriteString(helpStyle.Render("Type your question and press Enter | q: quit\n"))
	} else {
		b.WriteString(helpStyle.Render("Press Enter to ask another question | q: quit\n"))
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
