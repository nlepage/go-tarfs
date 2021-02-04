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

	for _, name := range []string{"/foo", "./foo", "foo/", "foo/../foo", "foo//bar"} {
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
	defer f.Close()

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
	defer f.Close()

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

func TestReadDir(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range []struct {
		name       string
		entriesLen int
	}{
		{".", 4},
		{"dir1", 3},
		{"dir2/dir21", 2},
	} {
		entries, err := fs.ReadDir(tfs, dir.name)
		if err != nil {
			t.Errorf("fs.ReadDir(tfs, %#v) should succeed, got %v", dir.name, err)
			continue
		}

		if len(entries) != dir.entriesLen {
			t.Errorf("len(entries) != %d for %#v, got %d", dir.entriesLen, dir.name, len(entries))
		}
	}

}
