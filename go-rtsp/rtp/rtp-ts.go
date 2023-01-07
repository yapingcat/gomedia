package rtp

import "bytes"

type TsPacker struct {
	CommPacker
	pt       uint8
	ssrc     uint32
	sequence uint16
}

func NewTsPacker(pt uint8, ssrc uint32, sequence uint16, mtu int) *TsPacker {
	return &TsPacker{
		pt:         pt,
		ssrc:       ssrc,
		sequence:   sequence,
		CommPacker: CommPacker{mtu: mtu},
	}
}

func (pack *TsPacker) Pack(data []byte, timestamp uint32) error {
	for len(data) > 0 {
		pkg := RtpPacket{}
		pkg.Header.PayloadType = pack.pt
		pkg.Header.SequenceNumber = pack.sequence
		pkg.Header.SSRC = pack.ssrc
		pkg.Header.Timestamp = timestamp
		pkg.Header.Marker = 0
		if len(data) > pack.mtu-RTP_FIX_HEAD_LEN {
			pkg.Payload = make([]byte, pack.mtu-RTP_FIX_HEAD_LEN)
			copy(pkg.Payload, data[:pack.mtu-RTP_FIX_HEAD_LEN])
			data = data[pack.mtu-RTP_FIX_HEAD_LEN:]
		} else {
			pkg.Payload = make([]byte, len(data))
			copy(pkg.Payload, data)
			data = data[:0]
		}
		if pack.onRtp != nil {
			pack.onRtp(&pkg)
		}
		if pack.onPacket != nil {
			return pack.onPacket(pkg.Encode())
		}
		pack.sequence++
	}
	return nil
}

type TsUnPacker struct {
	CommUnPacker
	timestamp    int64
	lastSequence uint16
	lost         bool
	frameBuffer  *bytes.Buffer
}

func NewTsUnPacker() *TsUnPacker {
	return &TsUnPacker{
		frameBuffer: new(bytes.Buffer),
		timestamp:   -1,
	}
}

func (unpacker *TsUnPacker) UnPack(pkt []byte) error {
	pkg := &RtpPacket{}
	if err := pkg.Decode(pkt); err != nil {
		return err
	}

	//first rtp packet
	if unpacker.timestamp == -1 {
		unpacker.timestamp = int64(pkg.Header.Timestamp)
		unpacker.lastSequence = pkg.Header.SequenceNumber
		unpacker.lost = false
		unpacker.frameBuffer.Write(pkg.Payload)
		return nil
	}

	if unpacker.lastSequence+1 != pkg.Header.SequenceNumber {
		unpacker.lost = true
	}

	//时间戳作为帧分割标志
	if unpacker.timestamp != int64(pkg.Header.Timestamp) {
		if unpacker.onFrame != nil {
			unpacker.onFrame(unpacker.frameBuffer.Bytes(), uint32(unpacker.timestamp), unpacker.lost)
		}

		if unpacker.lastSequence+1 == pkg.Header.SequenceNumber {
			unpacker.lost = false
		}
		unpacker.frameBuffer.Reset()
	}
	unpacker.timestamp = int64(pkg.Header.Timestamp)
	unpacker.lastSequence = pkg.Header.SequenceNumber
	unpacker.frameBuffer.Write(pkg.Payload)
	return nil
}
