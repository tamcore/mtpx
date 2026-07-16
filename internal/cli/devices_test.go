package cli

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func oneDevice() *backend.FakeBackend {
	return &backend.FakeBackend{Devices: []backend.Device{
		{Vendor: "Garmin", Product: "EPIX 2", VendorID: 0x091e, ProductID: 0x4f67, Bus: 0, Dev: 5, Serial: "abc123"},
	}}
}

func TestDevicesList(t *testing.T) {
	out, err := runCmd(t, oneDevice(), "devices")
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	for _, want := range []string{"Garmin EPIX 2", "091e:4f67", "bus 0 dev 5", "SN:abc123"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestDevicesAlias(t *testing.T) {
	if _, err := runCmd(t, oneDevice(), "device"); err != nil {
		t.Fatalf("device alias: %v", err)
	}
}

func TestDevicesNone(t *testing.T) {
	out, err := runCmd(t, &backend.FakeBackend{}, "devices")
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	if !strings.Contains(out, "no MTP devices found") {
		t.Errorf("output = %q", out)
	}
}

func TestDevicesJSON(t *testing.T) {
	out, err := runCmd(t, oneDevice(), "devices", "--json")
	if err != nil {
		t.Fatalf("devices --json: %v", err)
	}
	var devs []backend.Device
	if err := json.Unmarshal([]byte(out), &devs); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if len(devs) != 1 || devs[0].Product != "EPIX 2" {
		t.Fatalf("devices = %+v", devs)
	}
}

func TestDevicesError(t *testing.T) {
	bk := &backend.FakeBackend{DetectErr: errors.New("boom")}
	if _, err := runCmd(t, bk, "devices"); err == nil {
		t.Fatal("want detect error")
	}
}

func TestDevicesUnknownNoSerial(t *testing.T) {
	bk := &backend.FakeBackend{Devices: []backend.Device{{VendorID: 0x1234, ProductID: 0x5678}}}
	out, err := runCmd(t, bk, "devices")
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	if !strings.Contains(out, "unknown device") || !strings.Contains(out, "1234:5678") {
		t.Errorf("output = %q", out)
	}
	if strings.Contains(out, "SN:") {
		t.Errorf("should not print serial when empty: %q", out)
	}
}
