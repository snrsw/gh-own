package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		url         string
	}{
		{
			name:        "basic item",
			title:       "Test Title",
			description: "Test Description",
			url:         "https://example.com",
		},
		{
			name:        "empty values",
			title:       "",
			description: "",
			url:         "",
		},
		{
			name:        "searchable item",
			title:       "Searchable Title",
			description: "Description",
			url:         "url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.title, tt.description, tt.url)

			if got := item.Title(); got != tt.title {
				t.Errorf("Title() = %q, want %q", got, tt.title)
			}
			if got := item.Description(); got != tt.description {
				t.Errorf("Description() = %q, want %q", got, tt.description)
			}
			if got := item.FilterValue(); got != tt.title {
				t.Errorf("FilterValue() = %q, want %q", got, tt.title)
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
				NewItem("Item 1", "Desc 1", "url1"),
				NewItem("Item 2", "Desc 2", "url2"),
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
		expectedTab    int
		expectsCommand bool
	}{
		{
			name:           "tab cycles to next",
			key:            tea.KeyMsg{Type: tea.KeyTab},
			expectedTab:    1,
			expectsCommand: false,
		},
		{
			name:           "ctrl+c quits",
			key:            tea.KeyMsg{Type: tea.KeyCtrlC},
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

			newModel, cmd := m.Update(tt.key)
			m = newModel.(Model)

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
		m = newModel.(Model)
	}

	if m.activeTab != 0 {
		t.Errorf("after wrapping, activeTab = %d, want 0", m.activeTab)
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
			m = newModel.(Model)

			if m.width != tt.width {
				t.Errorf("width = %d, want %d", m.width, tt.width)
			}
			if m.height != tt.height {
				t.Errorf("height = %d, want %d", m.height, tt.height)
			}
		})
	}
}

func TestModel_View(t *testing.T) {
	m := NewModel([]Tab{NewTab("Test Tab", CreateList(nil))})

	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	if view := m.View(); view == "" {
		t.Error("View() should not be empty")
	}
}
