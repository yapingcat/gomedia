package rtsp

const (
	RTSP_CODEC_H264 = "H264"
	RTSP_CODEC_H265 = "H265"
	RTSP_CODEC_PCMU = "PCMU"
	RTSP_CODEC_PCMA = "PCMA"
)

type RtspCodec struct {
	EncodeName  string //H264,H265,PCMU,PCMA...
	PayloadType uint8
	ClockRate   uint16
}
