package ui

import (
	"context"
	"errors"
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
	next, _ := NewModel(context.Background(), &backend.FakeBackend{Objects: sampleObjects()}, "").
		Update(objectsMsg{sampleObjects()})
	return next.(Model)
}

// atPath returns a loaded model whose open columns end at the given dirs.
func atPath(t *testing.T, dirs ...string) Model {
	t.Helper()
	m := loadedModel(t)
	cols := []column{{dir: ""}}
	for _, d := range dirs {
		cols = append(cols, column{dir: d})
	}
	m.cols = cols
	return m
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
	if _, ok := m.Init()().(objectsMsg); !ok {
		t.Fatal("Init should load objects")
	}
}

func TestInitError(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{ListErr: errors.New("x")}, "")
	if _, ok := m.Init()().(errMsg); !ok {
		t.Fatal("Init should surface load error")
	}
}

func TestUpdateObjects(t *testing.T) {
	m := loadedModel(t)
	if m.loading {
		t.Error("loading should be false")
	}
	if len(m.entriesOf("")) != 2 {
		t.Errorf("root entries = %d, want 2", len(m.entriesOf("")))
	}
}

func TestUpdateError(t *testing.T) {
	m, _ := update(t, loadedModel(t), errMsg{errors.New("boom")})
	if m.err == nil || m.loading {
		t.Fatalf("err=%v loading=%v", m.err, m.loading)
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m, _ := update(t, loadedModel(t), tea.WindowSizeMsg{Width: 100, Height: 40})
	if m.width != 100 || m.height != 40 {
		t.Fatalf("size = %dx%d", m.width, m.height)
	}
}

func TestUpdateUnknownMsg(t *testing.T) {
	m, cmd := update(t, loadedModel(t), struct{}{})
	if cmd != nil || len(m.cols) != 1 {
		t.Fatal("unknown msg should be a no-op")
	}
}

func TestNavigateDownUp(t *testing.T) {
	m := loadedModel(t) // root: GARMIN, Music
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.focused().cursor != 1 {
		t.Fatalf("cursor after down = %d", m.focused().cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown}) // clamp at last
	if m.focused().cursor != 1 {
		t.Fatalf("cursor should clamp, got %d", m.focused().cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp}) // clamp at 0
	if m.focused().cursor != 0 {
		t.Fatalf("cursor should clamp at 0, got %d", m.focused().cursor)
	}
}

func TestDescendOpensColumn(t *testing.T) {
	m := loadedModel(t) // cursor on GARMIN
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.cols) != 2 || m.focused().dir != "GARMIN" {
		t.Fatalf("cols=%d focused=%q", len(m.cols), m.focused().dir)
	}
}

func TestBackClosesColumn(t *testing.T) {
	m := atPath(t, "GARMIN")
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyLeft})
	if len(m.cols) != 1 {
		t.Fatalf("cols after back = %d, want 1", len(m.cols))
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyLeft}) // at root, no-op
	if len(m.cols) != 1 {
		t.Fatalf("back at root should stay, got %d", len(m.cols))
	}
}

func TestDescendOnFileNoop(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity") // focused a.fit
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.cols) != 3 {
		t.Fatalf("entering a file opened a column: cols=%d", len(m.cols))
	}
}

func TestSelectToggle(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity") // focused a.fit (id 3)
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.selected[3] {
		t.Fatal("a.fit should be selected")
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace})
	if m.selected[3] {
		t.Fatal("a.fit should be deselected")
	}
}

func TestSelectDirNoop(t *testing.T) {
	m := loadedModel(t) // focused GARMIN (dir)
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeySpace})
	if len(m.selected) != 0 {
		t.Fatal("directories must not be selectable")
	}
}

func TestQuit(t *testing.T) {
	for _, k := range []tea.KeyMsg{{Type: tea.KeyCtrlC}, runes("q")} {
		_, cmd := update(t, loadedModel(t), k)
		if cmd == nil {
			t.Fatalf("%v should quit", k)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%v not QuitMsg", k)
		}
	}
}

func TestRefresh(t *testing.T) {
	m, cmd := update(t, loadedModel(t), runes("r"))
	if !m.loading || cmd == nil {
		t.Fatal("refresh should load")
	}
	if _, ok := cmd().(objectsMsg); !ok {
		t.Fatal("refresh should reload objects")
	}
}

func TestUnknownKeyNoop(t *testing.T) {
	m, cmd := update(t, loadedModel(t), runes("z"))
	if cmd != nil || m.focused().cursor != 0 {
		t.Fatal("unknown key should be a no-op")
	}
}

func TestReconcilePrunesMissingDir(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	objs := []backend.Object{
		{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
		{ID: 5, Path: "Music", Name: "Music", IsDir: true},
	}
	m, _ = update(t, m, objectsMsg{objs})
	if len(m.cols) != 2 || m.cols[1].dir != "GARMIN" {
		t.Fatalf("cols = %v", m.cols)
	}
}

func TestReconcileClampsCursor(t *testing.T) {
	m := loadedModel(t)
	m.cols = []column{{dir: "", cursor: 5}}
	m, _ = update(t, m, objectsMsg{sampleObjects()})
	if m.cols[0].cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cols[0].cursor)
	}
}

func TestFocusedEntryOutOfRange(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	if _, ok := m.focusedEntry(); ok {
		t.Fatal("no entries should give no focused entry")
	}
}

func TestWindowAround(t *testing.T) {
	tests := []struct {
		name               string
		cursor, n, rows    int
		wantStart, wantEnd int
	}{
		{"empty", 0, 0, 5, 0, 0},
		{"rows zero shows all", -1, 10, 0, 0, 10},
		{"rows exceed n", 0, 5, 10, 0, 5},
		{"near top", 2, 10, 4, 0, 4},
		{"scrolled", 8, 10, 4, 5, 9},
		{"preview no cursor", -1, 10, 4, 0, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, e := windowAround(tt.cursor, tt.n, tt.rows)
			if s != tt.wantStart || e != tt.wantEnd {
				t.Fatalf("windowAround(%d,%d,%d) = (%d,%d), want (%d,%d)",
					tt.cursor, tt.n, tt.rows, s, e, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestIndexOfPath(t *testing.T) {
	entries := []backend.Object{{Path: "a"}, {Path: "b"}}
	if indexOfPath(entries, "b") != 1 {
		t.Error("found")
	}
	if indexOfPath(entries, "z") != -1 {
		t.Error("missing")
	}
}
