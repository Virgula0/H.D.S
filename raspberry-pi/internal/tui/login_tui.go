package tui

// https://github.com/charmbracelet/bubbletea/blob/main/examples/textinput/main.go

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Login ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Login"))
)

const (
	ctrlC    = "ctrl+c"
	ctrlR    = "ctrl+r"
	esc      = "esc"
	tab      = "tab"
	shiftTab = "shift+tab"
	enter    = "enter"
	up       = "up"
	down     = "down"
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

func LoginModel() model {
	m := model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.CharLimit = 64
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) GetCredentials() (username, password string) {
	return m.inputs[0].Value(), m.inputs[1].Value()
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) changeCursorMode() (tea.Model, tea.Cmd) {
	m.cursorMode++
	if m.cursorMode > cursor.CursorHide {
		m.cursorMode = cursor.CursorBlink
	}
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
	}
	return m, tea.Batch(cmds...)
}

func (m model) focusNextInput(s string) (tea.Model, tea.Cmd) {

	// Did the user press enter while the submit button was focused?
	// If so, exit.
	if s == enter && m.focusIndex == len(m.inputs) {
		return m, tea.Quit
	}

	// Cycle indexes
	if s == up || s == shiftTab {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.inputs) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs)
	}

	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}

	return m, tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case ctrlC, esc:
			return m, tea.Quit

		// Change cursor mode
		case ctrlR:
			return m.changeCursorMode()

		// Set focus to next input
		case tab, shiftTab, enter, up, down:
			s := msg.String()
			return m.focusNextInput(s)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}

	b.WriteString(fmt.Sprintf("\n\n%s\n\n", *button))

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
