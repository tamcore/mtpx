package backend

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type fakeRunner struct {
	outputs map[string][]byte
	errs    map[string]error
	missing map[string]bool
	calls   [][]string
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		outputs: map[string][]byte{},
		errs:    map[string]error{},
		missing: map[string]bool{},
	}
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, append([]string{name}, args...))
	if err := r.errs[name]; err != nil {
		return nil, err
	}
	return r.outputs[name], nil
}

func (r *fakeRunner) LookPath(name string) (string, error) {
	if r.missing[name] {
		return "", errors.New("not found")
	}
	return "/usr/bin/" + name, nil
}

func assertCall(t *testing.T, fr *fakeRunner, idx int, want []string) {
	t.Helper()
	if idx >= len(fr.calls) {
		t.Fatalf("no call at index %d (calls=%v)", idx, fr.calls)
	}
	if !slices.Equal(fr.calls[idx], want) {
		t.Fatalf("call %d = %v, want %v", idx, fr.calls[idx], want)
	}
}

func TestLibmtpDetect(t *testing.T) {
	fr := newFakeRunner()
	fr.outputs[toolDetect] = readFixture(t, "detect.txt")
	b := &LibmtpBackend{runner: fr}

	devices, err := b.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(devices) != 1 || devices[0].Product != "EPIX 2" {
		t.Fatalf("unexpected devices: %+v", devices)
	}
	assertCall(t, fr, 0, []string{toolDetect})
}

func TestLibmtpDetectError(t *testing.T) {
	fr := newFakeRunner()
	fr.errs[toolDetect] = errors.New("boom")
	b := &LibmtpBackend{runner: fr}
	if _, err := b.Detect(context.Background()); err == nil {
		t.Fatal("want error")
	}
}

func TestLibmtpList(t *testing.T) {
	fr := newFakeRunner()
	fr.outputs[toolFolders] = readFixture(t, "folders.txt")
	fr.outputs[toolFiles] = readFixture(t, "files.txt")
	b := &LibmtpBackend{runner: fr}

	objects, err := b.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var foundActivity, foundFit bool
	for _, o := range objects {
		if o.Path == "GARMIN/Activity" && o.IsDir {
			foundActivity = true
		}
		if o.Path == "GARMIN/Activity/2024-01-15-07-00-00.fit" && !o.IsDir {
			foundFit = true
		}
	}
	if !foundActivity || !foundFit {
		t.Fatalf("List missing objects (activity=%v fit=%v)", foundActivity, foundFit)
	}
}

func TestLibmtpListFolderError(t *testing.T) {
	fr := newFakeRunner()
	fr.errs[toolFolders] = errors.New("no device")
	b := &LibmtpBackend{runner: fr}
	if _, err := b.List(context.Background()); err == nil {
		t.Fatal("want error from folders")
	}
}

func TestLibmtpListFileError(t *testing.T) {
	fr := newFakeRunner()
	fr.outputs[toolFolders] = readFixture(t, "folders.txt")
	fr.errs[toolFiles] = errors.New("no device")
	b := &LibmtpBackend{runner: fr}
	if _, err := b.List(context.Background()); err == nil {
		t.Fatal("want error from files")
	}
}

func TestLibmtpGet(t *testing.T) {
	fr := newFakeRunner()
	b := &LibmtpBackend{runner: fr}
	if err := b.Get(context.Background(), 16777224, "/tmp/out.img"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	assertCall(t, fr, 0, []string{toolGetfile, "16777224", "/tmp/out.img"})
}

func TestLibmtpGetError(t *testing.T) {
	fr := newFakeRunner()
	fr.errs[toolGetfile] = errors.New("fail")
	b := &LibmtpBackend{runner: fr}
	if err := b.Get(context.Background(), 1, "/tmp/x"); err == nil {
		t.Fatal("want error")
	}
}

func TestLibmtpPut(t *testing.T) {
	fr := newFakeRunner()
	b := &LibmtpBackend{runner: fr}
	id, err := b.Put(context.Background(), "/tmp/in.fit", "GARMIN/NewFiles/in.fit")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if id != 0 {
		t.Fatalf("want unreported id 0, got %d", id)
	}
	assertCall(t, fr, 0, []string{toolSendfile, "/tmp/in.fit", "GARMIN/NewFiles/in.fit"})
}

func TestLibmtpPutError(t *testing.T) {
	fr := newFakeRunner()
	fr.errs[toolSendfile] = errors.New("fail")
	b := &LibmtpBackend{runner: fr}
	if _, err := b.Put(context.Background(), "a", "b"); err == nil {
		t.Fatal("want error")
	}
}

func TestLibmtpDelete(t *testing.T) {
	fr := newFakeRunner()
	b := &LibmtpBackend{runner: fr}
	if err := b.Delete(context.Background(), 16778330); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	assertCall(t, fr, 0, []string{toolDelfile, "-n", "16778330"})
}

func TestLibmtpDeleteError(t *testing.T) {
	fr := newFakeRunner()
	fr.errs[toolDelfile] = errors.New("fail")
	b := &LibmtpBackend{runner: fr}
	if err := b.Delete(context.Background(), 1); err == nil {
		t.Fatal("want error")
	}
}

func TestLibmtpEnsureAvailable(t *testing.T) {
	fr := newFakeRunner()
	b := &LibmtpBackend{runner: fr}
	if err := b.EnsureAvailable(); err != nil {
		t.Fatalf("EnsureAvailable: %v", err)
	}
	fr.missing[toolSendfile] = true
	if err := b.EnsureAvailable(); err == nil {
		t.Fatal("want error when a tool is missing")
	}
}

func TestNewLibmtpBackend(t *testing.T) {
	b := NewLibmtpBackend()
	if _, ok := b.runner.(execRunner); !ok {
		t.Fatalf("runner = %T, want execRunner", b.runner)
	}
}
