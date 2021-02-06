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

func (f *file) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if !f.IsDir() {
		return nil, newErrNotDir("readdir", f.Name())
	}

	if f.readDirPos >= len(f.entries) {
		if n <= 0 {
			return nil, nil
		} else {
			return nil, io.EOF
		}
	}

	if n > 0 && f.readDirPos+n <= len(f.entries) {
		entries = f.entries[f.readDirPos : f.readDirPos+n]
		f.readDirPos += n
	} else {
		entries = f.entries[f.readDirPos:]
		f.readDirPos += len(entries)
	}

	return entries, err
}

type rootFile struct {
	tfs        *tarfs
	readDirPos int
}

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

var _ fs.ReadDirFile = &rootFile{}

func (rf *rootFile) ReadDir(n int) ([]fs.DirEntry, error) {
	entries, err := rf.tfs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	if rf.readDirPos >= len(entries) {
		if n <= 0 {
			return nil, nil
		} else {
			return nil, io.EOF
		}
	}

	if n > 0 && rf.readDirPos+n <= len(entries) {
		entries = entries[rf.readDirPos : rf.readDirPos+n]
		rf.readDirPos += n
	} else {
		entries = entries[rf.readDirPos:]
		rf.readDirPos += len(entries)
	}

	return entries, err
}

var _ fs.FileInfo = &rootFile{}

func (rf *rootFile) Name() string {
	return "."
}

func (rf *rootFile) Size() int64 {
	return 0
}

func (rf *rootFile) Mode() fs.FileMode {
	return fs.ModeDir | 0755
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
