package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestParseDetect(t *testing.T) {
	devices := parseDetect(readFixture(t, "detect.txt"))
	if len(devices) != 1 {
		t.Fatalf("want 1 device, got %d: %+v", len(devices), devices)
	}
	d := devices[0]
	if d.Vendor != "Garmin" || d.Product != "EPIX 2" {
		t.Errorf("vendor/product = %q/%q", d.Vendor, d.Product)
	}
	if d.VendorID != 0x091e || d.ProductID != 0x4f67 {
		t.Errorf("ids = %04x:%04x", d.VendorID, d.ProductID)
	}
	if d.Bus != 0 || d.Dev != 5 {
		t.Errorf("bus/dev = %d/%d", d.Bus, d.Dev)
	}
	if d.Serial != "0000000000000000" {
		t.Errorf("serial = %q", d.Serial)
	}
}

func TestParseDetectFallbackAndEnrichment(t *testing.T) {
	in := []byte("   091e:4f67 @ bus 1, dev 9\n" +
		"   Manufacturer: Garmin\n" +
		"   Model: EPIX\n" +
		"   Serial number: ABC123\n")
	devices := parseDetect(in)
	if len(devices) != 1 {
		t.Fatalf("want 1 device, got %d", len(devices))
	}
	d := devices[0]
	if d.VendorID != 0x091e || d.ProductID != 0x4f67 || d.Bus != 1 || d.Dev != 9 {
		t.Errorf("unexpected raw fields: %+v", d)
	}
	if d.Vendor != "Garmin" || d.Product != "EPIX" || d.Serial != "ABC123" {
		t.Errorf("enrichment failed: %+v", d)
	}
}

func TestParseDetectEmpty(t *testing.T) {
	if devices := parseDetect([]byte("nothing here\n")); len(devices) != 0 {
		t.Fatalf("want no devices, got %+v", devices)
	}
}

func TestParseFolders(t *testing.T) {
	entries := parseFolders(readFixture(t, "folders.txt"))
	want := map[uint32]folderEntry{
		16777216: {16777216, 0, "GARMIN"},
		16777252: {16777252, 1, "Activity"},
		16777253: {16777253, 1, "Workouts"},
		16777770: {16777770, 2, "Schedule"},
		16777270: {16777270, 1, "PaceBands"},
		16777218: {16777218, 0, "Music"},
	}
	got := make(map[uint32]folderEntry, len(entries))
	for _, e := range entries {
		got[e.id] = e
	}
	if len(entries) != 8 {
		t.Fatalf("want 8 folders, got %d", len(entries))
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("folder %d = %+v, want %+v", id, got[id], w)
		}
	}
}

func TestParseFiles(t *testing.T) {
	files := parseFiles(readFixture(t, "files.txt"))
	if len(files) != 6 {
		t.Fatalf("want 6 files, got %d: %+v", len(files), files)
	}
	byID := make(map[uint32]fileEntry, len(files))
	for _, f := range files {
		byID[f.id] = f
	}
	if f := byID[16777224]; f.name != "gmaptz.img" || f.size != 572416 || f.parentID != 16777216 || f.storageID != 0x00020001 {
		t.Errorf("gmaptz.img entry wrong: %+v", f)
	}
	if f := byID[16778330]; f.name != "abstractfile" || f.size != -1 {
		t.Errorf("abstract file entry wrong: %+v", f)
	}
}

func TestBuildTree(t *testing.T) {
	folders := parseFolders(readFixture(t, "folders.txt"))
	files := parseFiles(readFixture(t, "files.txt"))
	objects := buildTree(folders, files)

	byID := make(map[uint32]Object, len(objects))
	for _, o := range objects {
		byID[o.ID] = o
	}

	tests := []struct {
		id     uint32
		path   string
		isDir  bool
		parent uint32
	}{
		{16777216, "GARMIN", true, 0},
		{16777252, "GARMIN/Activity", true, 16777216},
		{16777770, "GARMIN/Workouts/Schedule", true, 16777253},
		{16777218, "Music", true, 0},
		{16777224, "GARMIN/gmaptz.img", false, 16777216},
		{16778300, "GARMIN/Activity/2024-01-15-07-00-00.fit", false, 16777252},
		{16778320, "GARMIN/Workouts/Schedule/Interval.fit", false, 16777770},
	}
	for _, tt := range tests {
		o, ok := byID[tt.id]
		if !ok {
			t.Errorf("object %d missing", tt.id)
			continue
		}
		if o.Path != tt.path || o.IsDir != tt.isDir || o.ParentID != tt.parent {
			t.Errorf("object %d = {path:%q dir:%v parent:%d}, want {path:%q dir:%v parent:%d}",
				tt.id, o.Path, o.IsDir, o.ParentID, tt.path, tt.isDir, tt.parent)
		}
	}
}

func TestBuildTreeUnknownParent(t *testing.T) {
	files := []fileEntry{{id: 5, parentID: 999, name: "orphan.fit", size: 1}}
	objects := buildTree(nil, files)
	if len(objects) != 1 || objects[0].Path != "orphan.fit" {
		t.Fatalf("unknown-parent file should keep bare name, got %+v", objects)
	}
}
