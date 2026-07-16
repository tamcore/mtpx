package cli

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tamcore/mtpx/internal/backend"
)

// testDev is a single fake device used to satisfy the device gate in tests.
func testDev() []backend.Device {
	return []backend.Device{{Vendor: "Test", Product: "Device"}}
}

// appearsAfter reports a device only from the nth Detect call onward.
type appearsAfter struct {
	*backend.FakeBackend
	n     int
	calls int
}

func (a *appearsAfter) Detect(context.Context) ([]backend.Device, error) {
	a.calls++
	if a.calls >= a.n {
		return testDev(), nil
	}
	return nil, nil
}

func TestEnsureDevicePresent(t *testing.T) {
	bk := &backend.FakeBackend{Devices: testDev()}
	if err := ensureDevice(context.Background(), bk, 0, time.Millisecond); err != nil {
		t.Fatalf("present: %v", err)
	}
}

func TestEnsureDeviceNoneImmediate(t *testing.T) {
	err := ensureDevice(context.Background(), &backend.FakeBackend{}, 0, time.Millisecond)
	if !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
}

func TestEnsureDeviceDetectError(t *testing.T) {
	bk := &backend.FakeBackend{DetectErr: errors.New("usb")}
	if err := ensureDevice(context.Background(), bk, 0, time.Millisecond); err == nil {
		t.Fatal("want detect error")
	}
}

func TestEnsureDeviceAppearsAfterPolling(t *testing.T) {
	bk := &appearsAfter{FakeBackend: &backend.FakeBackend{}, n: 3}
	if err := ensureDevice(context.Background(), bk, time.Second, time.Millisecond); err != nil {
		t.Fatalf("should find device after polling: %v", err)
	}
	if bk.calls < 3 {
		t.Fatalf("polled %d times, want >= 3", bk.calls)
	}
}

func TestEnsureDeviceTimeout(t *testing.T) {
	err := ensureDevice(context.Background(), &backend.FakeBackend{}, 5*time.Millisecond, time.Millisecond)
	if !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice on timeout, got %v", err)
	}
}

func TestEnsureDeviceContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ensureDevice(ctx, &backend.FakeBackend{}, time.Second, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestListRequiresDevice(t *testing.T) {
	if _, err := runCmd(t, &backend.FakeBackend{}, "list"); !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
}

func TestPullRequiresDevice(t *testing.T) {
	if _, err := runCmd(t, &backend.FakeBackend{}, "pull", "a", t.TempDir()); !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
}

func TestDeleteRequiresDevice(t *testing.T) {
	if _, err := runCmd(t, &backend.FakeBackend{}, "delete", "x"); !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
}

func TestPurgeRequiresDevice(t *testing.T) {
	if _, err := runCmd(t, &backend.FakeBackend{}, "purge"); !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
}

func TestTUIRequiresDevice(t *testing.T) {
	var launched bool
	launch := func(context.Context, backend.Backend, string) error {
		launched = true
		return nil
	}
	root := NewRootCmd("v", "c", &backend.FakeBackend{}, launch)
	root.SetArgs([]string{})
	if err := root.Execute(); !errors.Is(err, errNoDevice) {
		t.Fatalf("want errNoDevice, got %v", err)
	}
	if launched {
		t.Fatal("TUI should not launch without a device")
	}
}

func TestWaitFlagAllowsDevicePresent(t *testing.T) {
	// Exercises requireDevice reading a non-zero --wait when a device is already
	// present, so it returns without polling.
	if _, err := runCmd(t, &backend.FakeBackend{Devices: testDev()}, "list", "--wait", "5s"); err != nil {
		t.Fatalf("list --wait with device present: %v", err)
	}
}

func TestDevicesWaitFlag(t *testing.T) {
	out, err := runCmd(t, &backend.FakeBackend{Devices: testDev()}, "devices", "--wait", "5s")
	if err != nil {
		t.Fatalf("devices --wait: %v", err)
	}
	if out == "" {
		t.Fatal("expected device output")
	}
}
