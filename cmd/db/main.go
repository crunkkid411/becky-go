package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B8")).
			Padding(0, 1)

	recordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type Habit struct {
	ID        int
	Name      string
	Count     int
	LastDone  string
	CreatedAt time.Time
}

type model struct {
	habits   []Habit
	cursor   int
	query    string
	result   string
	db       *sql.DB
	message  string
}

type dbMsg struct {
	habits  []Habit
	err     error
}

type queryMsg struct {
	result string
	err    error
}

func initDB() (*sql.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(homeDir, ".becky")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dbDir, "habits.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS habits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		count INTEGER DEFAULT 0,
		last_done TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initialModel() model {
	db, err := initDB()
	if err != nil {
		return model{message: fmt.Sprintf("DB init error: %v", err)}
	}

	return model{
		db:     db,
		cursor: 0,
		habits: []Habit{},
	}
}

func (m model) Init() tea.Cmd {
	return loadHabits(m.db)
}

func loadHabits(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		rows, err := db.Query("SELECT id, name, count, last_done, created_at FROM habits ORDER BY created_at DESC")
		if err != nil {
			return dbMsg{err: err}
		}
		defer rows.Close()

		var habits []Habit
		for rows.Next() {
			var h Habit
			var lastDone sql.NullString
			var createdAt string
			if err := rows.Scan(&h.ID, &h.Name, &h.Count, &lastDone, &createdAt); err != nil {
				return dbMsg{err: err}
			}
			if lastDone.Valid {
				h.LastDone = lastDone.String
			}
			h.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
			habits = append(habits, h)
		}

		return dbMsg{habits: habits}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.habits)-1 {
				m.cursor++
			}
		case "n":
			// Add new habit
			newHabit := Habit{Name: "New Habit", CreatedAt: time.Now()}
			m.habits = append([]Habit{newHabit}, m.habits...)
			m.message = "Added new habit. Edit in config."
		case "x":
			// Delete current habit
			if len(m.habits) > 0 && m.db != nil {
				habit := m.habits[m.cursor]
				_, err := m.db.Exec("DELETE FROM habits WHERE id = ?", habit.ID)
				if err == nil {
					m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
					if m.cursor >= len(m.habits) && m.cursor > 0 {
						m.cursor--
					}
					m.message = fmt.Sprintf("Deleted: %s", habit.Name)
				}
			}
		case "d":
			// Mark as done today
			if len(m.habits) > 0 && m.db != nil {
				habit := m.habits[m.cursor]
				today := time.Now().Format("2006-01-02")
				_, err := m.db.Exec("UPDATE habits SET count = count + 1, last_done = ? WHERE id = ?", today, habit.ID)
				if err == nil {
					m.habits[m.cursor].Count++
					m.habits[m.cursor].LastDone = today
					m.message = fmt.Sprintf("Marked '%s' as done!", habit.Name)
				}
			}
		case "r":
			// Refresh
			return m, loadHabits(m.db)
		}
	case dbMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.habits = msg.habits
		}
	case tea.WindowSizeMsg:
		// Handle resize
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Becky DB") + "\n\n")
	b.WriteString(recordStyle.Render("Habit Tracker") + "\n\n")

	if len(m.habits) == 0 {
		b.WriteString("  No habits yet. Press 'n' to add one.\n\n")
	} else {
		for i, h := range m.habits {
			cursor := "  "
			style := recordStyle
			if i == m.cursor {
				cursor = "> "
				style = titleStyle
			}
			status := "○"
			if h.LastDone == time.Now().Format("2006-01-02") {
				status = "●"
			}
			b.WriteString(fmt.Sprintf("%s%s %s (Count: %d, Last: %s)\n", 
				cursor, status, style.Render(h.Name), h.Count, h.LastDone))
		}
	}

	b.WriteString("\n")
	if m.message != "" {
		b.WriteString(recordStyle.Render(m.message) + "\n\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate | d: done | x: delete | n: new | r: refresh | q: quit\n"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
