package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type entry interface {
	fs.DirEntry
	size() int64
	readdir(path string) ([]fs.DirEntry, error)
	readfile(path string) ([]byte, error)
	entries(op, path string) ([]fs.DirEntry, error)
	open() (fs.File, error)
}

type regEntry struct {
	fs.DirEntry
	name   string
	ra     io.ReaderAt
	offset int64
}

var _ entry = &regEntry{}

func (e *regEntry) size() int64 {
	info, _ := e.Info() // err is necessarily nil
	return info.Size()
}

func (e *regEntry) readdir(path string) ([]fs.DirEntry, error) {
	return nil, newErrNotDir("readdir", path)
}

func (e *regEntry) readfile(path string) ([]byte, error) {
	r, err := e.reader()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(make([]byte, 0, e.size()))

	if _, err := io.Copy(b, r); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e *regEntry) entries(op, path string) ([]fs.DirEntry, error) {
	return nil, newErrNotDir(op, path)
}

func (e *regEntry) open() (fs.File, error) {
	r, err := e.reader()
	if err != nil {
		return nil, err
	}

	return &file{e, &readSeeker{&readCounter{r, 0}, e}, -1, false}, nil
}

func (e *regEntry) reader() (io.Reader, error) {
	tr := tar.NewReader(io.NewSectionReader(e.ra, e.offset, 1<<63-1-e.offset))

	if _, err := tr.Next(); err != nil {
		return nil, err
	}

	return tr, nil
}

type dirEntry struct {
	fs.DirEntry
	_entries []fs.DirEntry
	sorted   bool
}

func newDirEntry(e fs.DirEntry) *dirEntry {
	return &dirEntry{e, make([]fs.DirEntry, 0, 10), false}
}

func (e *dirEntry) append(c fs.DirEntry) {
	e._entries = append(e._entries, c)
}

var _ entry = &dirEntry{}

func (e *dirEntry) size() int64 {
	return 0
}

func (e *dirEntry) readdir(path string) ([]fs.DirEntry, error) {
	if !e.sorted {
		sort.Sort(entriesByName(e._entries))
	}

	entries := make([]fs.DirEntry, len(e._entries))

	copy(entries, e._entries)

	return entries, nil
}

func (e *dirEntry) readfile(path string) ([]byte, error) {
	return nil, newErrDir("readfile", path)
}

func (e *dirEntry) entries(op, path string) ([]fs.DirEntry, error) {
	if !e.sorted {
		sort.Sort(entriesByName(e._entries))
	}

	return e._entries, nil
}

func (e *dirEntry) open() (fs.File, error) {
	return &file{e, nil, 0, false}, nil
}

type fakeDirFileInfo string

var _ fs.FileInfo = fakeDirFileInfo("")

func (e fakeDirFileInfo) Name() string {
	return string(e)
}

func (fakeDirFileInfo) Size() int64 {
	return 0
}

func (fakeDirFileInfo) Mode() fs.FileMode {
	return fs.ModeDir
}

func (fakeDirFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fakeDirFileInfo) IsDir() bool {
	return true
}

func (fakeDirFileInfo) Sys() interface{} {
	return nil
}

type entriesByName []fs.DirEntry

var _ sort.Interface = entriesByName{}

func (entries entriesByName) Less(i, j int) bool {
	return strings.Compare(entries[i].Name(), entries[j].Name()) < 0
}

func (entries entriesByName) Len() int {
	return len(entries)
}

func (entries entriesByName) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}
