package mpeg2

import (
	"github.com/yapingcat/gomedia/mpeg"
)

type psstream struct {
	sid       uint8
	cid       PS_STREAM_TYPE
	pts       uint64
	dts       uint64
	streamBuf []byte
}

func newpsstream(sid uint8, cid PS_STREAM_TYPE) *psstream {
	return &psstream{
		sid:       sid,
		cid:       cid,
		streamBuf: make([]byte, 0, 4096),
	}
}

type PSDemuxer struct {
	streamMap map[uint8]*psstream
	pkg       *PSPacket
	OnPacket  func(pkg PSPacket)
	OnFrame   func(frame []byte, cid PS_STREAM_TYPE, pts uint64, dts uint64)
}

func NewPSDemuxer() *PSDemuxer {
	return &PSDemuxer{
		streamMap: make(map[uint8]*psstream),
		pkg:       new(PSPacket),
		OnPacket:  nil,
		OnFrame:   nil,
	}
}

func (psdemuxer *PSDemuxer) Input(data []byte) error {
	bs := mpeg.NewBitStream(data)
	for !bs.EOS() {
		prefix_code := bs.NextBits(32)
		switch prefix_code {
		case 0x000001BA:
			if psdemuxer.pkg.Header == nil {
				psdemuxer.pkg.Header = new(PSPackHeader)
			}
			psdemuxer.pkg.Header.Decode(bs)
		case 0x000001BC:
			if psdemuxer.pkg.Psm == nil {
				psdemuxer.pkg.Psm = new(Program_stream_map)
			}
			psdemuxer.pkg.Psm.Decode(bs)
			for _, streaminfo := range psdemuxer.pkg.Psm.Stream_map {
				if _, found := psdemuxer.streamMap[streaminfo.Elementary_stream_id]; !found {
					stream := newpsstream(streaminfo.Elementary_stream_id, PS_STREAM_TYPE(streaminfo.Stream_type))
					psdemuxer.streamMap[stream.sid] = stream
				}
			}
		case 0x000001FF:
			//TODO Program Stream directory
			bs.SkipBits(32)
			length := bs.Uint16(16)
			bs.SkipBits(int(length))
		case 0x000001B9:
			continue
		default:
			if prefix_code&0xE0 == 0xC0 || prefix_code&0xE0 == 0xE0 {
				if psdemuxer.pkg.Pes == nil {
					psdemuxer.pkg.Pes = NewPesPacket()
				}
				psdemuxer.pkg.Pes.Decode(bs)
				if stream, found := psdemuxer.streamMap[psdemuxer.pkg.Pes.Stream_id]; found {
					psdemuxer.demuxPespacket(stream, psdemuxer.pkg.Pes)
				}
			} else {
					panic("unsupport")
			}
		}

	}

	return nil
}

func (psdemuxer *PSDemuxer) demuxPespacket(stream *psstream, pes *PesPacket) error {
	switch stream.cid {
	case PS_STREAM_AAC, PS_STREAM_G711A, PS_STREAM_G711U:
		return psdemuxer.demuxAudio(stream, pes)
	case PS_STREAM_H264, PS_STREAM_H265:
		return psdemuxer.demuxH26x(stream, pes)
	}
	return nil
}

func (psdemuxer *PSDemuxer) demuxAudio(stream *psstream, pes *PesPacket) error {
	if stream.pts != pes.Pts && len(stream.streamBuf) > 0 {
		if psdemuxer.OnFrame != nil {
			psdemuxer.OnFrame(stream.streamBuf, stream.cid, stream.pts, stream.dts)
		}
		stream.streamBuf = stream.streamBuf[:0]
	}
	stream.streamBuf = append(stream.streamBuf, pes.Pes_payload...)
	stream.pts = pes.Pts
	stream.dts = pes.Dts
	return nil
}

func (psdemuxer *PSDemuxer) demuxH26x(stream *psstream, pes *PesPacket) error {
	if len(stream.streamBuf) == 0 {
		stream.pts = pes.Pts
		stream.dts = pes.Dts
	}
	stream.streamBuf = append(stream.streamBuf, pes.Pes_payload...)
	framebeg := 0
	start, sc := mpeg.FindStarCode(stream.streamBuf, 0)
	framebeg = start
	for start >= 0 {
		end, sc2 := mpeg.FindStarCode(stream.streamBuf, start+int(sc))
		if end < 0 {
			break
		}
		if stream.cid == PS_STREAM_H264 {
			naluType := mpeg.H264NaluType(stream.streamBuf[start:])
			if naluType == mpeg.H264_NAL_AUD {
				framebeg = end
			} else if mpeg.IsH264VCLNaluType(naluType) {
				if psdemuxer.OnFrame != nil {
					psdemuxer.OnFrame(stream.streamBuf[framebeg:end], stream.cid, stream.pts, stream.dts)
					framebeg = end
				}
			}
		} else if stream.cid == PS_STREAM_H265 {
			naluType := mpeg.H265NaluType(stream.streamBuf[start:])
			if naluType == mpeg.H265_NAL_AUD {
				framebeg = end
			} else if mpeg.IsH265VCLNaluType(naluType) {
				if psdemuxer.OnFrame != nil {
					psdemuxer.OnFrame(stream.streamBuf[framebeg:end], stream.cid, stream.pts, stream.dts)
					framebeg = end
				}
			}
		}
		start = end
		sc = sc2
	}
	stream.streamBuf = stream.streamBuf[framebeg:]
	stream.pts = pes.Pts
	stream.dts = pes.Dts
	return nil
}
