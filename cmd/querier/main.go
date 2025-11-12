package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#5A4FCF"}).
			Padding(0, 1).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#02A96F", Dark: "#04B575"}).
			Bold(true)

	sessionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#999999", Dark: "#626262"}).
			Italic(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#6B46C1", Dark: "#874BFD"}).
			Padding(0, 1)

	responseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#04B575"}).
			Padding(1, 2).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FF5733"}).
			Bold(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#874BFD"})

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#626262"}).
			Italic(true)

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#EC4899", Dark: "#FF75B7"}).
			Padding(0, 1)
)

type Request struct {
	Query string `json:"query"`
}

type Response struct {
	Data string   `json:"data"`
	Refs []string `json:"refs"`
}

type responseMsg struct {
	response string
	err      error
}

type model struct {
	sessionID string
	orgID     string
	textarea  textarea.Model
	viewport  viewport.Model
	spinner   spinner.Model
	loading   bool
	responses []string
	err       error
	width     int
	height    int
	focused   bool
}

func initialModel() model {
	// Generate session ID
	sessionID := uuid.New().String()
	orgID := "11111111-1111-1111-1111-111111111111"

	// Initialize textarea
	ta := textarea.New()
	ta.Placeholder = "Ask a question..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 1000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	// Initialize viewport
	vp := viewport.New(80, 20)

	// Add welcome message
	welcomeMsg := responseStyle.Render("How can I help you today?")
	vp.SetContent(welcomeMsg)

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return model{
		sessionID: sessionID,
		orgID:     orgID,
		textarea:  ta,
		viewport:  vp,
		spinner:   s,
		loading:   false,
		responses: []string{welcomeMsg},
		focused:   true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes
		m.textarea.SetWidth(msg.Width - 4)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 12 // Leave room for header, input, and help

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.loading {
				return m, nil
			}

			query := strings.TrimSpace(m.textarea.Value())
			if query == "" {
				return m, nil
			}

			// Add user query to conversation
			userQuery := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFB86C")).
				Bold(true).
				Render("❯ " + query)
			m.responses = append(m.responses, userQuery)
			m.updateViewport()

			// Clear textarea and start loading
			m.textarea.Reset()
			m.loading = true

			// Make the API request
			return m, tea.Batch(
				m.spinner.Tick,
				m.makeRequest(query),
			)

		case tea.KeyTab:
			m.focused = !m.focused
			if m.focused {
				m.textarea.Focus()
			} else {
				m.textarea.Blur()
			}
		case tea.KeyUp:
			m.viewport, vpCmd = m.viewport.Update(msg)
		case tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			errorText := errorStyle.Render("❌ Error: " + msg.err.Error())
			m.responses = append(m.responses, errorText)
		} else {
			m.err = nil
			responseText := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#04B575")).
				Padding(1, 2).
				MarginTop(1).Width(m.viewport.Width - 2).
				Render("🤖 " + msg.response)
			m.responses = append(m.responses, responseText)
		}
		m.updateViewport()
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}
	case spinner.TickMsg:
		if m.loading {
			return m, spCmd
		}
	}

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m model) View() string {
	// Header
	title := titleStyle.Render("🤖 MCP Chat Interface")
	session := sessionStyle.Render(fmt.Sprintf("Session: %s", m.sessionID[:8]+"..."))
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", session)

	// Input section
	var inputSection string
	if m.focused {
		inputSection = focusedStyle.Render(m.textarea.View())
	} else {
		inputSection = inputStyle.Render(m.textarea.View())
	}

	// Loading indicator
	var loadingIndicator string
	if m.loading {
		loadingIndicator = spinnerStyle.Render(fmt.Sprintf("%s Querying...", m.spinner.View()))
	}

	// Help text
	help := helpStyle.Render("💡 Enter to send • ↑↓ to scroll • Ctrl+C to quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		headerStyle.Render("💬 Conversation:"),
		m.viewport.View(),
		"",
		headerStyle.Render("✏️  Your Question:"),
		inputSection,
		loadingIndicator,
		"",
		help,
	)

	return content
}

func (m *model) updateViewport() {
	content := strings.Join(m.responses, "\n\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m model) makeRequest(query string) tea.Cmd {
	return func() tea.Msg {
		// Create request payload
		reqData := Request{Query: query}
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return responseMsg{err: fmt.Errorf("marshaling request: %v", err)}
		}

		// Make POST request
		url := fmt.Sprintf("http://localhost:8100/ask?organization_id=%s&session_id=%s", m.orgID, m.sessionID)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return responseMsg{err: fmt.Errorf("making request: %v", err)}
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return responseMsg{err: fmt.Errorf("reading response: %v", err)}
		}

		if resp.StatusCode != http.StatusOK {
			return responseMsg{err: fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))}
		}

		// Parse response
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			return responseMsg{err: fmt.Errorf("parsing response: %v", err)}
		}

		responseText := strings.TrimSpace(response.Data)
		responseText = strings.ReplaceAll(responseText, "\\n", "\n")
		if len(response.Refs) > 0 {
			responseText = fmt.Sprintf("%s\n\nRead More:\n• %s",
				responseText,
				strings.Join(response.Refs, "\n• "),
			)
		}

		return responseMsg{response: responseText}
	}
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
