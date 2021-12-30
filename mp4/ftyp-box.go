package mp4

import (
	"encoding/binary"
)

var isom [4]byte = [4]byte{'i', 's', 'o', 'm'}
var iso2 [4]byte = [4]byte{'i', 's', 'o', '2'}
var avc1 [4]byte = [4]byte{'a', 'v', 'c', '1'}
var mp41 [4]byte = [4]byte{'m', 'p', '4', '1'}
var dash [4]byte = [4]byte{'d', 'a', 's', 'h'}

func mov_tag(tag [4]byte) uint32 {
	return binary.LittleEndian.Uint32(tag[:])
}

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

func (ftyp *FileTypeBox) decode(rh Reader, size uint32) (int, error) {
	buf := make([]byte, size-BasicBoxLen)
	if n, err := rh.ReadAtLeast(buf); err != nil {
		return n, err
	}
	ftyp.Major_brand = binary.LittleEndian.Uint32(buf[0:])
	ftyp.Minor_version = binary.BigEndian.Uint32(buf[4:])
	n := 8
	for ; BasicBoxLen+n < int(size); n += 4 {
		ftyp.Compatible_brands = append(ftyp.Compatible_brands, binary.LittleEndian.Uint32(buf[n:]))
	}
	return n, nil
}

func (ftyp *FileTypeBox) Encode() (int, []byte) {
	ftyp.Box.Size = ftyp.Size()
	offset, buf := ftyp.Box.Encode()
	binary.LittleEndian.PutUint32(buf[offset:], ftyp.Major_brand)
	offset += 4
	binary.BigEndian.PutUint32(buf[offset:], ftyp.Minor_version)
	offset += 4
	for i := 0; offset < int(ftyp.Box.Size); offset += 4 {
		binary.LittleEndian.PutUint32(buf[offset:], ftyp.Compatible_brands[i])
		i++
	}
	return offset, buf
}

func decodeFtypBox(demuxer *MovDemuxer, size uint32) (err error) {
	ftyp := FileTypeBox{}
	if _, err = ftyp.decode(demuxer.readerHandler, size); err != nil {
		return
	}
	demuxer.mp4Info.CompatibleBrands = ftyp.Compatible_brands
	demuxer.mp4Info.MajorBrand = ftyp.Major_brand
	demuxer.mp4Info.MinorVersion = ftyp.Minor_version
	return
}
