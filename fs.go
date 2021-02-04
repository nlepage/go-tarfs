package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
)

type tarfs struct {
	files map[string]*entry
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
	// Root is a directory
	if e.h == nil {
		return true
	}

	return e.h.FileInfo().IsDir()
}

func (e *entry) Type() fs.FileMode {
	return e.h.FileInfo().Mode()
}

func (e *entry) Info() (fs.FileInfo, error) {
	return e.h.FileInfo(), nil
}

// New creates a new tar fs.FS from r
func New(r io.Reader) (fs.FS, error) {
	tr := tar.NewReader(r)
	tfs := &tarfs{make(map[string]*entry)}

	tfs.files["."] = &entry{}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := filepath.Clean(h.Name)

		b := make([]byte, int(h.Size))
		if _, err := io.Copy(bytes.NewBuffer(b), tr); err != nil {
			return nil, err
		}

		e := &entry{h, b, nil}

		tfs.files[name] = e

		if parent, ok := tfs.files[filepath.Dir(name)]; ok {
			parent.entries = append(parent.entries, e)
		}
	}

	return tfs, nil
}

var _ fs.FS = &tarfs{}

func (tfs *tarfs) get(name string) (*entry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	e, ok := tfs.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return e, nil
}

func (tfs *tarfs) Open(name string) (fs.File, error) {
	e, err := tfs.get(name)
	if err != nil {
		return nil, err
	}

	return newFile(*e), nil
}

var _ fs.ReadDirFS = &tarfs{}

func (tfs *tarfs) ReadDir(name string) ([]fs.DirEntry, error) {
	e, err := tfs.get(name)
	if err != nil {
		return nil, err
	}

	if !e.IsDir() {
		return nil, newErrNotDir("readdir", name)
	}

	return e.entries, nil
}
