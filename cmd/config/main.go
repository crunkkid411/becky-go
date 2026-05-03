package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type Config struct {
	Theme        string `json:"theme"`
	Editor       string `json:"editor"`
	Terminal     string `json:"terminal"`
	LLMProvider  string `json:"llm_provider"`
	LLMModel     string `json:"llm_model"`
	STTEnabled   bool   `json:"stt_enabled"`
	NotifyEnabled bool  `json:"notify_enabled"`
}

type model struct {
	config     Config
	cursor     int
	keys       []string
	values     []string
	dirty      bool
	message    string
}

type saveMsg struct {
	err error
}

func defaultConfig() Config {
	return Config{
		Theme:        "default",
		Editor:       "code",
		Terminal:     "wezterm",
		LLMProvider:  "local",
		LLMModel:     "qwen-3.5-4b",
		STTEnabled:   false,
		NotifyEnabled: true,
	}
}

func loadConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig()
	}

	configPath := filepath.Join(homeDir, ".becky", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultConfig()
	}

	return config
}

func initialModel() model {
	config := loadConfig()
	keys := []string{"theme", "editor", "terminal", "llm_provider", "llm_model", "stt_enabled", "notify_enabled"}
	values := []string{
		config.Theme,
		config.Editor,
		config.Terminal,
		config.LLMProvider,
		config.LLMModel,
		fmt.Sprintf("%v", config.STTEnabled),
		fmt.Sprintf("%v", config.NotifyEnabled),
	}

	return model{
		config: config,
		cursor: 0,
		keys:   keys,
		values: values,
		dirty:  false,
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
			if m.cursor < len(m.keys)-1 {
				m.cursor++
			}
		case "enter":
			// Toggle boolean or edit string
			if m.keys[m.cursor] == "stt_enabled" || m.keys[m.cursor] == "notify_enabled" {
				current := m.values[m.cursor] == "true"
				m.values[m.cursor] = fmt.Sprintf("%v", !current)
				m.dirty = true
				m.message = fmt.Sprintf("Set %s to %v", m.keys[m.cursor], !current)
			} else {
				m.message = fmt.Sprintf("Edit %s in ~/.becky/config.json", m.keys[m.cursor])
			}
		case "s":
			if m.dirty {
				return m, saveConfig(m)
			}
		}
	case saveMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error saving: %v", msg.err)
		} else {
			m.message = "Configuration saved!"
			m.dirty = false
		}
	case tea.WindowSizeMsg:
		// Handle resize
	}
	return m, nil
}

func saveConfig(m model) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return saveMsg{err: err}
		}

		configDir := filepath.Join(homeDir, ".becky")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return saveMsg{err: err}
		}

		config := Config{
			Theme:        m.values[0],
			Editor:       m.values[1],
			Terminal:     m.values[2],
			LLMProvider:  m.values[3],
			LLMModel:     m.values[4],
			STTEnabled:   m.values[5] == "true",
			NotifyEnabled: m.values[6] == "true",
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return saveMsg{err: err}
		}

		configPath := filepath.Join(configDir, "config.json")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return saveMsg{err: err}
		}

		return saveMsg{}
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky Config") + "\n\n")
	b.WriteString(keyStyle.Render("Configuration Settings") + "\n\n")

	for i, key := range m.keys {
		cursor := "  "
		style := keyStyle
		if i == m.cursor {
			cursor = "> "
			style = valueStyle
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, style.Render(key), valueStyle.Render(m.values[i])))
	}

	b.WriteString("\n")
	if m.message != "" {
		b.WriteString(valueStyle.Render(m.message) + "\n\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate | Enter: toggle/edit | s: save | q: quit\n"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
