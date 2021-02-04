package tarfs

import (
	"bytes"
	"io"
	"io/fs"
	"time"
)

type file struct {
	entry
	r          io.Reader
	readDirPos int
}

func newFile(e entry) *file {
	return &file{e, bytes.NewReader(e.b), 0}
}

var _ fs.File = &file{}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.h.FileInfo(), nil
}

func (f *file) Read(b []byte) (int, error) {
	if f.IsDir() {
		return 0, newErrDir("read", f.Name())
	}

	return f.r.Read(b)
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

type rootFile struct{}

var _ fs.File = &rootFile{}

func (rf *rootFile) Stat() (fs.FileInfo, error) {
	return rf, nil
}

func (*rootFile) Read([]byte) (int, error) {
	return 0, newErrDir("read", ".")
}

func (*rootFile) Close() error {
	return nil
}

var _ fs.FileInfo = &rootFile{}

func (rf *rootFile) Name() string {
	return "."
}

func (rf *rootFile) Size() int64 {
	return 0
}

func (rf *rootFile) Mode() fs.FileMode {
	return fs.FileMode(fs.ModeDir | 0755)
}

func (rf *rootFile) ModTime() time.Time {
	return time.Time{}
}

func (rf *rootFile) IsDir() bool {
	return true
}

func (rf *rootFile) Sys() interface{} {
	return nil
}
