package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func TestViewLoading(t *testing.T) {
	m := NewModel(context.Background(), &backend.FakeBackend{}, "")
	if !strings.Contains(m.View(), "loading") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewError(t *testing.T) {
	m, _ := update(t, loadedModel(t), errMsg{errors.New("kaboom")})
	if !strings.Contains(m.View(), "kaboom") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewConfirm(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d")) // focused a.fit -> confirm
	if !strings.Contains(m.View(), "Delete 1 file(s)? (y/n)") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewMessage(t *testing.T) {
	m := loadedModel(t)
	m.message = "hi there"
	if !strings.Contains(m.View(), "hi there") {
		t.Fatalf("view = %q", m.View())
	}
}

// TestViewColumnsTrailActivePreview exercises a trail (non-active) column, the
// active column cursor, and the preview pane in one render.
func TestViewColumnsTrailActivePreview(t *testing.T) {
	m := atPath(t, "GARMIN") // root (trail) + GARMIN (active); focus Activity -> preview
	out := m.View()
	for _, want := range []string{"·", ">", "GARMIN/", "Activity/", "a.fit", "selected"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q:\n%s", want, out)
		}
	}
}

func TestRows(t *testing.T) {
	cases := map[int]int{0: defaultRows, 40: 40 - reservedRows, 4: 1}
	for height, want := range cases {
		if got := (Model{height: height}).rows(); got != want {
			t.Errorf("rows(height=%d) = %d, want %d", height, got, want)
		}
	}
}

func TestVisibleColumns(t *testing.T) {
	cols := make([]renderCol, 4)
	if got := visibleColumns(cols, 0); len(got) != 4 {
		t.Errorf("width 0 should show all, got %d", len(got))
	}
	if got := visibleColumns(cols[:3], 100); len(got) != 3 {
		t.Errorf("wide enough should show all, got %d", len(got))
	}
	if got := visibleColumns(cols, 50); len(got) != 2 { // 50/22 = 2
		t.Errorf("width 50 should show 2, got %d", len(got))
	}
	if got := visibleColumns(cols, 10); len(got) != 1 { // 10/22 = 0 -> 1
		t.Errorf("narrow should show 1, got %d", len(got))
	}
}

func TestColumnLinesActiveAndSelection(t *testing.T) {
	m := loadedModel(t)
	m.selected = map[uint32]bool{3: true}
	rc := renderCol{
		entries: []backend.Object{
			{ID: 3, Name: "a.fit"},
			{ID: 2, Name: "Activity", IsDir: true},
		},
		cursor: 0,
		active: true,
	}
	lines := m.columnLines(rc, 4)
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], ">x a.fit") {
		t.Errorf("line0 = %q", lines[0])
	}
	if !strings.Contains(lines[1], "Activity/") {
		t.Errorf("line1 = %q", lines[1])
	}
}

func TestColumnLinesTrailMarker(t *testing.T) {
	rc := renderCol{entries: []backend.Object{{ID: 1, Name: "x"}}, cursor: 0, active: false}
	lines := (Model{}).columnLines(rc, 2)
	if !strings.HasPrefix(lines[0], "· ") {
		t.Errorf("trail marker line0 = %q", lines[0])
	}
}

func TestColumnLinesPreviewNoCursor(t *testing.T) {
	rc := renderCol{entries: []backend.Object{{ID: 1, Name: "x"}}, cursor: -1}
	lines := (Model{}).columnLines(rc, 2)
	if strings.ContainsAny(lines[0], ">·") {
		t.Errorf("preview line should have no cursor marker: %q", lines[0])
	}
}

func TestColumnLinesEmpty(t *testing.T) {
	lines := (Model{}).columnLines(renderCol{}, 3)
	if !strings.Contains(lines[0], "(empty)") {
		t.Errorf("empty column = %q", lines[0])
	}
}

func TestColumnLinesWindows(t *testing.T) {
	var entries []backend.Object
	for i := 0; i < 10; i++ {
		entries = append(entries, backend.Object{ID: uint32(i), Name: string(rune('a' + i))})
	}
	rc := renderCol{entries: entries, cursor: 8, active: true}
	lines := (Model{}).columnLines(rc, 3)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "i") { // entry index 8 = 'i', the cursor
		t.Errorf("cursor entry should be visible: %q", joined)
	}
	if strings.Contains(joined, "a") { // index 0 scrolled off
		t.Errorf("top entry should be scrolled off: %q", joined)
	}
}

func TestPad(t *testing.T) {
	if got := pad("hi", 5); got != "hi   " {
		t.Errorf("pad short = %q", got)
	}
	if got := pad("toolongname", 5); len([]rune(got)) != 5 {
		t.Errorf("pad long len = %d (%q)", len([]rune(got)), got)
	}
}
