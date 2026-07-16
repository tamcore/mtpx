package vfs

import (
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func sample() []backend.Object {
	return []backend.Object{
		{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
		{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
		{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit", Size: 10},
		{ID: 4, Path: "GARMIN/Activity/b.fit", Name: "b.fit", Size: 20},
		{ID: 5, Path: "GARMIN/Workouts", Name: "Workouts", IsDir: true},
		{ID: 6, Path: "GARMIN/z.txt", Name: "z.txt", Size: 5},
		{ID: 7, Path: "Music", Name: "Music", IsDir: true},
	}
}

func paths(objs []backend.Object) []string {
	out := make([]string, len(objs))
	for i, o := range objs {
		out[i] = o.Path
	}
	return out
}

func eq(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestChildrenRoot(t *testing.T) {
	got := Children(sample(), "")
	eq(t, paths(got), []string{"GARMIN", "Music"})
}

func TestChildrenDirsBeforeFiles(t *testing.T) {
	got := Children(sample(), "GARMIN")
	eq(t, paths(got), []string{"GARMIN/Activity", "GARMIN/Workouts", "GARMIN/z.txt"})
}

func TestChildrenLeaf(t *testing.T) {
	got := Children(sample(), "GARMIN/Activity")
	eq(t, paths(got), []string{"GARMIN/Activity/a.fit", "GARMIN/Activity/b.fit"})
}

func TestFind(t *testing.T) {
	o, ok := Find(sample(), "GARMIN/Activity")
	if !ok || !o.IsDir || o.ID != 2 {
		t.Fatalf("Find = %+v ok=%v", o, ok)
	}
	if _, ok := Find(sample(), "does/not/exist"); ok {
		t.Fatal("want not found")
	}
}

func TestFilesUnder(t *testing.T) {
	got := FilesUnder(sample(), "GARMIN/Activity")
	eq(t, paths(got), []string{"GARMIN/Activity/a.fit", "GARMIN/Activity/b.fit"})
}

func TestFilesUnderRecursive(t *testing.T) {
	got := FilesUnder(sample(), "GARMIN")
	eq(t, paths(got), []string{"GARMIN/Activity/a.fit", "GARMIN/Activity/b.fit", "GARMIN/z.txt"})
}

func TestFilesUnderEmpty(t *testing.T) {
	if got := FilesUnder(sample(), "GARMIN/Workouts"); len(got) != 0 {
		t.Fatalf("want no files, got %v", paths(got))
	}
}
