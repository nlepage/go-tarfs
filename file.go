package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
)

type file struct {
	io.Reader
	h *tar.Header
}

func newFile(e entry) *file {
	return &file{bytes.NewReader(e.b), e.h}
}

var _ fs.File = &file{}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.h.FileInfo(), nil
}

func (f *file) Close() error {
	return nil
}
