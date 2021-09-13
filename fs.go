package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type tarfs struct {
	files       map[string]*entry
	rootEntries []fs.DirEntry
	rootEntry   *entry
}

type entry struct {
	h       *tar.Header
	b       []byte
	entries []fs.DirEntry
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

// New creates a new tar fs.FS from r
func New(r io.Reader) (fs.FS, error) {
	tr := tar.NewReader(r)
	tfs := &tarfs{make(map[string]*entry), make([]fs.DirEntry, 0, 10), nil}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := filepath.Clean(h.Name)

		buf := bytes.NewBuffer(make([]byte, 0, int(h.Size)))
		if _, err := io.Copy(buf, tr); err != nil {
			return nil, err
		}

		e := &entry{h, buf.Bytes(), nil}

		tfs.files[name] = e

		dir := filepath.Dir(name)
		if dir == "." {
			tfs.rootEntries = append(tfs.rootEntries, e)
		} else {
			if parent, ok := tfs.files[filepath.Dir(name)]; ok {
				parent.entries = append(parent.entries, e)
			}
		}
	}

	return tfs, nil
}

var _ fs.FS = &tarfs{}

func (tfs *tarfs) get(name, op string) (*entry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	e, ok := tfs.files[name]
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
		return newFile(*tfs.rootEntry), nil
	}

	e, err := tfs.get(name, "open")
	if err != nil {
		return nil, err
	}

	return newFile(*e), nil
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

	sort.Slice(e.entries, func(i, j int) bool { return e.entries[i].Name() < e.entries[j].Name() })

	return e.entries, nil
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

	buf := make([]byte, len(e.b))
	copy(buf, e.b)
	return buf, nil
}

var _ fs.StatFS = &tarfs{}

func (tfs *tarfs) Stat(name string) (fs.FileInfo, error) {
	e, err := tfs.get(name, "stat")
	if err != nil {
		return nil, err
	}

	return e.h.FileInfo(), nil
}

var _ fs.GlobFS = &tarfs{}

func (tfs *tarfs) Glob(pattern string) (matches []string, _ error) {
	for name := range tfs.files {
		match, err := filepath.Match(pattern, name)
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

	subfs := &tarfs{make(map[string]*entry), e.entries, e}
	prefix := dir + "/"
	for name, file := range tfs.files {
		if strings.HasPrefix(name, prefix) {
			subfs.files[strings.TrimPrefix(name, prefix)] = file
		}
	}

	return subfs, nil
}
