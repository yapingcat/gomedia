package rtp

import (
    "errors"
)

type G711Packer struct {
    CommPacker
    pt       uint8
    ssrc     uint32
    sequence uint16
}

func NewG711Packer(pt uint8, ssrc uint32, sequence uint16, mtu int) *G711Packer {
    return &G711Packer{
        pt:         pt,
        ssrc:       ssrc,
        sequence:   sequence,
        CommPacker: CommPacker{mtu: mtu},
    }
}

func (packer *G711Packer) Pack(data []byte, timestamp uint32) error {
    if len(data)+RTP_FIX_HEAD_LEN > packer.mtu {
        return errors.New("g711 frame size too large than mtu")
    }
    pkg := RtpPacket{}
    pkg.Header.PayloadType = packer.pt
    pkg.Header.SequenceNumber = packer.sequence
    pkg.Header.SSRC = packer.ssrc
    pkg.Header.Timestamp = timestamp
    pkg.Header.Marker = 1
    pkg.Payload = make([]byte, len(data))
    copy(pkg.Payload, data)
    if packer.onRtp != nil {
        packer.onRtp(&pkg)
    }
    if packer.onPacket != nil {
        return packer.onPacket(pkg.Encode())
    }
    packer.sequence++
    return nil
}

type G711UnPacker struct {
    CommUnPacker
}

func NewG711UnPacker() *G711UnPacker {
    return &G711UnPacker{}
}

func (unpacker *G711UnPacker) UnPack(pkt []byte) error {
    pkg := &RtpPacket{}
    if err := pkg.Decode(pkt); err != nil {
        return err
    }

    if unpacker.onRtp != nil {
        unpacker.onRtp(pkg)
    }

    if unpacker.onFrame != nil {
        unpacker.onFrame(pkg.Payload, pkg.Header.Timestamp, false)
    }
    return nil
}
