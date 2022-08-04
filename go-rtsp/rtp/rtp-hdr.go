package rtp

import (
	"encoding/binary"
	"errors"
)

const RTP_FIX_HEAD_LEN = 12

type RtpHdr struct {
	Version        uint8
	PaddingFlag    uint8
	ExtensionFlag  uint8
	CC             uint8
	Marker         uint8
	PayloadType    uint8
	SequenceNumber uint16
	Timestamp      uint32
	SSRC           uint32
	CSRC           []uint32
}

//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P|X|  CC   |M|     PT      |       sequence number         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |           synchronization source (SSRC) identifier            |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// |            contributing source (CSRC) identifiers             |
// |                             ....                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

func (head *RtpHdr) Encode() []byte {
	data := make([]byte, RTP_FIX_HEAD_LEN)
	data[0] = byte(0x80) | (head.PaddingFlag & 0x01 << 5) | (head.ExtensionFlag & 0x01 << 4) | (head.CC & 0x0F)
	data[1] = head.Marker&0x01<<7 | head.PayloadType&0x7F
	binary.BigEndian.PutUint16(data[2:], head.SequenceNumber)
	binary.BigEndian.PutUint32(data[4:], head.Timestamp)
	binary.BigEndian.PutUint32(data[8:], head.SSRC)

	for _, csrc := range head.CSRC {
		tmp := make([]byte, 4)
		binary.BigEndian.PutUint32(tmp, csrc)
		data = append(data, tmp...)
	}
	return data
}

func (head *RtpHdr) Decode(pkt []byte) (int, error) {
	if len(pkt) < RTP_FIX_HEAD_LEN {
		return 0, errors.New("length of rtp must >= 12")
	}
	head.Version = pkt[0] >> 6
	head.PaddingFlag = pkt[0] >> 5 & 0x01
	head.ExtensionFlag = pkt[0] >> 4 & 0x01
	head.CC = pkt[0] & 0x0F
	head.Marker = pkt[1] >> 7
	head.PayloadType = pkt[1] & 0x7F
	head.SequenceNumber = binary.BigEndian.Uint16(pkt[2:])
	head.Timestamp = binary.BigEndian.Uint32(pkt[4:])
	head.SSRC = binary.BigEndian.Uint32(pkt[8:])
	if len(pkt)-RTP_FIX_HEAD_LEN < 4*int(head.CC) {
		return 0, errors.New("need more space for csrc")
	}
	head.CSRC = make([]uint32, head.CC)
	for i := 0; i < int(head.CC); i++ {
		head.CSRC[i] = binary.BigEndian.Uint32(pkt[12+4*i:])
	}
	return RTP_FIX_HEAD_LEN + 4*len(head.CSRC), nil
}
