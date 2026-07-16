package backend

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
)

// The parsers below turn libmtp CLI output into structured data. They only act
// on lines that match the expected shapes, so connection banners, warnings and
// other noise the tools print are ignored.

var (
	detectRawRe     = regexp.MustCompile(`^\s+(.+): (.+) \(([0-9a-fA-F]{4}):([0-9a-fA-F]{4})\) @ bus (\d+), dev (\d+)$`)
	detectRawAltRe  = regexp.MustCompile(`^\s+([0-9a-fA-F]{4}):([0-9a-fA-F]{4}) @ bus (\d+), dev (\d+)$`)
	detectSerialRe  = regexp.MustCompile(`^\s+Serial number: (.+)$`)
	detectModelRe   = regexp.MustCompile(`^\s+Model: (.+)$`)
	detectManufRe   = regexp.MustCompile(`^\s+Manufacturer: (.+)$`)
	folderLineRe    = regexp.MustCompile(`^(\d+)\t( *)(.*)$`)
	fileIDRe        = regexp.MustCompile(`^File ID: (\d+)$`)
	fileNameRe      = regexp.MustCompile(`^\s+Filename: (.*)$`)
	fileSizeRe      = regexp.MustCompile(`^\s+File size (\d+) \(0x[0-9A-Fa-f]+\) bytes$`)
	fileAbstractRe  = regexp.MustCompile(`^\s+None\. \(abstract file, size = -1\)$`)
	fileParentRe    = regexp.MustCompile(`^\s+Parent ID: (\d+)$`)
	fileStorageIDRe = regexp.MustCompile(`^\s+Storage ID: 0x([0-9A-Fa-f]+)$`)
)

type folderEntry struct {
	id    uint32
	level int
	name  string
}

type fileEntry struct {
	id        uint32
	parentID  uint32
	storageID uint32
	name      string
	size      int64
}

func scanLines(out []byte) *bufio.Scanner {
	s := bufio.NewScanner(bytes.NewReader(out))
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return s
}

// parseDetect extracts the attached devices from `mtp-detect` output. The
// serial, model and manufacturer reported for the connected device are attached
// to the first device found.
func parseDetect(out []byte) []Device {
	var devices []Device
	var serial, model, manuf string

	s := scanLines(out)
	for s.Scan() {
		line := s.Text()
		if m := detectRawRe.FindStringSubmatch(line); m != nil {
			devices = append(devices, Device{
				Vendor:    m[1],
				Product:   m[2],
				VendorID:  parseHex16(m[3]),
				ProductID: parseHex16(m[4]),
				Bus:       atoi(m[5]),
				Dev:       atoi(m[6]),
			})
			continue
		}
		if m := detectRawAltRe.FindStringSubmatch(line); m != nil {
			devices = append(devices, Device{
				VendorID:  parseHex16(m[1]),
				ProductID: parseHex16(m[2]),
				Bus:       atoi(m[3]),
				Dev:       atoi(m[4]),
			})
			continue
		}
		if m := detectSerialRe.FindStringSubmatch(line); m != nil {
			serial = m[1]
		}
		if m := detectModelRe.FindStringSubmatch(line); m != nil {
			model = m[1]
		}
		if m := detectManufRe.FindStringSubmatch(line); m != nil {
			manuf = m[1]
		}
	}

	if len(devices) > 0 {
		devices[0].Serial = serial
		if devices[0].Product == "" {
			devices[0].Product = model
		}
		if devices[0].Vendor == "" {
			devices[0].Vendor = manuf
		}
	}
	return devices
}

// parseFolders extracts the folder entries from `mtp-folders` output. Each entry
// carries its nesting level, derived from two spaces of indentation per level.
func parseFolders(out []byte) []folderEntry {
	var entries []folderEntry
	s := scanLines(out)
	for s.Scan() {
		m := folderLineRe.FindStringSubmatch(s.Text())
		if m == nil {
			continue
		}
		entries = append(entries, folderEntry{
			id:    parseUint32(m[1]),
			level: len(m[2]) / 2,
			name:  m[3],
		})
	}
	return entries
}

// parseFiles extracts the file entries from `mtp-files` output.
func parseFiles(out []byte) []fileEntry {
	var entries []fileEntry
	var cur *fileEntry
	flush := func() {
		if cur != nil {
			entries = append(entries, *cur)
			cur = nil
		}
	}

	s := scanLines(out)
	for s.Scan() {
		line := s.Text()
		if m := fileIDRe.FindStringSubmatch(line); m != nil {
			flush()
			cur = &fileEntry{id: parseUint32(m[1])}
			continue
		}
		if cur == nil {
			continue
		}
		switch {
		case fileNameRe.MatchString(line):
			cur.name = fileNameRe.FindStringSubmatch(line)[1]
		case fileSizeRe.MatchString(line):
			cur.size = parseInt64(fileSizeRe.FindStringSubmatch(line)[1])
		case fileAbstractRe.MatchString(line):
			cur.size = -1
		case fileParentRe.MatchString(line):
			cur.parentID = parseUint32(fileParentRe.FindStringSubmatch(line)[1])
		case fileStorageIDRe.MatchString(line):
			cur.storageID = parseHex32(fileStorageIDRe.FindStringSubmatch(line)[1])
		}
	}
	flush()
	return entries
}

// buildTree combines parsed folders and files into a flat list of objects with
// full slash-separated paths. Folders come first, in listing order, followed by
// files.
func buildTree(folders []folderEntry, files []fileEntry) []Object {
	folderPath := make(map[uint32]string, len(folders))
	objects := make([]Object, 0, len(folders)+len(files))

	var stack []folderEntry
	for _, fe := range folders {
		for len(stack) > 0 && stack[len(stack)-1].level >= fe.level {
			stack = stack[:len(stack)-1]
		}
		var parentID uint32
		var parentPath string
		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			parentID = parent.id
			parentPath = folderPath[parent.id]
		}
		full := join(parentPath, fe.name)
		folderPath[fe.id] = full
		objects = append(objects, Object{
			ID:       fe.id,
			ParentID: parentID,
			Name:     fe.name,
			Path:     full,
			IsDir:    true,
		})
		stack = append(stack, fe)
	}

	for _, f := range files {
		full := join(folderPath[f.parentID], f.name)
		objects = append(objects, Object{
			ID:        f.id,
			ParentID:  f.parentID,
			StorageID: f.storageID,
			Name:      f.name,
			Path:      full,
			Size:      f.size,
		})
	}
	return objects
}

func join(dir, name string) string {
	if dir == "" {
		return name
	}
	return dir + "/" + name
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseUint32(s string) uint32 {
	n, _ := strconv.ParseUint(s, 10, 32)
	return uint32(n)
}

func parseHex16(s string) uint16 {
	n, _ := strconv.ParseUint(s, 16, 16)
	return uint16(n)
}

func parseHex32(s string) uint32 {
	n, _ := strconv.ParseUint(s, 16, 32)
	return uint32(n)
}
