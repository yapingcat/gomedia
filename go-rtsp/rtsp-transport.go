package rtsp

type LowerTransport int

const (
	TCP LowerTransport = iota
	UDP
)

type RtspTransport struct {
	proto       LowerTransport
	ports       [2]uint16
	interleaved [2]int
	mode        string
}

func NewRtspTransport() *RtspTransport {
	return &RtspTransport{
		proto: UDP,
	}
}

type TransportOption func(transport *RtspTransport)

func WithUdpPort(rtpPort uint16, rtcpPort uint16) TransportOption {
	return func(transport *RtspTransport) {
		transport.ports[0] = rtpPort
		transport.ports[1] = rtcpPort
	}
}

func WithTcpInterleaved(interleaved [2]int) TransportOption {
	return func(transport *RtspTransport) {
		transport.interleaved[0] = interleaved[0]
		transport.interleaved[1] = interleaved[1]
	}
}
