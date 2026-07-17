package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

// loadedAt returns a model backed by bk, loaded, and focused at the given dirs.
func loadedAt(t *testing.T, bk *backend.FakeBackend, dirs ...string) Model {
	t.Helper()
	next, _ := NewModel(context.Background(), bk, t.TempDir()).Update(objectsMsg{bk.Objects})
	m := next.(Model)
	cols := []column{{dir: ""}}
	for _, d := range dirs {
		cols = append(cols, column{dir: d})
	}
	m.cols = cols
	return m
}

// drainDelete presses "y" to start deleting, then runs the streamed commands to
// completion, returning the final model.
func drainDelete(t *testing.T, m Model) Model {
	t.Helper()
	m, cmd := update(t, m, runes("y"))
	for cmd != nil {
		m, cmd = update(t, m, cmd())
	}
	return m
}

func TestTargetsSelected(t *testing.T) {
	m := loadedModel(t)
	m.selected = map[uint32]bool{3: true, 4: true}
	if len(m.targets()) != 2 {
		t.Fatalf("targets = %d, want 2", len(m.targets()))
	}
}

func TestTargetsFocusedFile(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	tg := m.targets()
	if len(tg) != 1 || tg[0].Name != "a.fit" {
		t.Fatalf("targets = %+v", tg)
	}
}

func TestTargetsNone(t *testing.T) {
	if tg := loadedModel(t).targets(); tg != nil {
		t.Fatalf("targets = %+v, want nil", tg)
	}
}

func TestDeletePromptsConfirm(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	if m.mode != modeConfirm || len(m.pending) != 1 {
		t.Fatalf("mode=%v pending=%d", m.mode, len(m.pending))
	}
}

func TestDeleteNoTargets(t *testing.T) {
	m, _ := update(t, loadedModel(t), runes("d"))
	if m.mode != modeBrowse || !strings.Contains(m.message, "no files") {
		t.Fatalf("mode=%v msg=%q", m.mode, m.message)
	}
}

func TestConfirmYesStartsDeleting(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	m, cmd := update(t, m, runes("y"))
	if m.mode != modeDeleting || cmd == nil {
		t.Fatalf("mode=%v cmd=%v", m.mode, cmd)
	}
	if _, ok := cmd().(deletedOneMsg); !ok {
		t.Fatalf("first command = %T, want deletedOneMsg", cmd())
	}
}

func TestDeleteStreamsToCompletion(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects()}
	m := loadedAt(t, bk, "GARMIN", "GARMIN/Activity")
	m = m.toggle(3)
	m = m.toggle(4)
	m, _ = update(t, m, runes("d"))
	m = drainDelete(t, m)

	if m.mode != modeBrowse || !strings.Contains(m.message, "deleted 2 file(s)") {
		t.Fatalf("mode=%v msg=%q", m.mode, m.message)
	}
	if len(bk.Deleted) != 2 {
		t.Fatalf("backend deleted %v", bk.Deleted)
	}
	if len(m.log) != 2 {
		t.Fatalf("log = %v", m.log)
	}
	for _, o := range m.objects {
		if o.ID == 3 || o.ID == 4 {
			t.Fatalf("deleted object still cached: %+v", o)
		}
	}
	if len(m.selected) != 0 {
		t.Fatal("selection should clear")
	}
}

func TestDeleteStreamsFailure(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects(), DeleteErr: errors.New("busy")}
	m := loadedAt(t, bk, "GARMIN", "GARMIN/Activity")
	m = m.toggle(3)
	m, _ = update(t, m, runes("d"))
	m = drainDelete(t, m)

	if !strings.Contains(m.message, "1 failed") {
		t.Fatalf("msg = %q", m.message)
	}
	if len(m.log) != 1 || !strings.Contains(m.log[0], "failed") {
		t.Fatalf("log = %v", m.log)
	}
}

func TestConfirmCancel(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	m, _ = update(t, m, runes("n"))
	if m.mode != modeBrowse || m.message != "cancelled" || len(m.pending) != 0 {
		t.Fatalf("mode=%v msg=%q pending=%d", m.mode, m.message, len(m.pending))
	}
}

func TestConfirmEsc(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	if m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEsc}); m.mode != modeBrowse {
		t.Fatal("esc should cancel")
	}
}

func TestConfirmCtrlCQuits(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	_, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected QuitMsg")
	}
}

func TestConfirmOtherKeyStays(t *testing.T) {
	m := atPath(t, "GARMIN", "GARMIN/Activity")
	m, _ = update(t, m, runes("d"))
	m, cmd := update(t, m, runes("z"))
	if m.mode != modeConfirm || cmd != nil {
		t.Fatalf("unknown key should keep confirm: mode=%v", m.mode)
	}
}

func TestDeletingCtrlCQuits(t *testing.T) {
	m := loadedModel(t)
	m.mode = modeDeleting
	m.queue = []backend.Object{{ID: 1}}
	_, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected QuitMsg")
	}
}

func TestDeletingIgnoresOtherKeys(t *testing.T) {
	m := loadedModel(t)
	m.mode = modeDeleting
	_, cmd := update(t, m, runes("x"))
	if cmd != nil {
		t.Fatal("keys other than ctrl+c should be ignored while deleting")
	}
}

func TestPullCurrent(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects(), Contents: map[uint32][]byte{3: []byte("abc")}}
	m := loadedAt(t, bk, "GARMIN", "GARMIN/Activity")
	m2, cmd := update(t, m, runes("c"))
	if cmd == nil || m2.message != "pulling…" {
		t.Fatalf("msg=%q cmd=%v", m2.message, cmd)
	}
	if pm, ok := cmd().(pulledMsg); !ok || pm.count != 1 {
		t.Fatalf("pulledMsg = %+v ok=%v", pm, ok)
	}
	b, err := os.ReadFile(filepath.Join(m.destDir, "GARMIN", "Activity", "a.fit"))
	if err != nil || string(b) != "abc" {
		t.Fatalf("pulled content = %q err=%v", b, err)
	}
}

func TestPullNoTargets(t *testing.T) {
	m, cmd := update(t, loadedModel(t), runes("c"))
	if cmd != nil || !strings.Contains(m.message, "no files") {
		t.Fatalf("msg=%q cmd=%v", m.message, cmd)
	}
}

func TestPulledMsgSuccess(t *testing.T) {
	m := loadedModel(t)
	m.destDir = "/tmp/x"
	m, _ = update(t, m, pulledMsg{count: 3, failed: 0})
	if !strings.Contains(m.message, "pulled 3") {
		t.Fatalf("msg=%q", m.message)
	}
}

func TestPulledMsgFailures(t *testing.T) {
	m, _ := update(t, loadedModel(t), pulledMsg{count: 1, failed: 2})
	if !strings.Contains(m.message, "2 failed") {
		t.Fatalf("msg=%q", m.message)
	}
}

func TestPullCmdGetError(t *testing.T) {
	bk := &backend.FakeBackend{GetErr: errors.New("io")}
	m := NewModel(context.Background(), bk, t.TempDir())
	pm := m.pullCmd([]backend.Object{{ID: 3, Path: "GARMIN/Activity/a.fit"}})().(pulledMsg)
	if pm.failed != 1 {
		t.Fatalf("expected failure, got %+v", pm)
	}
}

func TestPullCmdMkdirError(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	bk := &backend.FakeBackend{Contents: map[uint32][]byte{3: []byte("abc")}}
	m := NewModel(context.Background(), bk, blocker)
	pm := m.pullCmd([]backend.Object{{ID: 3, Path: "GARMIN/Activity/a.fit"}})().(pulledMsg)
	if pm.failed != 1 {
		t.Fatalf("expected mkdir failure, got %+v", pm)
	}
}
