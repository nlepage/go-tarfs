package tarfs

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"
)

func TestOpenInvalid(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"/foo", "foo/", "foo/../foo", "foo//bar"} {
		if _, err := tfs.Open(name); !errors.Is(err, fs.ErrInvalid) {
			t.Errorf("tarfs.Open(%#v) should return fs.ErrInvalid, got %v", name, err)
		}
	}
}

func TestOpenNotExist(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"baz", "qwe", "foo/bar", "file11"} {
		if _, err := tfs.Open(name); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("tarfs.Open(%#v) should return fs.ErrNotExist, got %v", name, err)
		}
	}
}

func TestOpen(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"foo", "bar", "dir1", "dir1/file11"} {
		f, err := tfs.Open(name)
		if err != nil {
			t.Errorf("tarfs.Open(%#v) should succeed, got %v", name, err)
			continue
		}

		fi, err := f.Stat()
		if err != nil {
			t.Errorf("file{%#v}.Stat() should succeed, got %v", name, err)
			continue
		}

		if fi.Name() != path.Base(name) {
			t.Errorf("FileInfo.Name() is %#v, expected %#v", fi.Name(), path.Base(name))
		}
	}
}
