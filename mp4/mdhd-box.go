package mp4

import (
    "encoding/binary"
    "time"

    "github.com/yapingcat/gomedia/mpeg"
)

// aligned(8) class MediaHeaderBox extends FullBox(‘mdhd’, version, 0) { if (version==1) {
// 	unsigned int(64)  creation_time;
// 	unsigned int(64)  modification_time;
// 	unsigned int(32)  timescale;
// 	unsigned int(64)  duration;
//  } else { // version==0
// 	unsigned int(32)  creation_time;
// 	unsigned int(32)  modification_time;
// 	unsigned int(32)  timescale;
// 	unsigned int(32)  duration;
// }
// bit(1) pad = 0;
// unsigned int(5)[3] language; // ISO-639-2/T language code
// unsigned int(16) pre_defined = 0;
// }

func ff_mov_iso639_to_lang(lang [3]byte) (code int) {
    for i := 0; i < 3; i++ {
        c := lang[i]
        c -= 0x60
        if c > 0x1f {
            return -1
        }
        code <<= 5
        code |= int(c)
    }
    return
}

type MediaHeaderBox struct {
    Box               *FullBox
    Creation_time     uint64
    Modification_time uint64
    Timescale         uint32
    Duration          uint64
    Pad               uint8
    Language          [3]uint8
    Pre_defined       uint16
}

func NewMediaHeaderBox() *MediaHeaderBox {
    return &MediaHeaderBox{
        Box:               NewFullBox([4]byte{'m', 'd', 'h', 'd'}, 0),
        Creation_time:     uint64(time.Now().Unix() + 0x7C25B080),
        Modification_time: uint64(time.Now().Unix() + 0x7C25B080),
        Timescale:         1000,
        Language:          [3]byte{'u', 'n', 'd'},
    }
}

func (mdhd *MediaHeaderBox) Size() uint64 {
    if mdhd.Box.Version == 1 {
        return mdhd.Box.Size() + 32
    } else {
        return mdhd.Box.Size() + 20
    }
}

func (mdhd *MediaHeaderBox) Decode(buf []byte) (offset int, err error) {
    if offset, err = mdhd.Box.Decode(buf); err != nil {
        return 0, err
    }
    if mdhd.Box.Version == 1 {
        mdhd.Creation_time = binary.BigEndian.Uint64(buf[offset:])
        offset += 8
        mdhd.Modification_time = binary.BigEndian.Uint64(buf[offset:])
        offset += 8
        mdhd.Timescale = binary.BigEndian.Uint32(buf[offset:])
        offset += 4
        mdhd.Duration = binary.BigEndian.Uint64(buf[offset:])
        offset += 8
    } else {
        mdhd.Creation_time = uint64(binary.BigEndian.Uint32(buf[offset:]))
        offset += 4
        mdhd.Modification_time = uint64(binary.BigEndian.Uint32(buf[offset:]))
        offset += 4
        mdhd.Timescale = binary.BigEndian.Uint32(buf[offset:])
        offset += 4
        mdhd.Duration = uint64(binary.BigEndian.Uint32(buf[offset:]))
        offset += 4
    }
    bs := mpeg.NewBitStream(buf[offset:])
    mdhd.Pad = bs.GetBit()
    mdhd.Language[0] = bs.Uint8(5)
    mdhd.Language[1] = bs.Uint8(5)
    mdhd.Language[2] = bs.Uint8(5)
    mdhd.Pre_defined = 0
    return
}

func (mdhd *MediaHeaderBox) Encode() (int, []byte) {
    mdhd.Box.Box.Size = mdhd.Size()
    offset, buf := mdhd.Box.Encode()
    if mdhd.Box.Version == 1 {
        binary.BigEndian.PutUint64(buf[offset:], mdhd.Creation_time)
        offset += 8
        binary.BigEndian.PutUint64(buf[offset:], mdhd.Modification_time)
        offset += 8
        binary.BigEndian.PutUint32(buf[offset:], mdhd.Timescale)
        offset += 4
        binary.BigEndian.PutUint64(buf[offset:], mdhd.Duration)
        offset += 8
    } else {
        binary.BigEndian.PutUint32(buf[offset:], uint32(mdhd.Creation_time))
        offset += 4
        binary.BigEndian.PutUint32(buf[offset:], uint32(mdhd.Modification_time))
        offset += 4
        binary.BigEndian.PutUint32(buf[offset:], mdhd.Timescale)
        offset += 4
        binary.BigEndian.PutUint32(buf[offset:], uint32(mdhd.Duration))
        offset += 4
    }
    binary.BigEndian.PutUint16(buf[offset:], uint16(ff_mov_iso639_to_lang(mdhd.Language)&0x7FFF))
    offset += 2
    offset += 2
    return offset, buf
}

func makeMdhdBox() []byte {
    mdhd := NewMediaHeaderBox()
    _, boxdata := mdhd.Encode()
    return boxdata
}
