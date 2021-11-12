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
	cache     []byte
	OnPacket  func(pkg PSPacket)
	OnFrame   func(frame []byte, cid PS_STREAM_TYPE, pts uint64, dts uint64)
}

func NewPSDemuxer() *PSDemuxer {
	return &PSDemuxer{
		streamMap: make(map[uint8]*psstream),
		pkg:       new(PSPacket),
		cache:     make([]byte, 0, 256),
		OnPacket:  nil,
		OnFrame:   nil,
	}
}

func (psdemuxer *PSDemuxer) Input(data []byte) error {
	var bs *mpeg.BitStream
	if len(psdemuxer.cache) > 0 {
		psdemuxer.cache = append(psdemuxer.cache, data...)
		bs = mpeg.NewBitStream(psdemuxer.cache)
	} else {
		bs = mpeg.NewBitStream(data)
	}

	saveReseved := func() {
		tmpcache := make([]byte, bs.RemainBytes())
		copy(tmpcache, bs.RemainData())
		psdemuxer.cache = tmpcache
	}

	var ret error = nil
	for !bs.EOS() {
		if mpegerr, ok := ret.(Error); ok {
			if mpegerr.NeedMore() {
				saveReseved()
			}
			break
		}
		if bs.RemainBits() < 32 {
			ret = errNeedMore
			saveReseved()
			break
		}
		prefix_code := bs.NextBits(32)
		switch prefix_code {
		case 0x000001BA: //pack header
			if psdemuxer.pkg.Header == nil {
				psdemuxer.pkg.Header = new(PSPackHeader)
			}
			ret = psdemuxer.pkg.Header.Decode(bs)
		case 0x000001BB: //system header
			if psdemuxer.pkg.Header == nil {
				panic("psdemuxer.pkg.Header must not be nil")
			}
			if psdemuxer.pkg.Header.Sys_Header == nil {
				psdemuxer.pkg.Header.Sys_Header = new(System_header)
			}
			ret = psdemuxer.pkg.Header.Sys_Header.Decode(bs)
		case 0x000001BC: //program stream map
			if psdemuxer.pkg.Psm == nil {
				psdemuxer.pkg.Psm = new(Program_stream_map)
			}
			if ret = psdemuxer.pkg.Psm.Decode(bs); ret == nil {
				for _, streaminfo := range psdemuxer.pkg.Psm.Stream_map {
					if _, found := psdemuxer.streamMap[streaminfo.Elementary_stream_id]; !found {
						stream := newpsstream(streaminfo.Elementary_stream_id, PS_STREAM_TYPE(streaminfo.Stream_type))
						psdemuxer.streamMap[stream.sid] = stream
					}
				}
			}
		case 0x000001BD, 0x000001BE, 0x000001BF, 0x000001F0, 0x000001F1,
			0x000001F2, 0x000001F3, 0x000001F4, 0x000001F5, 0x000001F6,
			0x000001F7, 0x000001F8, 0x000001F9, 0x000001FA, 0x000001FB:
			if psdemuxer.pkg.CommPes == nil {
				psdemuxer.pkg.CommPes = new(CommonPesPacket)
			}
			ret = psdemuxer.pkg.CommPes.Decode(bs)
		case 0x000001FF: //program stream directory
			if psdemuxer.pkg.Psd == nil {
				psdemuxer.pkg.Psd = new(Program_stream_directory)
			}
			ret = psdemuxer.pkg.Psd.Decode(bs)
		case 0x000001B9: //MPEG_program_end_code
			continue
		default:
			if prefix_code&0xFFFFFFE0 == 0x000001C0 || prefix_code&0xFFFFFFE0 == 0x000001E0 {
				if psdemuxer.pkg.Pes == nil {
					psdemuxer.pkg.Pes = NewPesPacket()
				}
				if ret = psdemuxer.pkg.Pes.Decode(bs); ret == nil {
					if stream, found := psdemuxer.streamMap[psdemuxer.pkg.Pes.Stream_id]; found {
						psdemuxer.demuxPespacket(stream, psdemuxer.pkg.Pes)
					}
				}
			} else {
				bs.SkipBits(8)
			}
		}
	}

	if ret == nil && len(psdemuxer.cache) > 0 {
		psdemuxer.cache = nil
	}

	return ret
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
