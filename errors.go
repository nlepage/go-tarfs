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

func newErrNotDir(op, path string) error {
	return newErr(op, path, ErrNotDir)
}

func newErrDir(op, path string) error {
	return newErr(op, path, ErrDir)
}

func newErrClosed(op, path string) error {
	return newErr(op, path, fs.ErrClosed)
}

func newErrNotExist(op, path string) error {
	return newErr(op, path, fs.ErrNotExist)
}

func newErr(op, path string, err error) error {
	return &fs.PathError{Op: op, Path: path, Err: err}
}
