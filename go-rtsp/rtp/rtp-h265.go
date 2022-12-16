package rtp

import (
    "bytes"
    "errors"

    "github.com/yapingcat/gomedia/go-codec"
)

//h265 nalu head
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |F| nalu type |  layer id | tid |
// +---------------+-+-+-+-+-+-+-+-+

//rtp h265
//rfc7798
//fu
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |    PayloadHdr (Type=49)       |   FU header   | DONL (cond)   |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-|
// | DONL (cond)   |                                               |
// |-+-+-+-+-+-+-+-+                                               |
// |                         FU payload                            |
// |                                                               |
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// Fu header
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |S|E|  FuType   |
// +---------------+

type H265Packer struct {
    pt       uint8
    ssrc     uint32
    sequence uint16
    mtu      int
    onpkt    ON_RTP_PKT_FUNC
    onRtp    RTP_HOOK_FUNC
}

func NewH265Packer(pt uint8, ssrc uint32, sequence uint16, mtu int) *H265Packer {
    return &H265Packer{
        pt:       pt,
        ssrc:     ssrc,
        sequence: sequence,
        mtu:      mtu,
    }
}

func (h265 *H265Packer) HookRtp(cb RTP_HOOK_FUNC) {
    h265.onRtp = cb
}

func (h265 *H265Packer) SetMtu(mtu int) {
    h265.mtu = mtu
}

func (h265 *H265Packer) OnPacket(onPkt ON_RTP_PKT_FUNC) {
    h265.onpkt = onPkt
}

func (h265 *H265Packer) Pack(data []byte, timestamp uint32) error {
    codec.SplitFrame(data, func(nalu []byte) bool {
        if len(nalu)+RTP_FIX_HEAD_LEN < h265.mtu {
            h265.packSingleNalu(nalu, timestamp)
        } else {
            h265.packFu(nalu, timestamp)
        }
        return true
    })
    return nil
}

func (h265 *H265Packer) packSingleNalu(nalu []byte, timestamp uint32) error {
    pkg := RtpPacket{}
    pkg.Header.PayloadType = h265.pt
    pkg.Header.SequenceNumber = h265.sequence
    pkg.Header.SSRC = h265.ssrc
    pkg.Header.Timestamp = timestamp
    pkg.Header.Marker = 1
    pkg.Payload = make([]byte, len(nalu))
    copy(pkg.Payload, nalu)
    if h265.onRtp != nil {
        h265.onRtp(&pkg)
    }
    if h265.onpkt != nil {
        h265.onpkt(pkg.Encode())
    }
    return nil
}

func (h265 *H265Packer) packFu(nalu []byte, timestamp uint32) error {
    var payloadHdr [2]byte
    var fuHeader byte
    payloadHdr[0] = (nalu[0] & 0x81) | (0x31 << 1)
    payloadHdr[1] = nalu[1]
    fuHeader = ((nalu[0] >> 1) & 0x3f)
    start := true
    end := false
    nalu = nalu[2:]
    for len(nalu) > 0 {
        pkg := RtpPacket{}
        pkg.Header.PayloadType = h265.pt
        pkg.Header.SequenceNumber = h265.sequence
        pkg.Header.SSRC = h265.ssrc
        pkg.Header.Timestamp = timestamp
        length := 0
        if len(nalu)+RTP_FIX_HEAD_LEN+3 < h265.mtu {
            end = true
            length = len(nalu)
            pkg.Header.Marker = 1
        } else {
            length = h265.mtu - RTP_FIX_HEAD_LEN - 3
        }
        pkg.Payload = make([]byte, length+3)
        pkg.Payload[0] = payloadHdr[0]
        pkg.Payload[1] = payloadHdr[1]
        if start {
            pkg.Payload[2] = fuHeader | 0x80
        }
        if end {
            pkg.Payload[2] = fuHeader | 0x40
        }
        copy(pkg.Payload[3:], nalu[:length])
        if h265.onRtp != nil {
            h265.onRtp(&pkg)
        }
        if h265.onpkt != nil {
            h265.onpkt(pkg.Encode())
        }
        nalu = nalu[length:]
    }
    return nil
}

type H265UnPacker struct {
    onFrame      ON_FRAME_FUNC
    timestamp    uint32
    lastSequence uint16
    lost         bool
    frameBuffer  *bytes.Buffer
}

func NewH265UnPacker() *H265UnPacker {
    unpacker := &H265UnPacker{
        frameBuffer: new(bytes.Buffer),
    }
    unpacker.frameBuffer.Grow(1500)
    unpacker.frameBuffer.Write([]byte{0x00, 0x00, 0x00, 0x01})
    return unpacker
}

func (unpacker *H265UnPacker) OnFrame(onframe ON_FRAME_FUNC) {
    unpacker.onFrame = onframe
}

func (unpacker *H265UnPacker) UnPack(pkt []byte) error {
    pkg := &RtpPacket{}
    if err := pkg.Decode(pkt); err != nil {
        return err
    }

    packType := pkg.Payload[0] & 0x1f
    switch {
    case 0 < packType && packType < 40:
        unpacker.frameBuffer.Write(pkg.Payload)
        if unpacker.onFrame != nil {
            unpacker.onFrame(unpacker.frameBuffer.Bytes(), pkg.Header.Timestamp, false)
        }
        unpacker.frameBuffer.Truncate(4)
    case packType == 49:
        unpacker.unpackFu(pkg)
    default:
        return errors.New("unsupport h264 rtp packet type")
    }
    return nil
}

func (unpacker *H265UnPacker) unpackFu(pkt *RtpPacket) error {
    s := pkt.Payload[2] & 0x80
    e := pkt.Payload[2] & 0x40
    if s > 0 {
        if unpacker.frameBuffer.Len() > 4 {
            if unpacker.onFrame != nil {
                unpacker.onFrame(unpacker.frameBuffer.Bytes(), unpacker.timestamp, true)
            }
            unpacker.frameBuffer.Truncate(4)
        }
        unpacker.timestamp = pkt.Header.Timestamp
        unpacker.frameBuffer.WriteByte(pkt.Payload[0]&0x81 | ((pkt.Payload[2] & 0x3F) << 1))
        unpacker.frameBuffer.WriteByte(pkt.Payload[1])
    } else {
        if unpacker.lastSequence+1 != pkt.Header.SequenceNumber {
            unpacker.lost = true
        }
    }
    unpacker.lastSequence = pkt.Header.SequenceNumber
    unpacker.frameBuffer.Write(pkt.Payload[3:])
    if e > 0 {
        if unpacker.onFrame != nil {
            unpacker.onFrame(unpacker.frameBuffer.Bytes(), unpacker.timestamp, unpacker.lost)
        }
        unpacker.frameBuffer.Truncate(4)
    }
    return nil
}
