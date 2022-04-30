package ogg

import (
    "bytes"

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
    gptopts()
}

func createParser(cid codec.CodecID) oggParser {
    switch cid {
    case codec.CODECID_AUDIO_OPUS:
        return &opusDemuxer{
            lastpts: ^uint64(0),
        }
    case codec.CODECID_VIDEO_VP8:
        //TODO
    default:
        panic("unsupport codecid")
    }
    return nil
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
func (opus *opusDemuxer) gptopts() {

}
