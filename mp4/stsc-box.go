package mp4

import "encoding/binary"

// aligned(8) class SampleToChunkBox extends FullBox(‘stsc’, version = 0, 0) {
//     unsigned int(32) entry_count;
//     for (i=1; i <= entry_count; i++) {
//         unsigned int(32) first_chunk;
//         unsigned int(32) samples_per_chunk;
//         unsigned int(32) sample_description_index;
//     }
// }

type SampleToChunkBox struct {
    box        *FullBox
    stscentrys *movstsc
}

func NewSampleToChunkBox() *SampleToChunkBox {
    return &SampleToChunkBox{
        box: NewFullBox([4]byte{'s', 't', 's', 'c'}, 0),
    }
}

func (stsc *SampleToChunkBox) Size() uint64 {
    if stsc.stscentrys == nil {
        return stsc.box.Size()
    } else {
        return stsc.box.Size() + 4 + 12*uint64(stsc.stscentrys.entryCount)
    }
}

func (stsc *SampleToChunkBox) Encode() (int, []byte) {
    stsc.box.Box.Size = stsc.Size()
    offset, buf := stsc.Encode()
    binary.BigEndian.PutUint32(buf[offset:], stsc.stscentrys.entryCount)
    offset += 4
    for i := 0; i < int(stsc.stscentrys.entryCount); i++ {
        binary.BigEndian.PutUint32(buf[offset:], stsc.stscentrys.entrys[i].firstChunk)
        offset += 4
        binary.BigEndian.PutUint32(buf[offset:], stsc.stscentrys.entrys[i].samplesPerChunk)
        offset += 4
        binary.BigEndian.PutUint32(buf[offset:], stsc.stscentrys.entrys[i].sampleDescriptionIndex)
        offset += 4
    }
    return offset, buf
}

func makeStsc(stsc *movstsc) (boxdata []byte) {
    stscbox := NewSampleToChunkBox()
    stscbox.stscentrys = stsc
    _, boxdata = stscbox.Encode()
    return
}
