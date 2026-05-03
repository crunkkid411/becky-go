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
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type model struct {
	listening    bool
	transcribed  string
	intent       string
	lastAction   string
	status       string
	listenTime   time.Time
}

type tickMsg time.Time
type sttResult string
type intentResult string

func initialModel() model {
	return model{
		listening:  false,
		transcribed: "",
		intent:     "",
		lastAction: "",
		status:     "Ready",
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
		case "l":
			if !m.listening {
				m.listening = true
				m.status = "Listening..."
				m.listenTime = time.Now()
				return m, startListening()
			} else {
				m.listening = false
				m.status = "Stopped"
			}
		case "r":
			// Reset
			m.transcribed = ""
			m.intent = ""
			m.lastAction = ""
			m.status = "Ready"
			m.listening = false
		}
	case tickMsg:
		if m.listening {
			elapsed := time.Since(m.listenTime)
			if elapsed > 5*time.Second {
				m.listening = false
				m.status = "Timeout"
				return m, nil
			}
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	case sttResult:
		m.transcribed = string(msg)
		m.status = "Processing intent..."
		return m, processIntent(string(msg))
	case intentResult:
		m.intent = string(msg)
		m.lastAction = fmt.Sprintf("Executed: %s", string(msg))
		m.status = "Complete"
		m.listening = false
	}
	return m, nil
}

func startListening() tea.Cmd {
	return func() tea.Msg {
		// Simulate STT - in production, call stt binary via os/exec
		time.Sleep(2 * time.Second)
		return sttResult("open chrome")
	}
}

func processIntent(text string) tea.Cmd {
	return func() tea.Msg {
		// Simulate intent classification - in production, call intent binary
		time.Sleep(1 * time.Second)
		
		// Simple keyword matching for demo
		lower := strings.ToLower(text)
		if strings.Contains(lower, "chrome") {
			return intentResult("{\"action\": \"launch\", \"target\": \"chrome.exe\"}")
		} else if strings.Contains(lower, "notepad") {
			return intentResult("{\"action\": \"launch\", \"target\": \"notepad.exe\"}")
		} else if strings.Contains(lower, "time") {
			return intentResult("{\"action\": \"say\", \"text\": \"" + time.Now().Format("3:04 PM") + "\"}")
		}
		
		return intentResult("{\"action\": \"unknown\", \"text\": \"" + text + "\"}")
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky Voice") + "\n\n")

	// Status indicator
	statusIndicator := "○"
	style := statusStyle
	if m.listening {
		statusIndicator = "●"
		style = errorStyle
	}
	b.WriteString(fmt.Sprintf("%s %s: %s\n\n", statusIndicator, style.Render("Status"), m.status))

	// Transcription
	b.WriteString(statusStyle.Render("Transcribed:") + "\n")
	if m.transcribed != "" {
		b.WriteString("  " + m.transcribed + "\n")
	} else {
		b.WriteString("  (waiting for speech...)\n")
	}

	b.WriteString("\n")

	// Intent
	b.WriteString(statusStyle.Render("Intent:") + "\n")
	if m.intent != "" {
		b.WriteString("  " + m.intent + "\n")
	} else {
		b.WriteString("  (processing...)\n")
	}

	b.WriteString("\n")

	// Last action
	if m.lastAction != "" {
		b.WriteString(statusStyle.Render("Last Action:") + "\n")
		b.WriteString("  " + m.lastAction + "\n\n")
	}

	b.WriteString(helpStyle.Render("l: start/stop listening | r: reset | q: quit\n"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
