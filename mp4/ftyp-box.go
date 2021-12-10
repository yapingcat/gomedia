package mp4

import (
	"encoding/binary"
)

type FileTypeBox struct {
	Box               *BasicBox
	Major_brand       uint32
	Minor_version     uint32
	Compatible_brands []uint32
}

func NewFileTypeBox() *FileTypeBox {
	return &FileTypeBox{
		Box: NewBasicBox([4]byte{'f', 't', 'y', 'p'}),
	}
}

func (ftyp *FileTypeBox) Size() uint64 {
	return uint64(8 + len(ftyp.Compatible_brands)*4 + 8)
}

func (ftyp *FileTypeBox) Decode(buf []byte) (int, error) {
	if offset, err := ftyp.Box.Decode(buf); err != nil {
		return 0, err
	} else {
		_ = buf[ftyp.Box.Size]
		ftyp.Major_brand = binary.BigEndian.Uint32(buf[offset:])
		ftyp.Minor_version = binary.BigEndian.Uint32(buf[offset+4:])
		offset += 8
		for ; offset < int(ftyp.Box.Size); offset += 4 {
			ftyp.Compatible_brands = append(ftyp.Compatible_brands, binary.BigEndian.Uint32(buf[offset:]))
		}
		return offset, nil
	}
}

func (ftyp *FileTypeBox) Encode() (int, []byte) {
	ftyp.Box.Size = ftyp.Size()
	offset, buf := ftyp.Box.Encode()
	binary.BigEndian.PutUint32(buf[offset:], ftyp.Major_brand)
	binary.BigEndian.PutUint32(buf[offset:], ftyp.Minor_version)
	offset += 8
	for i := 0; offset < int(ftyp.Box.Size); offset += 4 {
		binary.BigEndian.PutUint32(buf[offset:], ftyp.Compatible_brands[i])
		i++
	}
	return offset, buf
}
