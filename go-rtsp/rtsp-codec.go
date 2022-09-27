package rtsp

const (
	RTSP_CODEC_H264 = "H264"
	RTSP_CODEC_H265 = "H265"
	RTSP_CODEC_PCMU = "PCMU"
	RTSP_CODEC_PCMA = "PCMA"
)

type RtspCodec struct {
	EncodeName   string //H264,H265,PCMU,PCMA...
	PayloadType  uint8
	ClockRate    uint32
	ChannelCount uint8
}

func NewCodec(name string, pt uint8, clock uint32, channel uint8) RtspCodec {
	return RtspCodec{EncodeName: name, PayloadType: pt, ClockRate: clock, ChannelCount: channel}
}

func NewVideoCodec(name string, pt uint8, clock uint32) RtspCodec {
	return RtspCodec{EncodeName: name, PayloadType: pt, ClockRate: clock}
}

func NewAudioCodec(name string, pt uint8, clock uint32, channelCount int) RtspCodec {
	return RtspCodec{EncodeName: name, PayloadType: pt, ClockRate: clock, ChannelCount: uint8(channelCount)}
}
