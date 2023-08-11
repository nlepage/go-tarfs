package tarfs

import (
	"errors"
	"io"
)

type readReaderAt interface {
	io.Reader
	io.ReaderAt
}

type readCounterIface interface {
	io.Reader
	Count() int64
}

type readCounter struct {
	io.Reader
	off int64
}

func (cr *readCounter) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.off += int64(n)
	return
}

func (cr *readCounter) Count() int64 {
	return cr.off
}

type readSeekCounter struct {
	io.ReadSeeker
	off int64
}

func (cr *readSeekCounter) Read(p []byte) (n int, err error) {
	n, err = cr.ReadSeeker.Read(p)
	cr.off += int64(n)
	return
}

func (cr *readSeekCounter) Seek(offset int64, whence int) (abs int64, err error) {
	abs, err = cr.ReadSeeker.Seek(offset, whence)
	cr.off = abs
	return
}

func (cr *readSeekCounter) Count() int64 {
	return cr.off
}

type readSeeker struct {
	*readCounter
	e *regEntry
}

var _ io.ReadSeeker = &readSeeker{}

func (rs *readSeeker) Seek(offset int64, whence int) (int64, error) {
	const op = "seek"

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = rs.off + offset
	case io.SeekEnd:
		abs = rs.e.size() + offset
	default:
		return 0, newErr(op, rs.e.name, errors.New("invalid whence"))
	}
	if abs < 0 {
		return 0, newErr(op, rs.e.name, errors.New("negative position"))
	}

	if abs < rs.off {
		r, err := rs.e.reader()
		if err != nil {
			return 0, err
		}

		rs.readCounter = &readCounter{r, 0}
	}

	if abs > rs.off {
		if _, err := io.CopyN(io.Discard, rs.readCounter, abs-rs.off); err != nil && err != io.EOF {
			return 0, err
		}
	}

	return abs, nil
}
