package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

func sampleObjects() []backend.Object {
	return []backend.Object{
		{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
		{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
		{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit"},
		{ID: 4, Path: "GARMIN/Activity/b.fit", Name: "b.fit"},
		{ID: 5, Path: "Music", Name: "Music", IsDir: true},
	}
}

func loadedModel(t *testing.T) Model {
	t.Helper()
	m := NewModel(context.Background(), &backend.FakeBackend{Objects: sampleObjects()}, "")
	next, _ := m.Update(objectsMsg{sampleObjects()})
	return next.(Model)
}

func update(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestInitLoads(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{Objects: sampleObjects()}, "")
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init returned nil cmd")
	}
	if _, ok := cmd().(objectsMsg); !ok {
		t.Fatalf("expected objectsMsg, got %T", cmd())
	}
}

func TestInitError(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{ListErr: errors.New("no device")}, "")
	if _, ok := m.Init()().(errMsg); !ok {
		t.Fatal("expected errMsg on load failure")
	}
}

func TestUpdateObjects(t *testing.T) {
	m := loadedModel(t)
	if m.loading {
		t.Error("loading should be false after objectsMsg")
	}
	if len(m.entries()) != 2 { // GARMIN, Music at root
		t.Errorf("root entries = %d, want 2", len(m.entries()))
	}
}

func TestUpdateError(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	m, _ = update(t, m, errMsg{errors.New("boom")})
	if m.err == nil || m.loading {
		t.Fatalf("err=%v loading=%v", m.err, m.loading)
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if m.height != 24 {
		t.Errorf("height = %d", m.height)
	}
}

func TestUpdateUnknownMsg(t *testing.T) {
	m := loadedModel(t)
	next, cmd := update(t, m, struct{}{})
	if cmd != nil {
		t.Error("unknown msg should produce no command")
	}
	if len(next.entries()) != 2 {
		t.Error("unknown msg should not change state")
	}
}

func TestNavigateDownUp(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("cursor after down = %d", m.cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown}) // at last, clamp
	if m.cursor != 1 {
		t.Fatalf("cursor should clamp at last, got %d", m.cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp}) // at first, clamp
	if m.cursor != 0 {
		t.Fatalf("cursor should clamp at 0, got %d", m.cursor)
	}
}

func TestEnterDirAndBack(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // into GARMIN
	if m.cwd != "GARMIN" {
		t.Fatalf("cwd = %q", m.cwd)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyBackspace}) // back to root
	if m.cwd != "" {
		t.Fatalf("cwd after back = %q", m.cwd)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyBackspace}) // at root, no-op
	if m.cwd != "" {
		t.Fatalf("back at root should stay, got %q", m.cwd)
	}
}

func TestEnterOnFileNoop(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // a.fit is a file
	if m.cwd != "GARMIN/Activity" {
		t.Fatalf("entering a file changed cwd to %q", m.cwd)
	}
}

func TestSelectToggle(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace}) // select a.fit (id 3)
	if !m.selected[3] {
		t.Fatal("a.fit should be selected")
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace}) // deselect
	if m.selected[3] {
		t.Fatal("a.fit should be deselected")
	}
}

func TestSelectDirNoop(t *testing.T) {
	m := loadedModel(t) // cursor on GARMIN (a dir)
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace})
	if len(m.selected) != 0 {
		t.Fatal("directories must not be selectable")
	}
}

func TestQuit(t *testing.T) {
	m := loadedModel(t)
	for _, k := range []tea.KeyMsg{{Type: tea.KeyCtrlC}, runes("q")} {
		_, cmd := update(t, m, k)
		if cmd == nil {
			t.Fatalf("%v should quit", k)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%v did not return QuitMsg", k)
		}
	}
}

func TestRefresh(t *testing.T) {
	m := loadedModel(t)
	next, cmd := update(t, m, runes("r"))
	if !next.loading || cmd == nil {
		t.Fatal("refresh should set loading and return a command")
	}
	if _, ok := cmd().(objectsMsg); !ok {
		t.Fatal("refresh command should reload objects")
	}
}

func TestUnknownKeyNoop(t *testing.T) {
	m := loadedModel(t)
	next, cmd := update(t, m, runes("z"))
	if cmd != nil || next.cursor != 0 {
		t.Fatal("unknown key should be a no-op")
	}
}

func TestClampCursorOnReload(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m.cursor = 1 // b.fit
	// reload with fewer objects so the cursor is now out of range
	fewer := []backend.Object{{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true}}
	m.cwd = ""
	m2, _ := update(t, m, objectsMsg{fewer})
	if m2.cursor < 0 {
		t.Fatalf("cursor = %d", m2.cursor)
	}
}

func TestCurrentOutOfRange(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	if _, ok := m.current(); ok {
		t.Fatal("current should be false with no entries")
	}
}

func TestVisibleRange(t *testing.T) {
	tests := []struct {
		name               string
		height, cursor, n  int
		wantStart, wantEnd int
	}{
		{"empty", 24, 0, 0, 0, 0},
		{"unknown height shows all", 0, 5, 10, 0, 10},
		{"tall enough shows all", 100, 0, 10, 0, 10},
		{"tiny height one row", 4, 0, 10, 0, 1},
		{"window near top", 10, 2, 10, 0, 4},
		{"window scrolled", 10, 9, 10, 6, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{height: tt.height, cursor: tt.cursor}
			start, end := m.visibleRange(tt.n)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Fatalf("visibleRange = (%d,%d), want (%d,%d)", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestParentDir(t *testing.T) {
	if parentDir("a/b/c") != "a/b" {
		t.Error("nested")
	}
	if parentDir("top") != "" {
		t.Error("top level")
	}
}

func TestViewLoading(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	if !strings.Contains(m.View(), "loading") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewError(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	m, _ = update(t, m, errMsg{errors.New("kaboom")})
	if !strings.Contains(m.View(), "kaboom") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewEmpty(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "Music" // no children
	if !strings.Contains(m.View(), "(empty)") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewEntries(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m = m.toggle(3) // select a.fit
	out := m.View()
	if !strings.Contains(out, "a.fit") || !strings.Contains(out, "b.fit") {
		t.Fatalf("view missing files: %q", out)
	}
	if !strings.Contains(out, "1 selected") {
		t.Fatalf("view missing selection count: %q", out)
	}
	if !strings.Contains(out, "[x]") {
		t.Fatalf("view missing selection marker: %q", out)
	}
}

func TestViewDirSuffixAndScroll(t *testing.T) {
	m := loadedModel(t)
	out := m.View()
	if !strings.Contains(out, "GARMIN/") || !strings.Contains(out, "> ") {
		t.Fatalf("root view = %q", out)
	}
}
