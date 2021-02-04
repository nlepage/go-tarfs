package tarfs

import (
	"errors"
	"io/fs"
)

// Generic errors
var (
	ErrNotDir = errors.New("not a directory")
	ErrDir    = errors.New("is a directory")
)

func newErrNotDir(op, name string) error {
	return &fs.PathError{Op: op, Path: name, Err: ErrNotDir}
}

func newErrDir(op, name string) error {
	return &fs.PathError{Op: op, Path: name, Err: ErrDir}
}
