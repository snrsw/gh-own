package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name          string
		repoName      string
		titleText     string
		description   string
		url           string
		wantTitle     string
	}{
		{
			name:        "with repo name",
			repoName:    "owner/repo",
			titleText:   "Test Title",
			description: "Test Description",
			url:         "https://example.com",
			wantTitle:   "owner/repo Test Title",
		},
		{
			name:        "empty values",
			repoName:    "",
			titleText:   "",
			description: "",
			url:         "",
			wantTitle:   "",
		},
		{
			name:        "without repo name",
			repoName:    "",
			titleText:   "Searchable Title",
			description: "Description",
			url:         "url",
			wantTitle:   "Searchable Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.repoName, tt.titleText, tt.description, tt.url)

			if got := item.Title(); got != tt.wantTitle {
				t.Errorf("Title() = %q, want %q", got, tt.wantTitle)
			}
			if got := item.Description(); got != tt.description {
				t.Errorf("Description() = %q, want %q", got, tt.description)
			}
			if got := item.FilterValue(); got != tt.wantTitle {
				t.Errorf("FilterValue() = %q, want %q", got, tt.wantTitle)
			}
		})
	}
}

func TestCreateList(t *testing.T) {
	tests := []struct {
		name     string
		items    []list.Item
		expected int
	}{
		{
			name:     "nil items",
			items:    nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			items:    []list.Item{},
			expected: 0,
		},
		{
			name: "two items",
			items: []list.Item{
				NewItem("", "Item 1", "Desc 1", "url1"),
				NewItem("", "Item 2", "Desc 2", "url2"),
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := CreateList(tt.items)
			if got := len(l.Items()); got != tt.expected {
				t.Errorf("len(Items()) = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestNewTab(t *testing.T) {
	l := CreateList(nil)
	tab := NewTab("Test Tab", l)

	if tab.name != "Test Tab" {
		t.Errorf("Tab.name = %q, want %q", tab.name, "Test Tab")
	}
}

func TestNewModel(t *testing.T) {
	tests := []struct {
		name            string
		tabs            []Tab
		expectedTabs    int
		expectedActive  int
		firstTabName    string
	}{
		{
			name:            "nil tabs creates default",
			tabs:            nil,
			expectedTabs:    1,
			expectedActive:  0,
			firstTabName:    "Empty",
		},
		{
			name:            "empty slice creates default",
			tabs:            []Tab{},
			expectedTabs:    1,
			expectedActive:  0,
			firstTabName:    "Empty",
		},
		{
			name: "two tabs",
			tabs: []Tab{
				NewTab("Tab 1", CreateList(nil)),
				NewTab("Tab 2", CreateList(nil)),
			},
			expectedTabs:   2,
			expectedActive: 0,
			firstTabName:   "Tab 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(tt.tabs)

			if got := len(m.tabs); got != tt.expectedTabs {
				t.Errorf("len(tabs) = %d, want %d", got, tt.expectedTabs)
			}
			if m.activeTab != tt.expectedActive {
				t.Errorf("activeTab = %d, want %d", m.activeTab, tt.expectedActive)
			}
			if m.tabs[0].name != tt.firstTabName {
				t.Errorf("tabs[0].name = %q, want %q", m.tabs[0].name, tt.firstTabName)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel(nil)
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_KeyMsg(t *testing.T) {
	tests := []struct {
		name           string
		key            tea.KeyMsg
		startTab       int
		expectedTab    int
		expectsCommand bool
	}{
		{
			name:           "tab cycles to next",
			key:            tea.KeyMsg{Type: tea.KeyTab},
			startTab:       0,
			expectedTab:    1,
			expectsCommand: false,
		},
		{
			name:           "shift+tab cycles to previous",
			key:            tea.KeyMsg{Type: tea.KeyShiftTab},
			startTab:       1,
			expectedTab:    0,
			expectsCommand: false,
		},
		{
			name:           "ctrl+c quits",
			key:            tea.KeyMsg{Type: tea.KeyCtrlC},
			startTab:       0,
			expectedTab:    0,
			expectsCommand: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel([]Tab{
				NewTab("Tab 1", CreateList(nil)),
				NewTab("Tab 2", CreateList(nil)),
			})
			m.activeTab = tt.startTab

			newModel, cmd := m.Update(tt.key)
			var ok bool
			m, ok = newModel.(Model)
			if !ok {
				t.Fatal("expected Model type")
			}

			if m.activeTab != tt.expectedTab {
				t.Errorf("activeTab = %d, want %d", m.activeTab, tt.expectedTab)
			}
			if tt.expectsCommand && cmd == nil {
				t.Error("expected command, got nil")
			}
			if !tt.expectsCommand && cmd != nil {
				t.Error("expected nil command")
			}
		})
	}
}

func TestModel_Update_TabWrap(t *testing.T) {
	m := NewModel([]Tab{
		NewTab("Tab 1", CreateList(nil)),
		NewTab("Tab 2", CreateList(nil)),
		NewTab("Tab 3", CreateList(nil)),
	})

	keyMsg := tea.KeyMsg{Type: tea.KeyTab}

	// Press tab 3 times to wrap around
	for i := 0; i < 3; i++ {
		newModel, _ := m.Update(keyMsg)
		var ok bool
		m, ok = newModel.(Model)
		if !ok {
			t.Fatal("expected Model type")
		}
	}

	if m.activeTab != 0 {
		t.Errorf("after wrapping, activeTab = %d, want 0", m.activeTab)
	}
}

func TestModel_Update_ShiftTabWrap(t *testing.T) {
	m := NewModel([]Tab{
		NewTab("Tab 1", CreateList(nil)),
		NewTab("Tab 2", CreateList(nil)),
		NewTab("Tab 3", CreateList(nil)),
	})

	// activeTab starts at 0; shift+tab should wrap to last tab (2)
	keyMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ := m.Update(keyMsg)
	var ok bool
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	if m.activeTab != 2 {
		t.Errorf("after shift+tab wrap, activeTab = %d, want 2", m.activeTab)
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard terminal", 80, 24},
		{"wide terminal", 200, 50},
		{"small terminal", 40, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel([]Tab{NewTab("Tab", CreateList(nil))})

			sizeMsg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			newModel, _ := m.Update(sizeMsg)
			var ok bool
			m, ok = newModel.(Model)
			if !ok {
				t.Fatal("expected Model type")
			}

			if m.width != tt.width {
				t.Errorf("width = %d, want %d", m.width, tt.width)
			}
			if m.height != tt.height {
				t.Errorf("height = %d, want %d", m.height, tt.height)
			}
		})
	}
}

func TestNewLoadingModel(t *testing.T) {
	m := NewLoadingModel(nil)

	if !m.loading {
		t.Error("NewLoadingModel() should have loading = true")
	}

	if cmd := m.Init(); cmd == nil {
		t.Error("Init() should return a non-nil command for spinner tick")
	}

	// Give it a size so View() can render
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	var ok bool
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View() in loading state should contain 'Loading', got %q", view)
	}
}

func TestModel_Update_SpinnerTick(t *testing.T) {
	t.Run("loading state forwards tick", func(t *testing.T) {
		m := NewLoadingModel(nil)
		tick := spinner.TickMsg{ID: m.spinner.ID()}

		newModel, cmd := m.Update(tick)
		var ok bool
		m, ok = newModel.(Model)
		if !ok {
			t.Fatal("expected Model type")
		}

		if cmd == nil {
			t.Error("in loading state, spinner tick should return a command")
		}
	})

	t.Run("loaded state ignores tick", func(t *testing.T) {
		m := NewModel([]Tab{
			NewTab("Tab 1", CreateList(nil)),
		})
		tick := spinner.TickMsg{}

		_, cmd := m.Update(tick)

		if cmd != nil {
			t.Error("in loaded state, spinner tick should return nil command")
		}
	})
}

func TestModel_Update_TabsMsg(t *testing.T) {
	m := NewLoadingModel(nil)

	tabs := []Tab{
		NewTab("Created (3)", CreateList(nil)),
		NewTab("Assigned (1)", CreateList(nil)),
	}

	newModel, _ := m.Update(TabsMsg(tabs))
	var ok bool
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	if m.loading {
		t.Error("after TabsMsg, loading should be false")
	}
	if len(m.tabs) != 2 {
		t.Errorf("tabs count = %d, want 2", len(m.tabs))
	}
	if m.tabs[0].name != "Created (3)" {
		t.Errorf("tabs[0].name = %q, want %q", m.tabs[0].name, "Created (3)")
	}

	// View should render tabs, not spinner
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ = m.Update(sizeMsg)
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	view := m.View()
	if strings.Contains(view, "Loading") {
		t.Error("after TabsMsg, View() should not contain 'Loading'")
	}
}

func TestModel_Update_ErrMsg(t *testing.T) {
	m := NewLoadingModel(nil)

	errMsg := ErrMsg{Err: errors.New("API timeout")}
	_, cmd := m.Update(errMsg)

	if cmd == nil {
		t.Fatal("ErrMsg should return a command (tea.Quit)")
	}
}

func TestFetchCmd(t *testing.T) {
	t.Run("success returns TabsMsg", func(t *testing.T) {
		tabs := []Tab{NewTab("Tab 1", CreateList(nil))}
		cmd := FetchCmd(func() ([]Tab, error) {
			return tabs, nil
		})

		msg := cmd()
		tabsMsg, ok := msg.(TabsMsg)
		if !ok {
			t.Fatalf("expected TabsMsg, got %T", msg)
		}
		if len(tabsMsg) != 1 {
			t.Errorf("TabsMsg len = %d, want 1", len(tabsMsg))
		}
	})

	t.Run("failure returns ErrMsg", func(t *testing.T) {
		cmd := FetchCmd(func() ([]Tab, error) {
			return nil, errors.New("network error")
		})

		msg := cmd()
		errMsg, ok := msg.(ErrMsg)
		if !ok {
			t.Fatalf("expected ErrMsg, got %T", msg)
		}
		if errMsg.Err.Error() != "network error" {
			t.Errorf("ErrMsg.Err = %q, want %q", errMsg.Err.Error(), "network error")
		}
	})
}

func TestModel_WindowResize_DuringLoading(t *testing.T) {
	m := NewLoadingModel(nil)

	// Resize while loading
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(sizeMsg)
	var ok bool
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m.width, m.height)
	}

	// Transition to loaded
	tabs := []Tab{NewTab("Tab 1", CreateList(nil))}
	newModel, _ = m.Update(TabsMsg(tabs))
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	// Tabs should have sizes applied from the earlier resize
	if m.outerW == 0 {
		t.Error("after transition, outerW should be set from earlier resize")
	}
}

func TestModel_Update_RefreshKey(t *testing.T) {
	fetch := FetchCmd(func() ([]Tab, error) {
		return []Tab{NewTab("Refreshed", CreateList(nil))}, nil
	})

	m := NewLoadingModel(fetch)
	// Transition to loaded state
	tabs := []Tab{
		NewTab("Tab 1", CreateList(nil)),
		NewTab("Tab 2", CreateList(nil)),
	}
	newModel, _ := m.Update(TabsMsg(tabs))
	m = newModel.(Model)

	// Press 'r' to refresh
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	if !m.loading {
		t.Error("after pressing 'r', loading should be true")
	}
	if cmd == nil {
		t.Error("after pressing 'r', command should not be nil")
	}
}

func TestModel_Update_RefreshKey_IgnoredDuringLoading(t *testing.T) {
	fetch := FetchCmd(func() ([]Tab, error) {
		return []Tab{NewTab("Tab", CreateList(nil))}, nil
	})
	m := NewLoadingModel(fetch)

	// Press 'r' while still loading
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	if !m.loading {
		t.Error("loading should remain true")
	}
	if cmd != nil {
		t.Error("no command should be returned when refreshing during loading")
	}
}

func TestModel_View(t *testing.T) {
	m := NewModel([]Tab{NewTab("Test Tab", CreateList(nil))})

	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	var ok bool
	m, ok = newModel.(Model)
	if !ok {
		t.Fatal("expected Model type")
	}

	if view := m.View(); view == "" {
		t.Error("View() should not be empty")
	}
}
