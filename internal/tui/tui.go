package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joelpeckham/ocrogram/internal/config"
	"github.com/joelpeckham/ocrogram/internal/service"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	labelStyle = lipgloss.NewStyle().Faint(true).Width(8)
	hintStyle  = lipgloss.NewStyle().Faint(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type model struct {
	cfg     config.Config
	running bool
	editing bool
	input   textinput.Model
	note    string
	err     error
}

type startedMsg struct{}
type stoppedMsg struct{}
type savedMsg struct{}
type errMsg struct{ err error }

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Screenshot folder"
	ti.CharLimit = 512
	ti.Width = 64

	cfg, err := config.Load()
	return model{
		cfg:     cfg,
		running: service.Running(),
		input:   ti,
		err:     err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editing {
		return m.updateEditing(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.note = "starting…"
			m.err = nil
			return m, startCmd(m.cfg)
		case "s":
			m.note = "stopping…"
			m.err = nil
			return m, stopCmd()
		case "f":
			m.editing = true
			m.err = nil
			m.note = ""
			m.input.SetValue(m.cfg.ScreenshotDir)
			m.input.CursorEnd()
			return m, m.input.Focus()
		}
	case startedMsg:
		m.running = true
		m.note = "started"
	case stoppedMsg:
		m.running = false
		m.note = "stopped"
	case savedMsg:
		m.note = "folder saved"
	case errMsg:
		m.err = msg.err
		m.note = ""
		m.running = service.Running()
	}
	return m, nil
}

func (m model) updateEditing(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.editing = false
			m.input.Blur()
			return m, nil
		case "enter":
			path := strings.TrimSpace(m.input.Value())
			m.editing = false
			m.input.Blur()
			if path == "" || path == m.cfg.ScreenshotDir {
				return m, nil
			}
			m.cfg.ScreenshotDir = path
			m.note = "saving…"
			return m, saveCmd(m.cfg, m.running)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("ocrogram"))
	b.WriteString("\n\n")

	if m.editing {
		b.WriteString(labelStyle.Render("Folder"))
		b.WriteString(" ")
		b.WriteString(m.input.View())
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Login"))
		b.WriteString(" ")
		b.WriteString(loginLabel(m.running))
		b.WriteString("\n\n")
		b.WriteString(hintStyle.Render("enter  save    esc  cancel"))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(labelStyle.Render("Folder"))
	b.WriteString(" ")
	b.WriteString(m.cfg.ScreenshotDir)
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Login"))
	b.WriteString(" ")
	b.WriteString(loginLabel(m.running))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err.Error()))
		b.WriteString("\n")
	} else if m.note != "" {
		b.WriteString("\n")
		b.WriteString(okStyle.Render(m.note))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("enter  start at login    s  stop    f  change folder    q  quit"))
	b.WriteString("\n")
	return b.String()
}

func loginLabel(running bool) string {
	if running {
		return okStyle.Render("running")
	}
	return "not running"
}

func startCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return errMsg{err}
		}
		if err := service.Start(); err != nil {
			return errMsg{err}
		}
		return startedMsg{}
	}
}

func stopCmd() tea.Cmd {
	return func() tea.Msg {
		if err := service.Stop(); err != nil {
			return errMsg{err}
		}
		return stoppedMsg{}
	}
}

func saveCmd(cfg config.Config, running bool) tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return errMsg{err}
		}
		if running {
			if err := service.Start(); err != nil {
				return errMsg{err}
			}
			return startedMsg{}
		}
		return savedMsg{}
	}
}

// Run opens the setup TUI.
func Run() error {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
