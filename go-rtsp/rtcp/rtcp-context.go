package rtcp

import (
	"time"

	"github.com/yapingcat/gomedia/go-rtsp/rtp"
)

//https://www.rfc-editor.org/rfc/rfc3550#section-17

type RtcpContext struct {
    ssrc             uint32
    senderSSRC       uint32
    maxSeq           uint16
    cycles           uint32
    baseSeq          uint32
    badSeq           uint32
    probation        uint32
    received         uint32
    expectPrior      uint32
    receivedPrior    uint32
    transit          uint32
    jitter           float64
    srNtpLsr         uint64
    srClockTime      uint64
    lastRtpClock     uint64
    lastRtpTimestamp uint32
    sampleRate       uint32
    sendBytes        uint64
    sendPackets      uint64
    bindwidth        int
}

const (
    MIN_SEQUENTIAL = 2
    RTP_SEQ_MOD    = 1 << 16
    MAX_DROPOUT    = 3000
    MAX_MISORDER   = 100
)

func NewRtcpContext(ssrc uint32, seq uint16, sampleRate uint32) *RtcpContext {
    return &RtcpContext{
        ssrc:       ssrc,
        maxSeq:     seq - 1,
        probation:  MIN_SEQUENTIAL,
        sampleRate: sampleRate,
    }
}

//rfc 3550 Computing the RTCP Transmission Interval
// func (ctx *RtcpContext) ComputeTransmitInterval() float32 {
// }

func (ctx *RtcpContext) GenerateApp(name string, data []byte) *App {
    app := NewApp()
    app.SSRC = ctx.ssrc
    app.Name = []byte(name)
    app.AppData = make([]byte, len(data))
    copy(app.AppData, data)
    return app
}

func (ctx *RtcpContext) GenerateBye() *Bye {
    bye := NewBye()
    bye.SC = 1
    bye.SSRCS = make([]uint32, 1)
    bye.SSRCS[0] = ctx.ssrc
    return bye
}

func (ctx *RtcpContext) GenerateSDES(sdesType uint8, txt string) *SourceDescription {
    sdes := NewSourceDescription()
    sdes.SC = 1
    sdes.Chunks = make([]SDESChunk, 1)
    sdes.Chunks[0].SSRC = ctx.ssrc
    sdes.Chunks[0].Item.Length = uint8(len(txt))
    sdes.Chunks[0].Item.Txt = []byte(txt)
    sdes.Chunks[0].Item.Type = sdesType
    return sdes
}

func (ctx *RtcpContext) GenerateSR() *SenderReport {
    sr := &SenderReport{Comm: Comm{PT: RTCP_SR}, RC: 0, SSRC: ctx.ssrc, RtpTimestamp: ctx.lastRtpTimestamp, SendPacketCount: uint32(ctx.sendPackets), SendOctetCount: uint32(ctx.sendBytes)}
    sr.NTP = UtcClockToNTP(time.Now())
    return sr
}

func (ctx *RtcpContext) GenerateRR() *ReceiverReport {
    rr := &ReceiverReport{Comm: Comm{PT: RTCP_RR}, RC: 1, SSRC: ctx.ssrc, Blocks: make([]ReportBlock, 1)}
    rr.Blocks[0].SSRC = ctx.senderSSRC
    block := ctx.getReportBlock()
    rr.Blocks[0] = block
    return rr
}

func (ctx *RtcpContext) ReceivedSR(sr *SenderReport) {
    ctx.srClockTime = uint64(time.Now().UnixMicro())
    ctx.srNtpLsr = sr.NTP
    ctx.senderSSRC = sr.SSRC
}

func (ctx *RtcpContext) SendRtp(pkt *rtp.RtpPacket) {
    ctx.sendBytes += uint64(len(pkt.Payload))
    ctx.sendPackets++
    ctx.lastRtpTimestamp = pkt.Header.Timestamp
}

//RFC3550 A.8 Estimating the Interarrival Jitter
// int transit = arrival - r->ts;
// int d = transit - s->transit;
// s->transit = transit;
// if (d < 0) d = -d;
// s->jitter += (1./16.) * ((double)d - s->jitter);

func (ctx *RtcpContext) ReceivedRtp(pkt *rtp.RtpPacket) {
    if ctx.updateSeq(pkt.Header.SequenceNumber) == 0 {
        return
    }
    rtpClock := uint64(time.Now().UnixMicro())
    if ctx.lastRtpClock == 0 {
        ctx.lastRtpClock = uint64(time.Now().UnixMicro())
        ctx.lastRtpTimestamp = pkt.Header.Timestamp
        ctx.jitter = 0
    } else {
        D := int64((rtpClock-ctx.lastRtpClock)*uint64(ctx.sampleRate)/1000000) - (int64(pkt.Header.Timestamp - ctx.lastRtpTimestamp))
        if D < 0 {
            D = -1 * D
        }
        ctx.jitter += (float64(D) - ctx.jitter) / 16
    }
    ctx.lastRtpClock = rtpClock
    ctx.lastRtpTimestamp = pkt.Header.Timestamp
}

func (ctx *RtcpContext) updateSeq(seq uint16) int {
    delta := seq - ctx.maxSeq
    if ctx.probation > 0 {
        if seq == ctx.maxSeq+1 {
            ctx.probation--
            ctx.maxSeq = seq
            if ctx.probation == 0 {
                ctx.initSeq(seq)
                ctx.received++
                return 1
            }
        } else {
            ctx.probation = MIN_SEQUENTIAL - 1
            ctx.maxSeq = seq
        }
        return 0
    } else if delta < MAX_DROPOUT {
        if seq < ctx.maxSeq {
            ctx.cycles += RTP_SEQ_MOD
        }
        ctx.maxSeq = seq
    } else if delta <= RTP_SEQ_MOD-MAX_MISORDER {
        if seq == uint16(ctx.badSeq) {
            ctx.initSeq(seq)
        } else {
            ctx.badSeq = uint32((seq + 1) & (RTP_SEQ_MOD - 1))
            return 0
        }
    } else {
        /* duplicate or reordered packet */
    }
    ctx.received++
    return 1
}

func (ctx *RtcpContext) getReportBlock() ReportBlock {
    rb := ReportBlock{}
    extendMax := ctx.cycles + uint32(ctx.maxSeq)
    expected := extendMax - ctx.baseSeq + 1
    lost := expected - ctx.received
    expectedInterval := expected - ctx.expectPrior
    ctx.expectPrior = expected
    receivedInterval := ctx.received - ctx.receivedPrior
    ctx.receivedPrior = ctx.received
    lostInterval := expectedInterval - receivedInterval
    fraction := uint32(0)
    if expectedInterval == 0 || lostInterval < 0 {
        fraction = 0
    } else {
        fraction = (lostInterval << 8) / expectedInterval
    }

    delay := time.Now().UnixMicro() - int64(ctx.srClockTime)
    lsr := ctx.srNtpLsr >> 8 & 0xFFFFFFFF
    dlsr := uint32(float32(delay) / 1000000 * 65536)

    rb.Lost = lost
    rb.Fraction = uint8(fraction)
    rb.ExtendHighestSeq = extendMax
    rb.Lsr = uint32(lsr)
    rb.Dlsr = dlsr
    rb.SSRC = ctx.senderSSRC
    rb.Jitter = uint32(ctx.jitter)
    return rb
}

func (ctx *RtcpContext) initSeq(seq uint16) {
    ctx.baseSeq = uint32(seq)
    ctx.maxSeq = seq
    ctx.badSeq = RTP_SEQ_MOD + 1 /* so seq == bad_seq is false */
    ctx.cycles = 0
    ctx.received = 0
    ctx.receivedPrior = 0
    ctx.expectPrior = 0
}
