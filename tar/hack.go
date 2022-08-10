package tar

import (
	"fmt"
	"io"
)

type SectionReader interface {
	Read(p []byte) (n int, err error)
	ReadAt(p []byte, off int64) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	// Size() int64
}

// SectionReader returns the current section reader.
func (tr *Reader) SectionReader() (SectionReader, error) {
	ra, ok := tr.r.(SectionReader)
	if !ok {
		return nil, fmt.Errorf("expected an SectionReader, got: %T", tr.r)
	}

	// TODO: keep it valid even after some Read already occured?
	off, err := ra.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	switch fr := tr.curr.(type) {
	case *regFileReader:
		return io.NewSectionReader(ra, off, fr.nb), nil
	// TODO: support *sparseFileReader
	default:
		return nil, fmt.Errorf("unsupported fileReader type: %T", fr)
	}
}
