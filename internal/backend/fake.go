package backend

import (
	"context"
	"fmt"
	"os"
	"path"
)

// FakeBackend is an in-memory Backend used by tests and dry runs. It models a
// device's object list and file contents without touching real hardware.
type FakeBackend struct {
	Devices  []Device
	Objects  []Object
	Contents map[uint32][]byte

	DetectErr error
	ListErr   error
	GetErr    error
	PutErr    error
	DeleteErr error

	// Deleted records the ids passed to Delete, in call order.
	Deleted []uint32

	nextID uint32
}

func (f *FakeBackend) Detect(_ context.Context) ([]Device, error) {
	if f.DetectErr != nil {
		return nil, f.DetectErr
	}
	return append([]Device(nil), f.Devices...), nil
}

func (f *FakeBackend) List(_ context.Context) ([]Object, error) {
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	return append([]Object(nil), f.Objects...), nil
}

func (f *FakeBackend) Get(_ context.Context, id uint32, destPath string) error {
	if f.GetErr != nil {
		return f.GetErr
	}
	content, ok := f.Contents[id]
	if !ok {
		return fmt.Errorf("backend: object %d not found", id)
	}
	return os.WriteFile(destPath, content, 0o600)
}

func (f *FakeBackend) Put(_ context.Context, localPath, remotePath string) (uint32, error) {
	if f.PutErr != nil {
		return 0, f.PutErr
	}
	data, err := os.ReadFile(localPath) //nolint:gosec // test double reads caller-provided path
	if err != nil {
		return 0, err
	}
	f.nextID++
	id := 1000 + f.nextID
	f.Objects = append(f.Objects, Object{
		ID:   id,
		Name: path.Base(remotePath),
		Path: remotePath,
		Size: int64(len(data)),
	})
	if f.Contents == nil {
		f.Contents = make(map[uint32][]byte)
	}
	f.Contents[id] = data
	return id, nil
}

func (f *FakeBackend) Delete(_ context.Context, id uint32) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	idx := -1
	for i, o := range f.Objects {
		if o.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("backend: object %d not found", id)
	}
	f.Objects = append(f.Objects[:idx:idx], f.Objects[idx+1:]...)
	f.Deleted = append(f.Deleted, id)
	return nil
}
