package ogg

import (
    "bytes"
    "encoding/binary"

    "github.com/yapingcat/gomedia/codec"
)

type oggCodec interface {
    codecid() codec.CodecID
    magic() []byte
    magicSize() int
}

type OpusCodec struct {
}

func (opus OpusCodec) codecid() codec.CodecID {
    return codec.CODECID_AUDIO_OPUS
}

func (opus OpusCodec) magic() []byte {
    return []byte("OpusHead")
}

func (opus OpusCodec) magicSize() int {
    return 8
}

type VP8Codec struct {
}

func (vp8 VP8Codec) codecid() codec.CodecID {
    return codec.CODECID_VIDEO_VP8
}

func (vp8 VP8Codec) magic() []byte {
    return []byte("OVP80")
}

func (vp8 VP8Codec) magicSize() int {
    return 5
}

var codecs []oggCodec

func init() {
    codecs = make([]oggCodec, 2)
    codecs[0] = OpusCodec{}
    codecs[1] = VP8Codec{}
}

type oggParser interface {
    header(stream *oggStream, packet []byte)
    packet(stream *oggStream, packet []byte) (frame []byte, pts uint64, dts uint64)
    gptopts(granulePos uint64) uint64
}

func createParser(cid codec.CodecID) oggParser {
    switch cid {
    case codec.CODECID_AUDIO_OPUS:
        return &opusDemuxer{
            lastpts: ^uint64(0),
        }
    case codec.CODECID_VIDEO_VP8:
        return &vp8Demuxer{
            lastpts: ^uint64(0),
            pktIdx:  0,
        }
    default:
        panic("unsupport codecid")
    }
}

type opusDemuxer struct {
    ctx     codec.OpusContext
    lastpts uint64
    granule uint64
}

// opus ID head
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |      'O'      |      'p'      |      'u'      |      's'      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |      'H'      |      'e'      |      'a'      |      'd'      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  Version = 1  | Channel Count |           Pre-skip            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                     Input Sample Rate (Hz)                    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |   Output Gain (Q7.8 in dB)    | Mapping Family|               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+               :
// |                                                               |
// :               Optional Channel Mapping Table...               :
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

func (opus *opusDemuxer) header(stream *oggStream, packet []byte) {
    if bytes.Equal([]byte("OpusHead"), packet[0:8]) {
        opus.ctx.ParseExtranData(packet)
    } else if bytes.Equal([]byte("OpusHead"), packet[0:8]) {
        return
    }

}

func (opus *opusDemuxer) packet(stream *oggStream, packet []byte) (frame []byte, pts uint64, dts uint64) {
    if bytes.Equal([]byte("OpusTags"), packet[0:4]) {
        return
    }

    if stream.lost == 1 {
        return packet, opus.lastpts, opus.lastpts
    }

    if opus.lastpts == ^uint64(0) {
        opus.lastpts = 0
    }
    frame = packet
    pts = opus.lastpts
    dts = pts

    if opus.granule != stream.currentPage.granulePos && !stream.currentPage.eos {
        opus.lastpts = 0
        opus.granule = stream.currentPage.granulePos
    }

    if opus.lastpts == 0 {
        var duration uint64
        for _, seg := range stream.currentPage.packets {
            duration += codec.OpusPacketDuration(seg)
        }
        opus.lastpts = opus.granule - duration - uint64(opus.ctx.Preskip)
    }

    duration := codec.OpusPacketDuration(packet)
    opus.lastpts = opus.lastpts + duration

    return
}
func (opus *opusDemuxer) gptopts(granulePos uint64) uint64 {
    return 0
}

//ffmpeg oggparsevp8.c
type vp8Demuxer struct {
    pktIdx            uint64
    lastpts           uint64
    granule           uint64
    width             uint16
    height            uint16
    sampleAspectratio uint32
    frameRate         uint32
}

func (vp8 *vp8Demuxer) header(stream *oggStream, packet []byte) {
    if !bytes.Equal([]byte("OVP80"), packet[0:5]) {
        return
    }

    switch packet[5] {
    case 0x01:
        if packet[6] != 1 {
            return
        }
        vp8.width = binary.BigEndian.Uint16(packet[8:])
        vp8.height = binary.BigEndian.Uint16(packet[10:])
        num := uint32(packet[12])
        num = (num << 8) | uint32(packet[13])
        num = (num << 8) | uint32(packet[14])
        den := uint32(packet[15])
        den = (den << 8) | uint32(packet[16])
        den = (den << 8) | uint32(packet[17])
        vp8.sampleAspectratio = num / den
        num = binary.BigEndian.Uint32(packet[18:])
        den = binary.BigEndian.Uint32(packet[22:])
        vp8.frameRate = num / den
    case 0x02:
        if packet[6] != 0x20 {
            return
        }
        //TODO Parse Comment
    default:
        return
    }
}

func (vp8 *vp8Demuxer) packet(stream *oggStream, packet []byte) (frame []byte, pts uint64, dts uint64) {

    if stream.lost == 1 {
        return packet, vp8.lastpts, vp8.lastpts
    }

    if vp8.granule != stream.currentPage.granulePos {
        vp8.lastpts = 0
        vp8.pktIdx = 0
        vp8.granule = stream.currentPage.granulePos
    }
    var duration uint64 = 0
    for i := int(vp8.pktIdx); i < len(stream.currentPage.packets); i++ {
        duration += uint64((stream.currentPage.packets[i][0] >> 4) & 1)
    }
    vp8.lastpts = vp8.gptopts(stream.currentPage.granulePos) - duration
    frame = packet
    pts = vp8.lastpts
    dts = pts
    vp8.pktIdx++
    return
}

func (vp8 *vp8Demuxer) gptopts(granulePos uint64) uint64 {
    var invcnt uint64 = 0
    if ((granulePos >> 30) & 3) == 0 {
        invcnt = 1
    }
    pts := (granulePos >> 32) - invcnt
    return pts
}
