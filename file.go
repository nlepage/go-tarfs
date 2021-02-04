package tarfs

import (
	"bytes"
	"io"
	"io/fs"
)

type file struct {
	entry
	io.Reader
	readDirPos int
}

func newFile(e entry) *file {
	return &file{e, bytes.NewReader(e.b), 0}
}

var _ fs.File = &file{}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.h.FileInfo(), nil
}

func (f *file) Close() error {
	return nil
}

var _ fs.ReadDirFile = &file{}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, newErrNotDir("readdir", f.Name())
	}

	if n <= 0 {
		return f.entries, nil
	}

	if f.readDirPos == len(f.entries) {
		return nil, io.EOF
	}

	start, end := f.readDirPos, f.readDirPos+n
	if end > len(f.entries) {
		end = len(f.entries)
	}
	f.readDirPos = end

	return f.entries[start:end], nil
}
