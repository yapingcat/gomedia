package rtp

import (
	"encoding/binary"

	"github.com/yapingcat/gomedia/codec"
)

// rfc6184 https://datatracker.ietf.org/doc/html/rfc6184
//
// Payload Packet    Single NAL    Non-Interleaved    Interleaved
// Type    Type      Unit Mode           Mode             Mode
// -------------------------------------------------------------
// 0      reserved      ig               ig               ig
// 1-23   NAL unit     yes              yes               no
// 24     STAP-A        no              yes               no
// 25     STAP-B        no               no              yes
// 26     MTAP16        no               no              yes
// 27     MTAP24        no               no              yes
// 28     FU-A          no              yes              yes
// 29     FU-B          no               no              yes
// 30-31  reserved      ig               ig               ig
//

type H264Packer struct {
	ssrc     uint32
	pt       uint8
	sequence uint16
	mtu      int
	cb       RTP_HOOK_FUNC
	stap_a   bool
	sps      []byte
	pps      []byte
}

func NewH264Packer(pt uint8, ssrc uint32, sequence uint16, mtu int) *H264Packer {
	return &H264Packer{
		pt:       pt,
		ssrc:     ssrc,
		sequence: sequence,
		mtu:      mtu,
	}
}

func (pack *H264Packer) HookRtp(cb RTP_HOOK_FUNC) {
	pack.cb = cb
}

func (pack *H264Packer) Pack(frame []byte, timestamp uint32) (pkts [][]byte) {
	codec.SplitFrame(frame, func(nalu []byte) bool {
		nalu_type := codec.H264NaluType(nalu)
		if pack.stap_a {
			switch nalu_type {
			case codec.H264_NAL_SPS:
				return true
			case codec.H264_NAL_PPS:
				return true
			}
			if pack.sps != nil && pack.pps != nil {
				pkts = append(pkts, pack.packStapA([][]byte{pack.sps, pack.pps}, timestamp)...)
				pack.sps = nil
				pack.pps = nil
			}
		}

		if len(frame)+RTP_FIX_HEAD_LEN < pack.mtu {
			pkts = append(pkts, pack.packSingleNalu(nalu, timestamp))
		} else {
			pkts = append(pkts, pack.packFuA(nalu, timestamp)...)
		}
		return true
	})
	return
}

func (pack *H264Packer) packSingleNalu(nalu []byte, timestamp uint32) []byte {
	pkg := RtpPacket{}
	pkg.Header.SSRC = pack.ssrc
	pkg.Header.SequenceNumber = pack.sequence
	pkg.Header.Timestamp = timestamp
	pkg.Header.Marker = 1
	pkg.Payload = nalu
	pack.sequence++
	if pack.cb != nil {
		pack.cb(&pkg)
	}
	return pkg.Encode()
}

//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | FU indicator  |   FU header   |                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               |
// |                                                               |
// |                         FU payload                            |
// |                                                               |
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// FU indicator
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |F|NRI|  Type   |
// +---------------+

// FU header
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |S|E|R|  Type   |
// +---------------+

func (pack *H264Packer) packFuA(nalu []byte, timestamp uint32) (pkts [][]byte) {
	var fuIndicator byte = nalu[0]&0xE0 | 0x1c
	var fuHeader byte = nalu[0]&0x1F | 0x80
	for {
		pkg := RtpPacket{}
		pkg.Header.SSRC = pack.ssrc
		pkg.Header.SequenceNumber = pack.sequence
		pkg.Header.Timestamp = timestamp
		if len(nalu)+RTP_FIX_HEAD_LEN < pack.mtu {
			pkg.Header.Marker = 1
			fuHeader |= 0x40
			pkg.Payload = make([]byte, 0, 2+len(nalu))
			pkg.Payload = append(pkg.Payload, fuIndicator)
			pkg.Payload = append(pkg.Payload, fuHeader)
			pkg.Payload = append(pkg.Payload, nalu[1:]...)
			if pack.cb != nil {
				pack.cb(&pkg)
			}
			pkts = append(pkts, pkg.Encode())
			pack.sequence++
			return
		}
		pkg.Payload = make([]byte, 0, 2+pack.mtu)
		pkg.Payload = append(pkg.Payload, fuIndicator)
		pkg.Payload = append(pkg.Payload, fuHeader)
		if fuHeader&0x80 > 0 {
			pkg.Payload = append(pkg.Payload, nalu[1:pack.mtu-1-RTP_FIX_HEAD_LEN]...)
			fuHeader &= 0x7F
			nalu = nalu[pack.mtu-1-RTP_FIX_HEAD_LEN:]
		} else {
			pkg.Payload = append(pkg.Payload, nalu[:pack.mtu-2-RTP_FIX_HEAD_LEN]...)
			nalu = nalu[pack.mtu-2-RTP_FIX_HEAD_LEN:]
		}
		pkts = append(pkts, pkg.Encode())
		pack.sequence++
	}
}

//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                          RTP Header                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |STAP-A NAL HDR |         NALU 1 Size           | NALU 1 HDR    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         NALU 1 Data                           |
// :                                                               :
// +               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |               | NALU 2 Size                   | NALU 2 HDR    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         NALU 2 Data                           |
// :                                                               :
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

func (pack *H264Packer) packStapA(nalus [][]byte, timestamp uint32) (pkts [][]byte) {
	pkg := RtpPacket{}
	pkg.Header.SSRC = pack.ssrc
	pkg.Header.SequenceNumber = pack.sequence
	pkg.Header.Timestamp = timestamp

	length := 0
	for _, nalu := range nalus {
		length += len(nalu) + 2
	}

	pkg.Payload = make([]byte, 1, length+1)
	pkg.Payload[0] = 24
	for _, nalu := range nalus {
		tmp := make([]byte, 2)
		binary.BigEndian.PutUint16(tmp, uint16(len(nalu)))
		pkg.Payload = append(pkg.Payload, tmp...)
		pkg.Payload = append(pkg.Payload, nalu...)
	}
	pkts = append(pkts, pkg.Encode())
	return
}
