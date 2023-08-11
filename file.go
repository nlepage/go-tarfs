package tarfs

import (
	"io"
	"io/fs"
)

type file struct {
	entry
	r          io.ReadSeeker
	readDirPos int
	closed     bool
}

var _ fs.File = &file{}

func (f *file) Stat() (fs.FileInfo, error) {
	const op = "stat"

	if f.closed {
		return nil, newErrClosed(op, f.Name())
	}

	return f.Info()
}

func (f *file) Read(b []byte) (int, error) {
	const op = "read"

	if f.closed {
		return 0, newErrClosed(op, f.Name())
	}

	if f.IsDir() {
		return 0, newErrDir(op, f.Name())
	}

	return f.r.Read(b)
}

func (f *file) Close() error {
	const op = "close"

	if f.closed {
		return newErrClosed(op, f.Name())
	}

	f.r = nil
	f.closed = true

	return nil
}

var _ io.Seeker = &file{}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	const op = "seek"

	if f.closed {
		return 0, newErrClosed(op, f.Name())
	}

	if f.IsDir() {
		return 0, newErrDir(op, f.Name())
	}

	return f.r.Seek(offset, whence)
}

var _ fs.ReadDirFile = &file{}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	const op = "readdir"

	if f.closed {
		return nil, newErrClosed(op, f.Name())
	}

	allEntries, err := f.entry.entries(op, f.Name())
	if err != nil {
		return nil, err
	}

	if f.readDirPos >= len(allEntries) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	if n <= 0 || f.readDirPos+n > len(allEntries) {
		n = len(allEntries) - f.readDirPos
	}

	entries := make([]fs.DirEntry, n)

	copy(entries, allEntries[f.readDirPos:])

	f.readDirPos += n

	return entries, nil
}
