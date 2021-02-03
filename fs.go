package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"path"
)

type tarfs struct {
	files map[string]entry
}

type entry struct {
	h *tar.Header
	b []byte
}

// New creates a new tar fs.FS from r
func New(r io.Reader) (fs.FS, error) {
	tfs := &tarfs{make(map[string]entry)}
	tr := tar.NewReader(r)

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		b := make([]byte, int(h.Size))
		if _, err := io.Copy(bytes.NewBuffer(b), tr); err != nil {
			return nil, err
		}

		tfs.files[path.Clean(h.Name)] = entry{h, b}
	}

	return tfs, nil
}

var _ fs.FS = &tarfs{}

func (tfs *tarfs) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	e, ok := tfs.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return newFile(e), nil
}
