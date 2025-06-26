package tarfs

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stdTarTestFiles(ctx context.Context) ([]string, error) {
	testdataDir, err := stdTarTestDir(ctx)
	if err != nil {
		return nil, err
	}
	dirents, err := os.ReadDir(testdataDir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(dirents))
	for _, dirent := range dirents {
		if !dirent.Type().IsRegular() || !strings.HasSuffix(dirent.Name(), ".tar") {
			continue
		}
		names = append(names, filepath.Join(testdataDir, dirent.Name()))
	}
	return names, nil
}

func stdTarTestDir(ctx context.Context) (string, error) {
	gorootBytes, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
	if err != nil {
		var cmdErr *exec.ExitError
		if errors.As(err, &cmdErr) {
			return "", fmt.Errorf("%w: %s", err, string(cmdErr.Stderr))
		}
		return "", err
	}
	goroot := strings.TrimSpace(string(gorootBytes))
	return filepath.Join(goroot, "src", "archive", "tar", "testdata"), nil
}

func collectTarContents(tarPath string) (map[string][]byte, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	collected := make(map[string][]byte)
	tr := tar.NewReader(f)
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return collected, err
		}
		if h.Typeflag == tar.TypeReg || h.Typeflag == tar.TypeRegA {
			bin, err := io.ReadAll(tr)
			if err != nil {
				return collected, err
			}
			collected[path.Clean(h.Name)] = bin
		}
	}
	return collected, nil
}

func colllectTarFs(fsys fs.FS) (map[string][]byte, error) {
	collected := make(map[string][]byte)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." || d.IsDir() || err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		bin, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		collected[path] = bin
		return nil
	})
	return collected, err
}

func TestFs_testdata_from_stdlib(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	testdataFilenames, err := stdTarTestFiles(ctx)

	require.NoError(err, "when reading $(go env GOROOT)/src/archive/tar/testdata")

	for _, testFilename := range testdataFilenames {
		baseName := filepath.Base(testFilename)
		if baseName == "gnu-sparse-big.tar" || baseName == "pax-sparse-big.tar" || baseName == "gnu-not-utf8.tar" {
			// big files take too long time to complete.
			// paths in not-utf8 can't pass fs.ValidPath test.
			continue
		}
		fromArchiveTar, err := collectTarContents(testFilename)
		if err != nil {
			// something that archive/tar can't handle
			continue
		}
		t.Run(baseName, func(t *testing.T) {
			f, err := os.Open(testFilename)
			require.NoError(err, "when opening %q", testFilename)
			defer f.Close()
			fsys, err := New(f)
			require.NoError(err, "when New(%q)", testFilename)
			fromTarFs, err := colllectTarFs(fsys)
			require.NoError(err, "reading %q with *Fs", testFilename)
			for name := range fromArchiveTar {
				// basically archive/tar is assumed correct.
				assert.Equal(fromArchiveTar[name], fromTarFs[name])
			}
		})
	}
}
