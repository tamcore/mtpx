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

func TestTargetsSelected(t *testing.T) {
	m := loadedModel(t)
	m.selected = map[uint32]bool{3: true, 4: true}
	if len(m.targets()) != 2 {
		t.Fatalf("targets = %d, want 2", len(m.targets()))
	}
}

func TestTargetsCurrentFile(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	tg := m.targets()
	if len(tg) != 1 || tg[0].Name != "a.fit" {
		t.Fatalf("targets = %+v", tg)
	}
}

func TestTargetsNone(t *testing.T) {
	m := loadedModel(t) // cursor on a directory, nothing selected
	if tg := m.targets(); tg != nil {
		t.Fatalf("targets = %+v, want nil", tg)
	}
}

func TestDeletePromptsConfirm(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	if m.mode != modeConfirm || len(m.pending) != 1 {
		t.Fatalf("mode=%v pending=%d", m.mode, len(m.pending))
	}
}

func TestDeleteNoTargets(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, runes("d"))
	if m.mode != modeBrowse || !strings.Contains(m.message, "no files") {
		t.Fatalf("mode=%v msg=%q", m.mode, m.message)
	}
}

func TestConfirmYesDeletes(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects()}
	next, _ := NewModel(context.Background(), bk, "").Update(objectsMsg{sampleObjects()})
	m := next.(Model)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	m, cmd := update(t, m, runes("y"))
	if m.mode != modeBrowse || !m.loading || cmd == nil {
		t.Fatalf("after y: mode=%v loading=%v cmd=%v", m.mode, m.loading, cmd)
	}
	dm, ok := cmd().(deletedMsg)
	if !ok || dm.count != 1 || dm.failed != 0 {
		t.Fatalf("deletedMsg = %+v ok=%v", dm, ok)
	}
	if len(bk.Deleted) != 1 {
		t.Fatalf("backend deleted %v", bk.Deleted)
	}
}

func TestConfirmCancel(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	m, _ = update(t, m, runes("n"))
	if m.mode != modeBrowse || m.message != "cancelled" || len(m.pending) != 0 {
		t.Fatalf("mode=%v msg=%q pending=%d", m.mode, m.message, len(m.pending))
	}
}

func TestConfirmEsc(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	if m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyEsc}); m.mode != modeBrowse {
		t.Fatal("esc should cancel confirm")
	}
}

func TestConfirmCtrlCQuits(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
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
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	m, cmd := update(t, m, runes("z"))
	if m.mode != modeConfirm || cmd != nil {
		t.Fatalf("unknown key should keep confirm: mode=%v", m.mode)
	}
}

func TestDeletedMsgFailures(t *testing.T) {
	m := loadedModel(t)
	m.selected = map[uint32]bool{3: true}
	m, _ = update(t, m, deletedMsg{count: 2, failed: 1})
	if !strings.Contains(m.message, "1 failed") || !m.loading {
		t.Fatalf("msg=%q loading=%v", m.message, m.loading)
	}
	if len(m.selected) != 0 {
		t.Fatal("selection should clear after delete")
	}
}

func TestDeletedMsgSuccess(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, deletedMsg{count: 2, failed: 0})
	if !strings.Contains(m.message, "deleted 2 file(s)") || !m.loading {
		t.Fatalf("msg=%q loading=%v", m.message, m.loading)
	}
}

func TestPullCurrent(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects(), Contents: map[uint32][]byte{3: []byte("abc")}}
	next, _ := NewModel(context.Background(), bk, t.TempDir()).Update(objectsMsg{sampleObjects()})
	m := next.(Model)
	m.cwd = "GARMIN/Activity"
	m2, cmd := update(t, m, runes("c"))
	if cmd == nil || m2.message != "pulling…" {
		t.Fatalf("c: msg=%q cmd=%v", m2.message, cmd)
	}
	pm, ok := cmd().(pulledMsg)
	if !ok || pm.count != 1 || pm.failed != 0 {
		t.Fatalf("pulledMsg = %+v ok=%v", pm, ok)
	}
	b, err := os.ReadFile(filepath.Join(m.destDir, "GARMIN", "Activity", "a.fit"))
	if err != nil || string(b) != "abc" {
		t.Fatalf("pulled content = %q err=%v", b, err)
	}
}

func TestPullNoTargets(t *testing.T) {
	m := loadedModel(t)
	m2, cmd := update(t, m, runes("c"))
	if cmd != nil || !strings.Contains(m2.message, "no files") {
		t.Fatalf("msg=%q cmd=%v", m2.message, cmd)
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
	m := loadedModel(t)
	m, _ = update(t, m, pulledMsg{count: 1, failed: 2})
	if !strings.Contains(m.message, "2 failed") {
		t.Fatalf("msg=%q", m.message)
	}
}

func TestPullCmdGetError(t *testing.T) {
	bk := &backend.FakeBackend{GetErr: errors.New("io")}
	m := NewModel(context.Background(), bk, t.TempDir())
	pm := m.pullCmd([]backend.Object{{ID: 3, Path: "GARMIN/Activity/a.fit"}})().(pulledMsg)
	if pm.failed != 1 {
		t.Fatalf("expected 1 failure, got %+v", pm)
	}
}

func TestPullCmdMkdirError(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	bk := &backend.FakeBackend{Contents: map[uint32][]byte{3: []byte("abc")}}
	m := NewModel(context.Background(), bk, blocker) // destDir is a file
	pm := m.pullCmd([]backend.Object{{ID: 3, Path: "GARMIN/Activity/a.fit"}})().(pulledMsg)
	if pm.failed != 1 {
		t.Fatalf("expected mkdir failure, got %+v", pm)
	}
}

func TestDeleteCmdFailure(t *testing.T) {
	bk := &backend.FakeBackend{DeleteErr: errors.New("busy"), Objects: []backend.Object{{ID: 3}}}
	m := NewModel(context.Background(), bk, "")
	dm := m.deleteCmd([]backend.Object{{ID: 3}})().(deletedMsg)
	if dm.failed != 1 {
		t.Fatalf("expected failure, got %+v", dm)
	}
}

func TestViewConfirmMode(t *testing.T) {
	m := loadedModel(t)
	m.cwd = "GARMIN/Activity"
	m, _ = update(t, m, runes("d"))
	if !strings.Contains(m.View(), "Delete 1 file(s)? (y/n)") {
		t.Fatalf("view = %q", m.View())
	}
}

func TestViewMessage(t *testing.T) {
	m := loadedModel(t)
	m.message = "hello there"
	if !strings.Contains(m.View(), "hello there") {
		t.Fatalf("view = %q", m.View())
	}
}
