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
	return makeSectionReader(tr.curr)
}

func makeSectionReader(fr fileReader) (SectionReader, error) {
	n := fr.logicalRemaining()
	switch fr := fr.(type) {

	case *regFileReader:
		ra, ok := fr.r.(SectionReader)
		if !ok {
			return nil, fmt.Errorf("expected an SectionReader, got: %T", fr.r)
		}
		if fr.off < 0 {
			return nil, fmt.Errorf("unexpected negative offset: %d", fr.off)
		}
		return io.NewSectionReader(ra, fr.off, n), nil

	case *sparseFileReader:
		underlyingRa, err := makeSectionReader(fr.fr)
		if err != nil {
			return nil, err
		}
		return io.NewSectionReader(&sparseFileReaderAt{
			fr:  underlyingRa,
			sp:  fr.sp,
			pos: 1,
		}, 0, n), nil

	default:
		return nil, fmt.Errorf("unsupported fileReader type: %T", fr)
	}
}

type sparseFileReaderAt struct {
	fr  SectionReader // Underlying fileReader
	sp  sparseHoles   // Normalized list of sparse holes
	pos int64         // Current position in sparse file
}

func (fr *sparseFileReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	physicalPos := off

	var remainingZeros int64
	nextHole := -1
	for i, hole := range fr.sp {
		holeStart, holeEnd := hole.Offset, hole.endOffset()
		if off < holeStart {
			nextHole = i
			break
		}
		if off < holeEnd {
			remainingZeros = holeEnd - off
			nextHole = i
			break
		}
		// start and end are before off: account for the hole
		physicalPos -= hole.Length
	}

	for len(b) > 0 {
		var nf int // Bytes read in fragment
		if remainingZeros > 0 {
			nf = len(b)
			if int(remainingZeros) < nf {
				nf = int(remainingZeros)
			}
			for i := range b[:nf] {
				b[i] = 0
			}
			b = b[nf:]
			n += nf
			remainingZeros = 0
			nextHole++
		} else if nextHole != -1 && nextHole < len(fr.sp) {
			// dense data up to next hole
			hole := fr.sp[nextHole]
			if off < hole.Offset {
				// read some physical data
				nf = len(b)
				if int(hole.Offset-off) < nf {
					nf = int(hole.Offset - off)
				}
				nf, err = fr.fr.ReadAt(b[:nf], physicalPos)
				b = b[nf:]
				n += nf
				physicalPos += int64(nf)
				if err != nil {
					return n, err
				}
			}
			remainingZeros = hole.Length
		} else {
			// dense data up to the end
			nf, err = fr.fr.ReadAt(b, physicalPos)
			n += nf
			return n, err
		}
	}
	return n, nil
}
