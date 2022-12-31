package rtp

import (
    "encoding/binary"
    "errors"

    "github.com/yapingcat/gomedia/go-codec"
)

//RFC3640
// mpeg4-generic
// +---------+-----------+-----------+---------------+
// | RTP     | AU Header | Auxiliary | Access Unit   |
// | Header  | Section   | Section   | Data Section  |
// +---------+-----------+-----------+---------------+
// 	<----------RTP Packet Payload----------->
//
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+- .. -+-+-+-+-+-+-+-+-+-+
// |AU-headers-length|AU-header|AU-header|      |AU-header|padding|
// |                 |   (1)   |   (2)   |      |   (n)   | bits  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+- .. -+-+-+-+-+-+-+-+-+-+

// Au-headers-length 2 bytes

//      Au-header
// +---------------------------------------+
// |     AU-size                           |
// +---------------------------------------+
// |     AU-Index / AU-Index-delta         |
// +---------------------------------------+
// |     CTS-flag                          |
// +---------------------------------------+
// |     CTS-delta                         |
// +---------------------------------------+
// |     DTS-flag                          |
// +---------------------------------------+
// |     DTS-delta                         |
// +---------------------------------------+
// |     RAP-flag                          |
// +---------------------------------------+
// |     Stream-state                      |
// +---------------------------------------+

type AACPacker struct {
    CommPacker
    pt       uint8
    ssrc     uint32
    sequence uint16
}

func NewAACPacker(pt uint8, ssrc uint32, sequence uint16, mtu int) *AACPacker {
    return &AACPacker{
        pt:         pt,
        ssrc:       ssrc,
        sequence:   sequence,
        CommPacker: CommPacker{mtu: mtu},
    }
}

func (aac *AACPacker) Pack(data []byte, timestamp uint32) error {
    if len(data)+4+RTP_FIX_HEAD_LEN > aac.mtu {
        return errors.New("unsupport fragment aac into multi rtp packet")
    }
    pkg := RtpPacket{}
    pkg.Header.PayloadType = aac.pt
    pkg.Header.SequenceNumber = aac.sequence
    pkg.Header.SSRC = aac.ssrc
    pkg.Header.Timestamp = timestamp
    pkg.Header.Marker = 1
    pkg.Payload = make([]byte, 4+len(data))
    binary.BigEndian.PutUint16(pkg.Payload, 16)
    size := uint16(len(data))
    pkg.Payload[2] = uint8(size >> 5)
    pkg.Payload[3] = uint8((size & 0x1F) << 3)
    copy(pkg.Payload[4:], data)
    aac.sequence++
    if aac.onRtp != nil {
        aac.onRtp(&pkg)
    }
    if aac.onPacket != nil {
        return aac.onPacket(pkg.Encode())
    }

    return nil
}

type AACUnPacker struct {
    CommUnPacker
    sizeLenth   int
    indexLength int
    asc         []byte
}

func NewAACUnPacker(sizeLength int, indexLength int, asc []byte) *AACUnPacker {
    unpacker := &AACUnPacker{
        sizeLenth:   sizeLength,
        indexLength: indexLength,
        asc:         make([]byte, len(asc)),
    }
    copy(unpacker.asc, asc)
    return unpacker
}

func (aac *AACUnPacker) UnPack(pkt []byte) error {
    pkg := &RtpPacket{}
    if err := pkg.Decode(pkt); err != nil {
        return err
    }
    if len(pkg.Payload) < 4 {
        return errors.New("aac rtp pakcet less than 4 byte")
    }

    if aac.onRtp != nil {
        aac.onRtp(pkg)
    }

    headLength := (binary.BigEndian.Uint16(pkg.Payload) + 7) / 8
    auNum := headLength / 2
    pkg.Payload = pkg.Payload[2:]
    tmp := make([]int, auNum)
    for i := 0; i < int(auNum); i++ {
        bs := codec.NewBitStream(pkg.Payload)
        tmp[i] = int(bs.Uint16(aac.sizeLenth))
        pkg.Payload = pkg.Payload[2:]
    }

    for i := 0; i < len(tmp); i++ {
        var adts []byte
        if len(aac.asc) > 0 {
            adtsHdr, _ := codec.ConvertASCToADTS(aac.asc, tmp[i]+7)
            adts = adtsHdr.Encode()
        }
        adts = append(adts, pkg.Payload[:tmp[i]]...)
        if aac.onFrame != nil {
            aac.onFrame(adts, pkg.Header.Timestamp, false)
        }
        pkg.Payload = pkg.Payload[tmp[i]:]
    }
    return nil
}
