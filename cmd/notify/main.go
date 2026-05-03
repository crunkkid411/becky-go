package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type Notification struct {
	Title     string
	Body      string
	Timestamp time.Time
	Sent      bool
}

type model struct {
	notifications []Notification
	cursor        int
	title         string
	body          string
	status        string
	lastResult    string
}

type notifyMsg struct {
	err error
}

func initialModel() model {
	return model{
		notifications: []Notification{},
		cursor:        0,
		title:         "Test Notification",
		body:          "This is a test notification from Becky",
		status:        "Ready",
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
			}
		case "down", "j":
			if m.cursor < len(m.notifications)-1 {
				m.cursor++
			}
		case "s":
			// Send notification
			if m.title != "" && m.body != "" {
				return m, sendNotification(m.title, m.body)
			}
		case "t":
			// Send test notification
			return m, sendNotification("Becky Test", fmt.Sprintf("Test at %s", time.Now().Format("15:04:05")))
		case "c":
			// Clear history
			m.notifications = []Notification{}
			m.lastResult = "History cleared"
		case "r":
			// Reset status
			m.status = "Ready"
		}
	case notifyMsg:
		if msg.err != nil {
			m.status = "Failed"
			m.lastResult = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = "Sent"
			newNotif := Notification{
				Title:     m.title,
				Body:      m.body,
				Timestamp: time.Now(),
				Sent:      true,
			}
			m.notifications = append([]Notification{newNotif}, m.notifications...)
			m.lastResult = "Notification sent successfully"
		}
	case tea.WindowSizeMsg:
		// Handle resize
	}
	return m, nil
}

func sendNotification(title, body string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		
		if runtime.GOOS == "windows" {
			// Use PowerShell for Windows notifications
			psScript := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml(@"
<toast>
    <visual>
        <binding template="ToastText02">
            <text id="1">%s</text>
            <text id="2">%s</text>
        </binding>
    </visual>
</toast>
"@)

$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Becky").Show($toast)
`, title, body)
			cmd = exec.Command("powershell", "-Command", psScript)
		} else if runtime.GOOS == "darwin" {
			// macOS osascript
			cmd = exec.Command("osascript", "-e", 
				fmt.Sprintf(`display notification "%s" with title "%s"`, body, title))
		} else {
			// Linux - try notify-send
			cmd = exec.Command("notify-send", title, body)
		}

		err := cmd.Run()
		return notifyMsg{err: err}
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky Notify") + "\n\n")

	// Current input
	b.WriteString(successStyle.Render("Current Notification:") + "\n")
	b.WriteString(fmt.Sprintf("  Title: %s\n", m.title))
	b.WriteString(fmt.Sprintf("  Body: %s\n\n", m.body))

	// Status
	statusColor := successStyle
	if m.status == "Failed" {
		statusColor = errorStyle
	}
	b.WriteString(fmt.Sprintf("%s: %s\n\n", statusColor.Render("Status"), m.status))

	// Last result
	if m.lastResult != "" {
		b.WriteString(successStyle.Render("Last Result: ") + m.lastResult + "\n\n")
	}

	// History
	if len(m.notifications) > 0 {
		b.WriteString(successStyle.Render("Recent Notifications:") + "\n\n")
		for i, n := range m.notifications {
			cursor := "  "
			style := successStyle
			if i == m.cursor {
				cursor = "> "
				style = titleStyle
			}
			timeStr := n.Timestamp.Format("15:04:05")
			b.WriteString(fmt.Sprintf("%s%s [%s] %s\n", cursor, style.Render(n.Title), timeStr, n.Body))
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("s: send | t: test | c: clear | ↑/↓: navigate | q: quit\n"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
