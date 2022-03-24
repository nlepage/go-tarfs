package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

type tarfs struct {
	entries     map[string]fs.DirEntry
	rootEntries []fs.DirEntry
	rootEntry   fs.DirEntry
}

type entry struct {
	h       *tar.Header
	b       []byte
	entries []fs.DirEntry
}

type entries interface {
	append(fs.DirEntry)
	get() []fs.DirEntry
}

type fileInfo interface {
	FileInfo() (fs.FileInfo, error)
}

var _ fs.DirEntry = &entry{}

func (e *entry) Name() string {
	return e.h.FileInfo().Name()
}

func (e *entry) IsDir() bool {
	return e.h.FileInfo().IsDir()
}

func (e *entry) Type() fs.FileMode {
	return e.h.FileInfo().Mode() & fs.ModeType
}

func (e *entry) Info() (fs.FileInfo, error) {
	return e.h.FileInfo(), nil
}

var _ entries = &entry{}

func (e *entry) append(c fs.DirEntry) {
	e.entries = append(e.entries, c)
}

func (e *entry) get() []fs.DirEntry {
	return e.entries
}

var _ fileInfo = &entry{}

func (e *entry) FileInfo() (fs.FileInfo, error) {
	return e.h.FileInfo(), nil
}

type fakeDirEntry struct {
	name    string
	entries []fs.DirEntry
}

var _ fs.DirEntry = &fakeDirEntry{}

func (e *fakeDirEntry) Name() string {
	return e.name
}

func (*fakeDirEntry) IsDir() bool {
	return true
}

func (*fakeDirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (e *fakeDirEntry) Info() (fs.FileInfo, error) {
	return e, nil
}

var _ fs.FileInfo = &fakeDirEntry{}

func (*fakeDirEntry) Mode() fs.FileMode {
	return fs.ModeDir
}

func (*fakeDirEntry) Size() int64 {
	return 0
}

func (*fakeDirEntry) ModTime() time.Time {
	return time.Time{}
}

func (*fakeDirEntry) Sys() interface{} {
	return nil
}

var _ entries = &fakeDirEntry{}

func (e *fakeDirEntry) append(c fs.DirEntry) {
	e.entries = append(e.entries, c)
}

func (e *fakeDirEntry) get() []fs.DirEntry {
	return e.entries
}

var _ fileInfo = &fakeDirEntry{}

func (e *fakeDirEntry) FileInfo() (fs.FileInfo, error) {
	return e, nil
}

// New creates a new tar fs.FS from r
func New(r io.Reader) (fs.FS, error) {
	tr := tar.NewReader(r)
	tfs := &tarfs{make(map[string]fs.DirEntry), make([]fs.DirEntry, 0, 10), nil}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := path.Clean(h.Name)
		if name == "." {
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, int(h.Size)))
		if _, err := io.Copy(buf, tr); err != nil {
			return nil, err
		}

		e := &entry{h, buf.Bytes(), nil}

		tfs.append(name, e)
	}

	return tfs, nil
}

func (tfs *tarfs) append(name string, e fs.DirEntry) {
	tfs.entries[name] = e

	dir := path.Dir(name)

	if dir == "." {
		tfs.rootEntries = append(tfs.rootEntries, e)
		return
	}

	if parent, ok := tfs.entries[dir]; ok {
		parent := parent.(entries)
		parent.append(e)
		return
	}

	parent := &fakeDirEntry{path.Base(dir), nil}

	tfs.append(dir, parent)

	parent.append(e)
}

var _ fs.FS = &tarfs{}

func (tfs *tarfs) get(name, op string) (fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	e, ok := tfs.entries[name]
	if !ok {
		return nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
	}

	return e, nil
}

func (tfs *tarfs) Open(name string) (fs.File, error) {
	if name == "." {
		if tfs.rootEntry == nil {
			return &rootFile{tfs: tfs}, nil
		}
		return newFile(tfs.rootEntry), nil
	}

	e, err := tfs.get(name, "open")
	if err != nil {
		return nil, err
	}

	return newFile(e), nil
}

var _ fs.ReadDirFS = &tarfs{}

func (tfs *tarfs) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "." {
		return tfs.rootEntries, nil
	}

	e, err := tfs.get(name, "readdir")
	if err != nil {
		return nil, err
	}

	if !e.IsDir() {
		return nil, newErrNotDir("readdir", name)
	}

	entries := e.(entries).get()

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	return entries, nil
}

var _ fs.ReadFileFS = &tarfs{}

func (tfs *tarfs) ReadFile(name string) ([]byte, error) {
	if name == "." {
		return nil, newErrDir("readfile", name)
	}

	e, err := tfs.get(name, "readfile")
	if err != nil {
		return nil, err
	}

	if e.IsDir() {
		return nil, newErrDir("readfile", name)
	}

	ee := e.(*entry)

	buf := make([]byte, len(ee.b))
	copy(buf, ee.b)
	return buf, nil
}

var _ fs.StatFS = &tarfs{}

func (tfs *tarfs) Stat(name string) (fs.FileInfo, error) {
	if name == "." {
		if tfs.rootEntry == nil {
			return &rootFile{tfs: tfs}, nil
		}
		return tfs.rootEntry.Info()
	}

	e, err := tfs.get(name, "stat")
	if err != nil {
		return nil, err
	}

	return e.(fileInfo).FileInfo()
}

var _ fs.GlobFS = &tarfs{}

func (tfs *tarfs) Glob(pattern string) (matches []string, _ error) {
	for name := range tfs.entries {
		match, err := path.Match(pattern, name)
		if err != nil {
			return nil, err
		}
		if match {
			matches = append(matches, name)
		}
	}
	return
}

var _ fs.SubFS = &tarfs{}

func (tfs *tarfs) Sub(dir string) (fs.FS, error) {
	if dir == "." {
		return tfs, nil
	}

	e, err := tfs.get(dir, "sub")
	if err != nil {
		return nil, err
	}

	if !e.IsDir() {
		return nil, newErrNotDir("sub", dir)
	}

	subfs := &tarfs{make(map[string]fs.DirEntry), e.(entries).get(), e}
	prefix := dir + "/"
	for name, file := range tfs.entries {
		if strings.HasPrefix(name, prefix) {
			subfs.entries[strings.TrimPrefix(name, prefix)] = file
		}
	}

	return subfs, nil
}
