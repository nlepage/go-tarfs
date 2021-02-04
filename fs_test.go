package tarfs

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
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
		defer f.Close()

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

func TestReadDirNotDir(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"foo", "dir1/file12"} {
		if _, err := fs.ReadDir(tfs, name); !errors.Is(err, ErrNotDir) {
			t.Errorf("tarfs.ReadDir(tfs, %#v) should return ErrNotDir, got %v", name, err)
		}
	}
}

func TestReadFile(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for name, content := range map[string]string{
		"dir1/dir11/file111": "file111",
		"dir2/dir21/file212": "file212",
		"foo":                "foo",
	} {
		b, err := fs.ReadFile(tfs, name)
		if err != nil {
			t.Errorf("fs.ReadFile(tfs, %#v) should succeed, got %v", name, err)
			continue
		}

		if string(b) != content {
			t.Errorf("%s content should be %#v, got %#v", name, content, string(b))
		}
	}
}

func TestStat(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range []struct {
		path  string
		name  string
		isDir bool
	}{
		{"dir1/dir11/file111", "file111", false},
		{"foo", "foo", false},
		{"dir2/dir21", "dir21", true},
	} {
		fi, err := fs.Stat(tfs, file.path)
		if err != nil {
			t.Errorf("fs.Stat(tfs, %#v) should succeed, got %v", file.path, err)
			continue
		}

		if fi.Name() != file.name {
			t.Errorf("FileInfo.Name() should be %#v, got %#v", file.name, fi.Name())
		}

		if fi.IsDir() != file.isDir {
			t.Errorf("FileInfo.IsDir() should be %t, got %t", file.isDir, fi.IsDir())
		}
	}
}

func TestGlob(t *testing.T) {
	f, err := os.Open("test.tar")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tfs, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	for pattern, expected := range map[string][]string{
		"*/*2*":   {"dir1/file12", "dir2/dir21"},
		"*":       {"bar", "dir1", "dir2", "foo"},
		"*/*/*":   {"dir1/dir11/file111", "dir2/dir21/file211", "dir2/dir21/file212"},
		"*/*/*/*": nil,
	} {
		actual, err := fs.Glob(tfs, pattern)
		if err != nil {
			t.Errorf("fs.Glob(tfs, %#v) should succeed, got %v", pattern, err)
			continue
		}

		assert.ElementsMatchf(t, expected, actual, "matches for pattern %#v should be %#v, got %#v", pattern, expected, actual)
	}
}

func TestSub(t *testing.T) {
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
		subfs, err := fs.Sub(tfs, dir.name)
		if err != nil {
			t.Errorf("fs.Sub(tfs, %#v) should succeed, got %v", dir.name, err)
			continue
		}

		entries, err := fs.ReadDir(subfs, ".")
		if err != nil {
			t.Errorf("fs.ReadDir(subfs, %#v) should succeed, got %v", dir.name, err)
			continue
		}

		if len(entries) != dir.entriesLen {
			t.Errorf("len(entries) != %d for %#v, got %d", dir.entriesLen, dir.name, len(entries))
		}
	}
}
