package tarfs

import (
	"errors"
	"io/fs"
)

var (
	// ErrNotDir may be returned by fs.ReadDir()
	ErrNotDir = errors.New("not a directory")
)

func newErrNotDir(op, name string) error {
	return &fs.PathError{Op: "readdir", Path: name, Err: ErrNotDir}
}
