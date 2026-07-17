package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

func TestFilterEntries(t *testing.T) {
	entries := []backend.Object{{Name: "apple"}, {Name: "Banana"}, {Name: "cherry"}}
	if got := filterEntries(entries, ""); len(got) != 3 {
		t.Errorf("empty filter should keep all, got %d", len(got))
	}
	if got := filterEntries(entries, "AN"); len(got) != 1 || got[0].Name != "Banana" {
		t.Errorf("case-insensitive filter = %+v", got)
	}
	if got := filterEntries(entries, "z"); len(got) != 0 {
		t.Errorf("no match should be empty, got %+v", got)
	}
}

func TestSearchKeyEntersMode(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("s"))
	if m.mode != modeSearch || m.filter != "" {
		t.Fatalf("mode=%v filter=%q", m.mode, m.filter)
	}
}

func TestSearchFilters(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity") // a.fit, b.fit
	m, _ = update(t, m, runes("s"))
	m, _ = update(t, m, runes("a"))
	if m.filter != "a" {
		t.Fatalf("filter = %q", m.filter)
	}
	fe := m.focusedEntries()
	if len(fe) != 1 || fe[0].Name != "a.fit" {
		t.Fatalf("focused entries = %+v", fe)
	}
}

func TestSearchBackspace(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.mode = modeSearch
	m.filter = "ab"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.filter != "a" {
		t.Fatalf("filter = %q", m.filter)
	}
	m.filter = ""
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyBackspace}) // empty, no change
	if m.filter != "" {
		t.Fatalf("backspace on empty filter = %q", m.filter)
	}
}

func TestSearchEnterKeepsFilter(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.mode = modeSearch
	m.filter = "a"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeBrowse || m.filter != "a" {
		t.Fatalf("mode=%v filter=%q", m.mode, m.filter)
	}
}

func TestSearchEscClears(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.mode = modeSearch
	m.filter = "a"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeBrowse || m.filter != "" {
		t.Fatalf("mode=%v filter=%q", m.mode, m.filter)
	}
}

func TestSearchUpDown(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("s")) // filter "", cursor 0, entries a.fit, b.fit
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.focused().cursor != 1 {
		t.Fatalf("down cursor = %d", m.focused().cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyDown}) // clamp
	if m.focused().cursor != 1 {
		t.Fatalf("down should clamp, got %d", m.focused().cursor)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyUp}) // clamp
	if m.focused().cursor != 0 {
		t.Fatalf("up should clamp, got %d", m.focused().cursor)
	}
}

func TestSearchCtrlCQuits(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.mode = modeSearch
	_, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected QuitMsg")
	}
}

func TestSearchIgnoresOtherKeys(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.mode = modeSearch
	m.filter = "x"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyHome})
	if m.filter != "x" || m.mode != modeSearch {
		t.Fatalf("non-input key changed state: filter=%q mode=%v", m.filter, m.mode)
	}
}

func TestFilterClearedOnDescend(t *testing.T) {
	m := atPath(t, "GARMIN") // focused GARMIN, entry Activity
	m.filter = "act"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if m.filter != "" || m.focused().dir != "GARMIN/Activity" {
		t.Fatalf("descend should clear filter: filter=%q dir=%q", m.filter, m.focused().dir)
	}
}

func TestFilterClearedOnBack(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.filter = "x"
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.filter != "" {
		t.Fatalf("back should clear filter, got %q", m.filter)
	}
}

func TestFocusedEntryFilteredNoMatch(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m.filter = "zzz"
	if _, ok := m.focusedEntry(); ok {
		t.Fatal("no match should give no focused entry")
	}
}

func TestViewSearchHeader(t *testing.T) {
	m := loadedModel(t)
	m.mode = modeSearch
	m.filter = "act"
	if !strings.Contains(m.View(), "[search: act]") {
		t.Fatalf("view = %q", m.View())
	}
}
