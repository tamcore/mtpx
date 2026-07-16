// Package vfs provides read-only navigation over the flat object list returned
// by a backend, treating the objects' slash-separated paths as a tree.
package vfs

import (
	"slices"
	"strings"

	"github.com/tamcore/mtpx/internal/backend"
)

// parentDir returns the directory portion of a slash path, or "" for a
// top-level entry.
func parentDir(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return ""
	}
	return p[:i]
}

// Children returns the direct children of dir, directories first and then files,
// each group sorted by name. Pass "" for the top level.
func Children(objects []backend.Object, dir string) []backend.Object {
	var out []backend.Object
	for _, o := range objects {
		if parentDir(o.Path) == dir {
			out = append(out, o)
		}
	}
	slices.SortFunc(out, func(a, b backend.Object) int {
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out
}

// Find returns the object at the given path, if present.
func Find(objects []backend.Object, path string) (backend.Object, bool) {
	for _, o := range objects {
		if o.Path == path {
			return o, true
		}
	}
	return backend.Object{}, false
}

// FilesUnder returns every file (never a directory) nested under dir at any
// depth, sorted by path.
func FilesUnder(objects []backend.Object, dir string) []backend.Object {
	prefix := dir + "/"
	var out []backend.Object
	for _, o := range objects {
		if o.IsDir {
			continue
		}
		if strings.HasPrefix(o.Path, prefix) {
			out = append(out, o)
		}
	}
	slices.SortFunc(out, func(a, b backend.Object) int {
		return strings.Compare(a.Path, b.Path)
	})
	return out
}
