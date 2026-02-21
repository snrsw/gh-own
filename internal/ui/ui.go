// Package ui provides the user interface components for the application.
package ui

import (
	"log/slog"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	repoName, titleText, description, url string
}

func NewItem(repoName, titleText, description, url string) Item {
	return Item{
		repoName:    repoName,
		titleText:   titleText,
		description: description,
		url:         url,
	}
}

func (i Item) Title() string {
	if i.repoName == "" {
		return i.titleText
	}
	return i.repoName + " " + i.titleText
}
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.Title() }

func CreateList(items []list.Item) list.Model {
	delegate := newGithubDelegate()
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	configureHelp(&l)
	return l
}

type Tab struct {
	name string
	list list.Model
}

func NewTab(name string, list list.Model) Tab {
	return Tab{
		name: name,
		list: list,
	}
}

type Model struct {
	tabs      []Tab
	activeTab int
	width     int
	height    int
	outerW    int
	outerH    int
	loading   bool
	spinner   spinner.Model
	err       error
	fetchCmd  tea.Cmd
}

// TabsMsg signals that data loading is complete and tabs are ready.
type TabsMsg []Tab

// ErrMsg signals that data loading failed.
type ErrMsg struct{ Err error }

func NewModel(tabs []Tab) Model {
	if len(tabs) == 0 {
		tabs = []Tab{NewTab("Empty", CreateList(nil))}
	}
	return Model{
		tabs:      tabs,
		activeTab: 0,
	}
}

func NewLoadingModel(fetch tea.Cmd) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorAccent)
	return Model{
		loading:  true,
		spinner:  s,
		tabs:     []Tab{NewTab("Empty", CreateList(nil))},
		fetchCmd: fetch,
	}
}

// FetchCmd wraps a data-fetching function into a tea.Cmd.
// On success it returns TabsMsg; on failure it returns ErrMsg.
func FetchCmd(fn func() ([]Tab, error)) tea.Cmd {
	return func() tea.Msg {
		tabs, err := fn()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return TabsMsg(tabs)
	}
}

func (m Model) Init() tea.Cmd {
	if m.loading {
		return tea.Batch(m.spinner.Tick, m.fetchCmd)
	}
	return nil
}

// Err returns the error from a failed fetch, if any.
func (m Model) Err() error {
	return m.err
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), nil
	case tea.KeyMsg:
		if mm, cmd, handled := m.handleKey(msg); handled {
			return mm, cmd
		}
	case ErrMsg:
		m.err = msg.Err
		return m, tea.Quit
	case TabsMsg:
		m.loading = false
		m.tabs = []Tab(msg)
		if len(m.tabs) == 0 {
			m.tabs = []Tab{NewTab("Empty", CreateList(nil))}
		}
		m.activeTab = 0
		if m.width > 0 {
			m = m.handleWindowSize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.tabs[m.activeTab].list, cmd = m.tabs[m.activeTab].list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.loading {
		return DocStyle.Render(m.spinner.View() + " Loading...")
	}

	var doc strings.Builder

	doc.WriteString(m.tabsView())

	doc.WriteString(
		WindowStyle.
			Width(m.outerW).
			Height(m.outerH).
			Render(m.tabs[m.activeTab].list.View()),
	)

	doc.WriteString("\n")
	doc.WriteString(helpView())

	out := DocStyle.Render(doc.String())
	return out
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}

		if err := cmd.Start(); err != nil {
			slog.Error("failed to open URL", "url", url, "error", err)
		}
		return nil
	}
}

func (m Model) tabsView() string {
	activeStyle, inactiveStyle := GithubTabStyles()

	var tabs []string
	for i, t := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, activeStyle.Render(t.name))
		} else {
			tabs = append(tabs, inactiveStyle.Render(t.name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	line := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#D0D7DE", Dark: "#30363D"}).
		Render(strings.Repeat("─", m.outerW))

	return lipgloss.JoinVertical(lipgloss.Left, row, line)
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) Model {
	m.width, m.height = msg.Width, msg.Height

	docH, docV := DocStyle.GetFrameSize()
	winH, winV := WindowStyle.GetFrameSize()

	m.outerW = max(20, m.width-docH)
	innerW := max(20, m.outerW-winH)

	tabsH := lipgloss.Height(m.tabsView())
	helpH := lipgloss.Height(helpView()) + 1 // +1 for newline
	m.outerH = max(5, m.height-docV-tabsH-helpH)

	innerH := max(5, m.outerH-winV)

	for i := range m.tabs {
		m.tabs[i].list.SetSize(innerW, innerH)
	}
	return m
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit, true

	case "tab":
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
		return m, nil, true

	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		return m, nil, true

	case "enter":
		return m.handleEnter()

	case "r":
		return m.handleRefresh()
	}
	return m, nil, false
}

func (m Model) handleRefresh() (Model, tea.Cmd, bool) {
	if m.loading || m.fetchCmd == nil {
		return m, nil, true
	}
	m.loading = true
	return m, tea.Batch(m.spinner.Tick, m.fetchCmd), true
}

func (m Model) handleEnter() (Model, tea.Cmd, bool) {
	if m.tabs[m.activeTab].list.FilterState() == list.Filtering {
		return m, nil, true
	}

	sel := m.tabs[m.activeTab].list.SelectedItem()
	it, ok := sel.(Item)
	if !ok || it.url == "" {
		return m, nil, true
	}

	return m, openURLCmd(it.url), true
}
