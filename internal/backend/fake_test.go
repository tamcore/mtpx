package backend

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var _ Backend = (*FakeBackend)(nil)

func TestFakeDetect(t *testing.T) {
	dev := Device{VendorID: 0x091e, Product: "Garmin", Serial: "ABC"}
	f := &FakeBackend{Devices: []Device{dev}}

	got, err := f.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) != 1 || got[0].Serial != "ABC" {
		t.Fatalf("unexpected devices: %+v", got)
	}
}

func TestFakeDetectError(t *testing.T) {
	wantErr := errors.New("boom")
	f := &FakeBackend{DetectErr: wantErr}

	if _, err := f.Detect(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestFakeList(t *testing.T) {
	f := &FakeBackend{Objects: []Object{
		{ID: 1, Name: "GARMIN", Path: "GARMIN", IsDir: true},
		{ID: 2, Name: "b.fit", Path: "GARMIN/b.fit", Size: 10},
	}}

	got, err := f.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 objects, got %d", len(got))
	}
}

func TestFakeListError(t *testing.T) {
	wantErr := errors.New("nope")
	f := &FakeBackend{ListErr: wantErr}

	if _, err := f.List(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestFakeGet(t *testing.T) {
	f := &FakeBackend{
		Objects:  []Object{{ID: 7, Name: "x.fit", Size: 3}},
		Contents: map[uint32][]byte{7: []byte("abc")},
	}
	dst := filepath.Join(t.TempDir(), "out.fit")

	if err := f.Get(context.Background(), 7, dst); err != nil {
		t.Fatalf("Get: %v", err)
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !bytes.Equal(b, []byte("abc")) {
		t.Fatalf("content = %q, want %q", b, "abc")
	}
}

func TestFakeGetUnknown(t *testing.T) {
	f := &FakeBackend{}
	dst := filepath.Join(t.TempDir(), "out")

	if err := f.Get(context.Background(), 99, dst); err == nil {
		t.Fatal("want error for unknown id")
	}
}

func TestFakeGetError(t *testing.T) {
	wantErr := errors.New("io fail")
	f := &FakeBackend{GetErr: wantErr}

	if err := f.Get(context.Background(), 1, filepath.Join(t.TempDir(), "x")); !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestFakePut(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.fit")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	f := &FakeBackend{}

	id, err := f.Put(context.Background(), src, "GARMIN/NewFiles/in.fit")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if id == 0 {
		t.Fatal("want non-zero id")
	}

	objs, err := f.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var found bool
	for _, o := range objs {
		if o.ID == id && o.Path == "GARMIN/NewFiles/in.fit" && o.Name == "in.fit" && o.Size == 5 {
			found = true
		}
	}
	if !found {
		t.Fatalf("put object not found in %+v", objs)
	}
}

func TestFakePutMissingSource(t *testing.T) {
	f := &FakeBackend{}
	if _, err := f.Put(context.Background(), filepath.Join(t.TempDir(), "nope"), "a/b"); err == nil {
		t.Fatal("want error for missing source file")
	}
}

func TestFakePutError(t *testing.T) {
	wantErr := errors.New("put fail")
	f := &FakeBackend{PutErr: wantErr}
	if _, err := f.Put(context.Background(), "x", "y"); !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestFakeDelete(t *testing.T) {
	f := &FakeBackend{Objects: []Object{{ID: 1}, {ID: 2}}}

	if err := f.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	objs, err := f.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(objs) != 1 || objs[0].ID != 2 {
		t.Fatalf("delete failed, remaining: %+v", objs)
	}
	if len(f.Deleted) != 1 || f.Deleted[0] != 1 {
		t.Fatalf("delete log = %+v, want [1]", f.Deleted)
	}
}

func TestFakeDeleteUnknown(t *testing.T) {
	f := &FakeBackend{}
	if err := f.Delete(context.Background(), 42); err == nil {
		t.Fatal("want error for unknown id")
	}
}

func TestFakeDeleteError(t *testing.T) {
	wantErr := errors.New("del fail")
	f := &FakeBackend{Objects: []Object{{ID: 1}}, DeleteErr: wantErr}
	if err := f.Delete(context.Background(), 1); !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}
