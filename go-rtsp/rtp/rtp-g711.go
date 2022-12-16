package rtp

import (
    "errors"
)

type G711Packer struct {
    pt       uint8
    ssrc     uint32
    sequence uint16
    mtu      int
    onpkt    ON_RTP_PKT_FUNC
    onRtp    RTP_HOOK_FUNC
}

func NewG711Packer(pt uint8, ssrc uint32, sequence uint16, mtu int) *G711Packer {
    return &G711Packer{
        pt:       pt,
        ssrc:     ssrc,
        sequence: sequence,
        mtu:      mtu,
    }
}

func (g711 *G711Packer) HookRtp(cb RTP_HOOK_FUNC) {
    g711.onRtp = cb
}

func (g711 *G711Packer) SetMtu(mtu int) {
    g711.mtu = mtu
}

func (g711 *G711Packer) OnPacket(onPkt ON_RTP_PKT_FUNC) {
    g711.onpkt = onPkt
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
    if g711.onpkt != nil {
        return g711.onpkt(pkg.Encode())
    }
    return nil
}

type G711UnPacker struct {
    onFrame ON_FRAME_FUNC
}

func NewG711UnPacker() *G711UnPacker {
    return &G711UnPacker{}
}

func (g711 *G711UnPacker) OnFrame(onframe ON_FRAME_FUNC) {
    g711.onFrame = onframe
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
