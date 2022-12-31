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

func (g711 *G711Packer) Pack(data []byte, timestamp uint32) error {
    if len(data)+4+RTP_FIX_HEAD_LEN > g711.mtu {
        return errors.New("unsupport fragment g711 into multi rtp packet")
    }
    pkg := RtpPacket{}
    pkg.Header.PayloadType = g711.pt
    pkg.Header.SequenceNumber = g711.sequence
    pkg.Header.SSRC = g711.ssrc
    pkg.Header.Timestamp = timestamp
    pkg.Header.Marker = 1
    pkg.Payload = make([]byte, len(data))
    copy(pkg.Payload, data)
    if g711.onRtp != nil {
        g711.onRtp(&pkg)
    }
    if g711.onPacket != nil {
        return g711.onPacket(pkg.Encode())
    }
    return nil
}

type G711UnPacker struct {
    CommUnPacker
}

func NewG711UnPacker() *G711UnPacker {
    return &G711UnPacker{}
}

func (g711 *G711UnPacker) UnPack(pkt []byte) error {
    pkg := &RtpPacket{}
    if err := pkg.Decode(pkt); err != nil {
        return err
    }
    if g711.onFrame != nil {
        g711.onFrame(pkg.Payload, pkg.Header.Timestamp, false)
    }
    return nil
}
