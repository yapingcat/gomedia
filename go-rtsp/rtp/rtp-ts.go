package rtp

import (
    "errors"
)

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

func (ts *TsPacker) Pack(data []byte, timestamp uint32) error {
    if len(data)+4+RTP_FIX_HEAD_LEN > ts.mtu {
        return errors.New("unsupport fragment g711 into multi rtp packet")
    }
    pkg := RtpPacket{}
    pkg.Header.PayloadType = ts.pt
    pkg.Header.SequenceNumber = ts.sequence
    pkg.Header.SSRC = ts.ssrc
    pkg.Header.Timestamp = timestamp
    pkg.Header.Marker = 1
    pkg.Payload = make([]byte, len(data))
    copy(pkg.Payload, data)
    if ts.onRtp != nil {
        ts.onRtp(&pkg)
    }
    if ts.onPacket != nil {
        return ts.onPacket(pkg.Encode())
    }
    return nil
}

type TsUnPacker struct {
    CommUnPacker
}

func NewTsUnPacker() *TsUnPacker {
    return &TsUnPacker{}
}

func (ts *TsUnPacker) UnPack(pkt []byte) error {
    pkg := &RtpPacket{}
    if err := pkg.Decode(pkt); err != nil {
        return err
    }
    if ts.onFrame != nil {
        ts.onFrame(pkg.Payload, pkg.Header.Timestamp, false)
    }
    return nil
}
