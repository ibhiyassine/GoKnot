package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ibhiyassine/GoKnot/internal/domain"
	"github.com/ibhiyassine/GoKnot/internal/loadbalancer"
)

/* Model Definition
 */
type focus int

const (
	backendFocused focus = iota
	actionFocused
)

type availableActions string

const (
	addBackend    availableActions = "Add Backend"
	removeBackend availableActions = "Remove Backend"
)

type backendModel struct {
	lb            loadbalancer.LoadBalancer
	backends      []*domain.Backend
	backendCursor int
}

type adminActionsModel struct {
	adminPort    int                // For sending the requests of the admin
	actions      []availableActions // Action possible as an admin
	actionCursor int
}

type popupInput struct {
	showPopup bool
	input     textinput.Model
}

type Model struct {
	backendModel
	adminActionsModel
	popupInput
	feedbackMsg    string
	focusedSection focus
}

func InitialModel(lb loadbalancer.LoadBalancer, adminPort int) Model {

	textInput := textinput.New()
	textInput.Placeholder = "Put the backend URL"
	textInput.CharLimit = 256
	textInput.Width = 40

	return Model{
		backendModel: backendModel{
			backends: lb.GetBackends(),
			lb:       lb,
		},
		adminActionsModel: adminActionsModel{
			actions:   []availableActions{addBackend, removeBackend},
			adminPort: adminPort,
		},
		popupInput: popupInput{
			showPopup: false,
			input:     textInput,
		},
		focusedSection: backendFocused,
	}
}

func (m Model) Init() tea.Cmd {

	// We want the TUI to refresh automatically every 500 ms
	return tickCmd()
}

// Helper types
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// =========================================================================
	// 1. GLOBAL EVENTS (Must happen regardless of popup state)
	// =========================================================================
	switch msg := msg.(type) {

	// The Heartbeat: Always update data and restart timer
	case tickMsg:
		m.backends = m.lb.GetBackends()
		// Safety check for cursor bounds
		if m.backendCursor >= len(m.backends) && len(m.backends) > 0 {
			m.backendCursor = len(m.backends) - 1
		}
		return m, tickCmd()

	// The API Feedback: Always show success/error
	case apiResultMsg:
		if msg.err != nil {
			m.feedbackMsg = "Error: " + msg.err.Error()
		} else {
			m.feedbackMsg = "Success: " + msg.message
			// FORCE REFRESH: Immediately update the list so we don't have to wait 500ms
			m.backends = m.lb.GetBackends()
		}
	}

	// =========================================================================
	// 2. POPUP MODE (Hijack keyboard)
	// =========================================================================
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEsc:
				m.showPopup = false
				m.input.Blur()
				m.input.Reset()
				return m, nil
			case tea.KeyEnter:
				url := m.input.Value()
				m.showPopup = false
				m.input.Blur()
				m.input.Reset()
				return m, m.addBackendCmd(url)
			}
		}
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	// =========================================================================
	// 3. DASHBOARD MODE (Normal Navigation)
	// =========================================================================
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyboardInput(msg.String())
	}

	return m, nil
}

func (m Model) handleKeyboardInput(key string) (Model, tea.Cmd) {
	switch key {
	case "esc", "q":
		return m, tea.Quit

	case "tab":
		if m.focusedSection == actionFocused {
			m.focusedSection = backendFocused
		} else if m.focusedSection == backendFocused {
			m.focusedSection = actionFocused
		}

	case "enter":
		if m.focusedSection == actionFocused {
			return m.executeMenuAction()
		}

	// Navigation
	case "up", "k":
		if m.focusedSection == backendFocused && m.backendCursor > 0 {
			m.backendCursor--
		}
	case "down", "j":
		if m.focusedSection == backendFocused && m.backendCursor < len(m.backends)-1 {
			m.backendCursor++
		}
	case "left", "h":
		if m.focusedSection == actionFocused && m.actionCursor > 0 {
			m.actionCursor--
		}
	case "right", "l":
		if m.focusedSection == actionFocused && m.actionCursor < len(m.actions)-1 {
			m.actionCursor++
		}
	}
	return m, nil

}

func (m Model) executeMenuAction() (Model, tea.Cmd) {
	selectedAction := m.actions[m.actionCursor]
	switch selectedAction {
	case addBackend:
		m.showPopup = true
		m.input.Focus()
		return m, textinput.Blink
	case removeBackend:
		// We will remove the selected backend from the table
		if len(m.backends) > 0 {
			targetURL := m.backends[m.backendCursor].URL.String()
			// Delete
			return m, m.deleteBackendCmd(targetURL)

		}
	}

	return m, nil

}

type apiResultMsg struct {
	message string
	err     error
}

func (m Model) addBackendCmd(url string) tea.Cmd {
	return func() tea.Msg {
		request := map[string]string{"url": url}
		jsonBody, err := json.Marshal(request)
		if err != nil {
			return apiResultMsg{err: err}
		}

		// Send the post request to the admin API
		adminEndpoint := fmt.Sprintf("http://localhost:%d/backends", m.adminPort)
		resp, err := http.Post(adminEndpoint, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return apiResultMsg{err: err}
		}
		defer resp.Body.Close()

		return apiResultMsg{message: "Added " + url}
	}
}

func (m Model) deleteBackendCmd(url string) tea.Cmd {
	return func() tea.Msg {
		// prepare JSON
		request := map[string]string{"url": url}

		jsonBody, err := json.Marshal(request)
		if err != nil {
			return apiResultMsg{err: err}
		}

		adminEndpoint := fmt.Sprintf("http://localhost:%d/backends", m.adminPort)
		req, err := http.NewRequest(http.MethodDelete, adminEndpoint, bytes.NewBuffer(jsonBody))

		if err != nil {
			return apiResultMsg{err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			return apiResultMsg{err: err}
		}
		defer resp.Body.Close()

		return apiResultMsg{message: "Removed " + url}
	}
}

// NOTE: This code is generated by AI for now because i don't understant anything about UI/UX and designing
// Of course i will understand the code
// =============================================================================
// STYLES (Simple & Functional)
// =============================================================================
var (
	// Layout
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// Colors
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#7D56F4")). // Purple header
			Padding(0, 1)

	// The "Selected" state
	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // Pink/Magenta for focus
			Bold(true)

	// The "Unselected" state
	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Grey

	// Status indicators
	statusAlive = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
	statusDead  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red
)

// =============================================================================
// THE VIEW
// =============================================================================
func (m Model) View() string {
	// 1. POPUP MODE (Overlay)
	if m.showPopup {
		return appStyle.Render(fmt.Sprintf(
			"\n\n%s\n\n%s\n\n%s",
			titleStyle.Render(" ADD BACKEND "),
			"Enter URL (e.g. http://localhost:9003):",
			m.input.View(),
		) + "\n\n(Esc to Cancel)")
	}

	// 2. MAIN DASHBOARD
	var s strings.Builder

	// -- Header --
	s.WriteString(titleStyle.Render("GoKnot Admin ") + "\n\n")

	// -- Section A: Backends List --
	s.WriteString("BACKENDS:\n")

	// Header Row
	s.WriteString(fmt.Sprintf("  %-30s | %-10s | %s\n", "URL", "Status", "Conns"))
	s.WriteString("  ------------------------------------------------------------\n")

	if len(m.backends) == 0 {
		s.WriteString("  (No backends found)\n")
	}

	for i, b := range m.backends {
		// Determine styling for this row
		var cursor string
		if m.backendCursor == i {
			cursor = "> " // The cursor pointer
		} else {
			cursor = "  " // default empty
		}
		rowStyle := blurredStyle

		// If Backend section is focused
		if m.focusedSection == backendFocused {
			if m.backendCursor == i {
				rowStyle = focusedStyle
			} else {
				rowStyle = lipgloss.NewStyle() // Normal white text
			}
		}

		// Status coloring
		status := "ALIVE"
		stStyle := statusAlive
		if !b.IsAlive() {
			status = "DEAD"
			stStyle = statusDead
		}

		// Render the row
		s.WriteString(fmt.Sprintf("%s%s | %s | %d\n",
			rowStyle.Render(cursor),
			rowStyle.Render(fmt.Sprintf("%-30s", b.URL.String())),
			stStyle.Render(fmt.Sprintf("%-10s", status)),
			b.CurrentConns,
		))
	}
	s.WriteString("\n")

	// -- Section B: Action Menu --
	s.WriteString("ACTIONS:\n")
	var actionsView strings.Builder
	for i, action := range m.actions {
		// Styling logic for buttons
		style := blurredStyle
		prefix := "[ ] "

		if m.focusedSection == actionFocused {
			if m.actionCursor == i {
				style = focusedStyle // Highlight active button
				prefix = "[x] "
			} else {
				style = lipgloss.NewStyle() // Normal text
			}
		}

		actionsView.WriteString(style.Render(prefix+string(action)) + "   ")
	}
	s.WriteString("  " + actionsView.String() + "\n\n")

	// -- Section C: Feedback Bar --
	if m.feedbackMsg != "" {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render("LOG: " + m.feedbackMsg))
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Ready."))
	}

	s.WriteString("\n\n(Tab to switch focus, q to quit)")

	return appStyle.Render(s.String())
}
