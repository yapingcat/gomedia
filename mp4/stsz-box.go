package mp4

import "encoding/binary"

// aligned(8) class SampleSizeBox extends FullBox(‘stsz’, version = 0, 0) {
// 		unsigned int(32) sample_size;
// 		unsigned int(32) sample_count;
// 		if (sample_size==0) {
// 		for (i=1; i <= sample_count; i++) {
// 		unsigned int(32) entry_size;
// 		}
// 	}
// }

type SampleSizeBox struct {
    box  *FullBox
    stsz *movstsz
}

func NewSampleSizeBox() *SampleSizeBox {
    return &SampleSizeBox{
        box: NewFullBox([4]byte{'s', 't', 's', 'z'}, 0),
    }
}

func (stsz *SampleSizeBox) Size() uint64 {
    if stsz.stsz == nil {
        return stsz.box.Size()
    } else if stsz.stsz.sampleSize == 0 {
        return stsz.box.Size() + 8 + 4*uint64(stsz.stsz.sampleCount)
    } else {
        return stsz.box.Size() + 8
    }
}

func (stsz *SampleSizeBox) Encode() (int, []byte) {
    stsz.box.Box.Size = stsz.Size()
    offset, buf := stsz.box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], stsz.stsz.sampleSize)
    offset += 4
    binary.BigEndian.PutUint32(buf[offset:], stsz.stsz.sampleCount)
    offset += 4
    if stsz.stsz.sampleSize == 0 {
        for i := 0; i < int(stsz.stsz.sampleCount); i++ {
            binary.BigEndian.PutUint32(buf[offset:], stsz.stsz.entrySizelist[i])
            offset += 4
        }
    }
    return offset, buf
}

func makeStsz(stsz *movstsz) (boxdata []byte) {
    stszbox := NewSampleSizeBox()
    stszbox.stsz = stsz
    _, boxdata = stszbox.Encode()
    return
}
