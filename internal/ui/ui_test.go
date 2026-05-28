package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ReubenPercival/flux/internal/monitor"
)

func TestNewModel(t *testing.T) {
	mon := monitor.NewMonitor()
	m := NewModel(mon)
	if m.monitor == nil {
		t.Fatal("monitor should not be nil")
	}
}

func TestModelInit(t *testing.T) {
	mon := monitor.NewMonitor()
	m := NewModel(mon)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModelView(t *testing.T) {
	mon := monitor.NewMonitor()
	m := NewModel(mon)
	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModelQuit(t *testing.T) {
	mon := monitor.NewMonitor()
	m := NewModel(mon)
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	model, cmd := m.Update(keyMsg)
	if cmd == nil {
		t.Error("expected tea.Quit cmd on Ctrl+C")
	}
	if _, ok := model.(Model); !ok {
		t.Fatal("expected Model type")
	}
}
