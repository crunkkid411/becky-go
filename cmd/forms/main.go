package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF75B8")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type choice struct {
	title       string
	description string
}

type model struct {
	choices      []choice
	cursor       int
	textInput    textinput.Model
	mode         string // "menu" or "input"
	selected     *choice
	inputValue   string
	submitted    bool
}

type doneMsg struct{}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your response..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	choices := []choice{
		{"Confirm", "Yes, proceed with this action"},
		{"Cancel", "No, go back"},
		{"Custom", "Enter a custom response"},
	}

	return model{
		choices:   choices,
		cursor:    0,
		textInput: ti,
		mode:      "menu",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.mode == "menu" && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.mode == "menu" && m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if m.mode == "menu" {
				if m.cursor == 2 {
					m.mode = "input"
					return m, m.textInput.Focus()
				}
				m.selected = &m.choices[m.cursor]
				m.submitted = true
				return m, tea.Quit
			} else if m.mode == "input" {
				m.inputValue = m.textInput.Value()
				m.submitted = true
				return m, tea.Quit
			}
		case "esc":
			if m.mode == "input" {
				m.mode = "menu"
				m.textInput.Blur()
			}
		}
	case tea.WindowSizeMsg:
		// Handle resize
	}

	var cmd tea.Cmd
	if m.mode == "input" {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky Forms") + "\n\n")

	if !m.submitted {
		if m.mode == "menu" {
			b.WriteString("Please select an option:\n\n")
			for i, choice := range m.choices {
				cursor := "  "
				style := normalStyle
				if i == m.cursor {
					cursor = "> "
					style = selectedStyle
				}
				b.WriteString(cursor + style.Render(choice.title) + "\n")
				b.WriteString("   " + normalStyle.Render(choice.description) + "\n\n")
			}
			b.WriteString(helpStyle.Render("↑/↓: navigate | Enter: select | q: quit\n"))
		} else if m.mode == "input" {
			b.WriteString("Enter custom response:\n\n")
			b.WriteString(m.textInput.View() + "\n\n")
			b.WriteString(helpStyle.Render("Enter: submit | Esc: back | q: quit\n"))
		}
	} else {
		b.WriteString(selectedStyle.Render("Selected: ") + m.choices[m.cursor].title + "\n")
		if m.inputValue != "" {
			b.WriteString(normalStyle.Render("Response: "+m.inputValue) + "\n")
		}
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
