package mp4

import (
    "encoding/binary"
)

// aligned(8) abstract class SampleEntry (unsigned int(32) format) extends Box(format){
// 	const unsigned int(8)[6] reserved = 0;
// 	unsigned int(16) data_reference_index;
// 	}

type SampleEntry struct {
    box                  *BasicBox
    data_reference_index uint16
}

func NewSampleEntry(format [4]byte) *SampleEntry {
    return &SampleEntry{
        box:                  NewBasicBox(format),
        data_reference_index: 1,
    }
}

func (entry *SampleEntry) Size() uint64 {
    return 8 + 8
}

func (entry *SampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.box.Decode(buf); err != nil {
        return 0, err
    }
    offset += 6
    entry.data_reference_index = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    return
}

func (entry *SampleEntry) Encode() (int, []byte) {
    offset, buf := entry.box.Encode()
    offset += 6
    binary.BigEndian.PutUint16(buf[offset:], entry.data_reference_index)
    offset += 2
    return offset, buf
}

// class HintSampleEntry() extends SampleEntry (protocol) {
// 		unsigned int(8) data [];
// }

type HintSampleEntry struct {
    Entry *SampleEntry
    Data  byte
}

// class AudioSampleEntry(codingname) extends SampleEntry (codingname){
//  const unsigned int(32)[2] reserved = 0;
// 	template unsigned int(16) channelcount = 2;
// 	template unsigned int(16) samplesize = 16;
// 	unsigned int(16) pre_defined = 0;
// 	const unsigned int(16) reserved = 0 ;
// 	template unsigned int(32) samplerate = { default samplerate of media}<<16;
// }

type AudioSampleEntry struct {
    entry        *SampleEntry
    channelcount uint16
    samplesize   uint16
    samplerate   uint32
}

func NewAudioSampleEntry(format [4]byte) *AudioSampleEntry {
    return &AudioSampleEntry{
        entry: NewSampleEntry(format),
    }
}

func (entry *AudioSampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.entry.Decode(buf); err != nil {
        return 0, err
    }
    offset += 8
    entry.channelcount = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.samplesize = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    offset += 4
    entry.samplerate = binary.BigEndian.Uint32(buf[offset:])
    entry.samplerate = entry.samplerate >> 16
    offset += 4
    return
}

func (entry *AudioSampleEntry) Size() uint64 {
    return entry.entry.Size() + 20
}

func (entry *AudioSampleEntry) Encode() (int, []byte) {
    entry.entry.box.Size = entry.Size()
    offset, buf := entry.entry.Encode()
    offset += 8
    binary.BigEndian.PutUint16(buf[offset:], entry.channelcount)
    offset += 2
    binary.BigEndian.PutUint16(buf[offset:], entry.samplesize)
    offset += 2
    offset += 4
    binary.BigEndian.PutUint32(buf[offset:], entry.samplerate<<16)
    offset += 4
    return offset, buf
}

// class VisualSampleEntry(codingname) extends SampleEntry (codingname){
//  unsigned int(16) pre_defined = 0;
// 	const unsigned int(16) reserved = 0;
// 	unsigned int(32)[3] pre_defined = 0;
// 	unsigned int(16) width;
// 	unsigned int(16) height;
// 	template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
//  template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
//  const unsigned int(32) reserved = 0;
// 	template unsigned int(16) frame_count = 1;
// 	string[32] compressorname;
// 	template unsigned int(16) depth = 0x0018;
// 	int(16) pre_defined = -1;
// 	// other boxes from derived specifications
// 	CleanApertureBox clap; // optional
// 	PixelAspectRatioBox pasp; // optional
// }

type VisualSampleEntry struct {
    entry           *SampleEntry
    width           uint16
    height          uint16
    horizresolution uint32
    vertresolution  uint32
    frame_count     uint16
    compressorname  [32]byte
}

func NewVisualSampleEntry(format [4]byte) *VisualSampleEntry {
    return &VisualSampleEntry{
        entry:           NewSampleEntry(format),
        horizresolution: 0x00480000,
        vertresolution:  0x00480000,
        frame_count:     1,
    }
}

func (entry *VisualSampleEntry) Size() uint64 {
    return entry.entry.Size() + 70
}

func (entry *VisualSampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.entry.Decode(buf); err != nil {
        return 0, err
    }
    offset += 16
    entry.width = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.height = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.horizresolution = binary.BigEndian.Uint32(buf[offset:])
    offset += 4
    entry.vertresolution = binary.BigEndian.Uint32(buf[offset:])
    offset += 8
    entry.frame_count = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    copy(entry.compressorname[:], buf[offset:offset+32])
    offset += 32
    offset += 4
    return
}

func (entry *VisualSampleEntry) Encode() (int, []byte) {
    offset, buf := entry.entry.Encode()
    offset += 16
    binary.BigEndian.PutUint16(buf[offset:], entry.width)
    offset += 2
    binary.BigEndian.PutUint16(buf[offset:], entry.height)
    offset += 2
    binary.BigEndian.PutUint32(buf[offset:], entry.horizresolution)
    offset += 4
    binary.BigEndian.PutUint32(buf[offset:], entry.vertresolution)
    offset += 8
    binary.BigEndian.PutUint16(buf[offset:], entry.frame_count)
    offset += 2
    copy(buf[offset:offset+32], entry.compressorname[:])
    offset += 32
    binary.BigEndian.PutUint16(buf[offset:], 0x0018)
    offset += 2
    binary.BigEndian.PutUint16(buf[offset:], 0xFFFF)
    offset += 2
    return offset, buf
}

// aligned(8) class SampleDescriptionBox (unsigned int(32) handler_type) extends FullBox('stsd', 0, 0){
// 	int i ;
// 	unsigned int(32) entry_count;
// 	   for (i = 1 ; i <= entry_count ; i++){
// 		  switch (handler_type){
// 			 case ‘soun’: // for audio tracks
// 				AudioSampleEntry();
// 				break;
// 			 case ‘vide’: // for video tracks
// 				VisualSampleEntry();
// 				break;
// 			 case ‘hint’: // Hint track
// 				HintSampleEntry();
// 				break;
// 			 case ‘meta’: // Metadata track
// 				MetadataSampleEntry();
// 				break;
// 		}
// 	}
// }

type SampleEntryType uint8

const (
    SAMPLE_AUDIO SampleEntryType = iota
    SAMPLE_VIDEO
)

type SampleDescriptionBox struct {
    box         *FullBox
    entry_count uint32
}

func NewSampleDescriptionBox() *SampleDescriptionBox {
    return &SampleDescriptionBox{
        box: NewFullBox([4]byte{'s', 't', 's', 'd'}, 0),
    }
}

func (stsd *SampleDescriptionBox) Size() uint64 {
    return stsd.box.Size() + 4
}

func (stsd *SampleDescriptionBox) Decode(buf []byte, handler_type [4]byte) (offset int, err error) {

    return
}

func (entry *SampleDescriptionBox) Encode() (int, []byte) {
    offset, buf := entry.box.Encode()
    binary.BigEndian.PutUint32(buf[offset:], entry.entry_count)
    offset += 4
    return offset, buf
}

func makeStsd(track *mp4track, handler_type HandlerType) []byte {

    var avbox []byte
    if track.cid == MOV_CODEC_H264 {
        avbox = makeAvcCBox(track.extra)
    } else if track.cid == MOV_CODEC_H265 {
        avbox = makeAvcCBox(track.extra)
    } else if track.cid == MOV_CODEC_AAC {
        avbox = makeAvcCBox(track.extra)
    }

    var se []byte
    var offset int
    if handler_type.equal(vide) {
        entry := NewVisualSampleEntry(TranscodeCidToCodecName(track.cid))
        entry.width = uint16(track.width)
        entry.height = uint16(track.height)
        entry.entry.box.Size = entry.Size() + uint64(len(avbox))
        offset, se = entry.Encode()
    } else if handler_type.equal(soun) {
        entry := NewAudioSampleEntry(TranscodeCidToCodecName(track.cid))
        entry.channelcount = uint16(track.chanelCount)
        entry.samplerate = track.sampleRate
        entry.samplesize = uint16(track.sampleBits)
        entry.entry.box.Size = entry.Size() + uint64(len(avbox))
        offset, se = entry.Encode()
    }
    copy(se[offset:], avbox)

    stsd := NewSampleDescriptionBox()
    stsd.box.Box.Size = stsd.Size() + uint64(len(se))
    offset2, stsdbox := stsd.Encode()
    copy(stsdbox[offset2:], se)
    return stsdbox
}

func makeAvcCBox(extra extraData) []byte {
    if extra == nil {
        panic("avcc extraData is nil")
    }
    tmp := extra.export()
    AVCC.Size = 8 + uint64(len(tmp))
    offset, boxdata := AVCC.Encode()
    copy(boxdata[offset:], tmp)
    return boxdata
}

func makeHvcCBox(extra extraData) []byte {
    if extra == nil {
        panic("avcc extraData is nil")
    }
    tmp := extra.export()
    HVCC.Size = 8 + uint64(len(tmp))
    offset, boxdata := HVCC.Encode()
    copy(boxdata[offset:], tmp)
    return boxdata
}

func makeEsdsBox(extra extraData) []byte {
    return nil
}
