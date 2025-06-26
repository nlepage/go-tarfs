package tarfs

import (
	"archive/tar"
	"io"
)

func reconstructSparse(sr *io.SectionReader, hdr *tar.Header, blk *block) (sparseHoles, error) {
	if hdr.Typeflag == tar.TypeXGlobalHeader {
		return nil, nil
	}

	var p parser
	for {
		n, err := io.ReadFull(sr, blk[:])
		if (err != nil && err != io.EOF) || n == 0 {
			return nil, err
		}
		switch flag := blk.toV7().typeFlag()[0]; flag {
		case tar.TypeXHeader, tar.TypeXGlobalHeader:
			size := p.parseNumeric(blk.toV7().size())
			size += (-size) & (blockSize - 1)
			_, _ = sr.Seek(size, io.SeekCurrent)
			continue
		case tar.TypeGNULongName, tar.TypeGNULongLink:
			size := p.parseNumeric(blk.toV7().size())
			size += (-size) & (blockSize - 1)
			_, _ = sr.Seek(size, io.SeekCurrent)
			continue
		default:
			return handleSparseFile(sr, hdr, blk)
		}
	}
}

func handleSparseFile(sr io.Reader, hdr *tar.Header, rawHdr *block) (sparseHoles, error) {
	var spd sparseDatas
	var err error
	if hdr.Typeflag == tar.TypeGNUSparse {
		spd, err = readOldGNUSparseMap(sr, rawHdr)
	} else {
		spd, err = readGNUSparsePAXHeaders(sr, hdr)
	}

	if err == nil && spd != nil {
		return invertSparseEntries(spd, hdr.Size), nil
	}

	return nil, err
}

func readOldGNUSparseMap(sr io.Reader, blk *block) (sparseDatas, error) {
	var p parser
	s := blk.toGNU().sparse()
	spd := make(sparseDatas, 0, s.maxEntries())
	for {
		for i := 0; i < s.maxEntries(); i++ {
			// This termination condition is identical to GNU and BSD tar.
			if s.entry(i).offset()[0] == 0x00 {
				break // Don't return, need to process extended headers (even if empty)
			}
			offset := p.parseNumeric(s.entry(i).offset())
			length := p.parseNumeric(s.entry(i).length())
			if p.err != nil {
				return nil, p.err
			}
			spd = append(spd, sparseEntry{Offset: offset, Length: length})
		}

		if s.isExtended()[0] > 0 {
			// There are more entries. Read an extension header and parse its entries.
			if _, err := mustReadFull(sr, blk[:]); err != nil {
				return nil, err
			}
			s = blk.toSparse()
			continue
		}
		return spd, nil // Done
	}
}

func readGNUSparsePAXHeaders(sr io.Reader, hdr *tar.Header) (sparseDatas, error) {
	// Identify the version of GNU headers.
	var is1x0 bool
	major, minor := hdr.PAXRecords[paxGNUSparseMajor], hdr.PAXRecords[paxGNUSparseMinor]
	switch {
	case major == "0" && (minor == "0" || minor == "1"):
		is1x0 = false
	case major == "1" && minor == "0":
		is1x0 = true
	case major != "" || minor != "":
		return nil, nil // Unknown GNU sparse PAX version
	case hdr.PAXRecords[paxGNUSparseMap] != "":
		is1x0 = false // 0.0 and 0.1 did not have explicit version records, so guess
	default:
		return nil, nil // Not a PAX format GNU sparse file.
	}

	// Read the sparse map according to the appropriate format.
	if is1x0 {
		return readGNUSparseMap1x0(sr)
	}
	return readGNUSparseMap0x1(hdr.PAXRecords)
}
