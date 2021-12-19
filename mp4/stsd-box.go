package mp4

import (
    "encoding/binary"
)

// aligned(8) abstract class SampleEntry (unsigned int(32) format) extends Box(format){
// 	const unsigned int(8)[6] reserved = 0;
// 	unsigned int(16) data_reference_index;
// 	}

type SampleEntry struct {
    Box                  *BasicBox
    Data_reference_index uint16
}

func NewSampleEntry(format [4]byte) *SampleEntry {
    return &SampleEntry{
        Box:                  NewBasicBox(format),
        Data_reference_index: 1,
    }
}

func (entry *SampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.Box.Decode(buf); err != nil {
        return 0, err
    }
    offset += 6
    entry.Data_reference_index = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    return
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
    Entry        *SampleEntry
    Channelcount uint16
    Samplesize   uint16
    Samplerate   uint32
}

func (entry *AudioSampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.Entry.Decode(buf); err != nil {
        return 0, err
    }
    offset += 8
    entry.Channelcount = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.Samplesize = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    offset += 4
    entry.Samplerate = binary.BigEndian.Uint32(buf[offset:])
    offset += 4
    return
}

// class VisualSampleEntry(codingname) extends SampleEntry (codingname){ unsigned int(16) pre_defined = 0;
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
    Entry           *SampleEntry
    Width           uint16
    Height          uint16
    Horizresolution uint32
    Vertresolution  uint32
    Frame_count     uint16
    Compressorname  [32]byte
}

func NewVisualSampleEntry(format [4]byte) *VisualSampleEntry {
    return &VisualSampleEntry{
        Entry: NewSampleEntry(format),
    }
}

func (entry *VisualSampleEntry) Decode(buf []byte) (offset int, err error) {
    if offset, err = entry.Entry.Decode(buf); err != nil {
        return 0, err
    }
    offset += 14
    entry.Width = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.Height = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    entry.Horizresolution = binary.BigEndian.Uint32(buf[offset:])
    offset += 4
    entry.Vertresolution = binary.BigEndian.Uint32(buf[offset:])
    offset += 8
    entry.Frame_count = binary.BigEndian.Uint16(buf[offset:])
    offset += 2
    copy(entry.Compressorname[:], buf[offset:offset+32])
    offset += 32
    offset += 4
    return
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
    entryType   uint8
}

func NewSampleDescriptionBox() *SampleDescriptionBox {
    return &SampleDescriptionBox{
        box: NewFullBox([4]byte{'s', 't', 's', 'd'}, 0),
    }
}

func (stsd *SampleDescriptionBox) Decode(buf []byte, handler_type [4]byte) (offset int, err error) {

    return
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

}
