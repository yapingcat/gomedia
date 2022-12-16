package rtcp

import "github.com/yapingcat/gomedia/go-rtsp/rtp"

//https://www.rfc-editor.org/rfc/rfc3550#section-17

type RtcpContext struct {
	ssrc          uint32
	maxseq        uint16
	cycles        uint32
	baseSeq       uint32
	badSeq        uint32
	probation     uint32
	received      uint32
	expectPrior   uint32
	receivedPrior uint32
	transit       uint32
	jitter        uint32
}

func NewRtcpContext(ssrc uint32, seq uint16) *RtcpContext {
	return &RtcpContext{
		ssrc:    ssrc,
		baseSeq: uint32(seq),
		maxseq:  seq,
		badSeq:  65537,
	}
}

func (ctx *RtcpContext) ComputeTransmitInterval() float32 {

}

func (ctx *RtcpContext) GenerateSR() *SenderReport {

}

func (ctx *RtcpContext) GenerateRR() *ReceiverReport {

}

func (ctx *RtcpContext) AddRtp(pkg *rtp.RtpPacket) {

}
