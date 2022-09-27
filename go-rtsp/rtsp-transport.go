package rtsp

import (
	"fmt"
	"strings"
)

type LowerTransport int

const (
	UDP LowerTransport = iota
	TCP
)

const (
	MODE_PLAY   = "PLAY"
	MODE_RECORD = "RECORD"
)

type RtspTransport struct {
	isMultiCast  bool
	proto        LowerTransport
	client_ports [2]uint16
	server_ports [2]uint16
	interleaved  [2]int
	mode         string
}

type TransportOption func(transport *RtspTransport)

func WithEnableUdp() TransportOption {
	return func(transport *RtspTransport) {
		transport.proto = UDP
	}
}

func WithClientUdpPort(rtpPort uint16, rtcpPort uint16) TransportOption {
	return func(transport *RtspTransport) {
		transport.client_ports[0] = rtpPort
		transport.client_ports[1] = rtcpPort
	}
}

func WithServerUdpPort(rtpPort uint16, rtcpPort uint16) TransportOption {
	return func(transport *RtspTransport) {
		transport.server_ports[0] = rtpPort
		transport.server_ports[1] = rtcpPort
	}
}

func WithTcpInterleaved(interleaved [2]int) TransportOption {
	return func(transport *RtspTransport) {
		transport.interleaved[0] = interleaved[0]
		transport.interleaved[1] = interleaved[1]
	}
}

func WithMode(mode string) TransportOption {
	return func(transport *RtspTransport) {
		transport.mode = mode
	}
}

func NewRtspTransport(opt ...TransportOption) *RtspTransport {
	transport := &RtspTransport{
		isMultiCast: false,
		proto:       TCP,
		mode:        MODE_PLAY,
	}
	for _, o := range opt {
		o(transport)
	}
	return transport
}

func (transport *RtspTransport) SetServerUdpPort(rtpPort uint16, rtcpPort uint16) {
	transport.client_ports[0] = rtpPort
	transport.client_ports[1] = rtcpPort
}

func (transport *RtspTransport) SetClientUdpPort(rtpPort uint16, rtcpPort uint16) {
	transport.client_ports[0] = rtpPort
	transport.client_ports[1] = rtcpPort
}

func (transport *RtspTransport) SetInterleaved(interleaved [2]int) {
	transport.interleaved[0] = interleaved[0]
	transport.interleaved[1] = interleaved[1]
}

// Transport: RTP/AVP;multicast;ttl=127;mode="PLAY",
//            RTP/AVP;unicast;client_port=3456-3457;mode="PLAY"

func (transport *RtspTransport) Decode(data []byte) error {
	return transport.DecodeString(string(data))
}

func (transport *RtspTransport) DecodeString(data string) error {
	items := strings.Split(data, ";")
	for _, item := range items {
		kv := strings.Split(item, "=")
		switch kv[0] {
		case "RTP/AVP/TCP":
			transport.proto = TCP
		case "multicast":
			transport.isMultiCast = true
		case "unicast":
			transport.isMultiCast = false
		case "mode":
			transport.mode = kv[1]
		case "client_port":
			fmt.Sscanf(kv[1], "%d-%d", transport.client_ports[0], transport.client_ports[1])
		case "server_port":
			fmt.Sscanf(kv[1], "%d-%d", transport.server_ports[0], transport.server_ports[1])
		case "interleaved":
			fmt.Sscanf(kv[1], "%d-%d", transport.interleaved[0], transport.interleaved[1])
		}
	}
	return nil
}

func (transport *RtspTransport) EncodeString() string {
	str := ""
	if transport.proto == TCP {
		str += "RTP/AVP/TCP"
	} else {
		str += "RTP/AVP"
	}
	if transport.isMultiCast {
		str += ";multicast"
	}

	if transport.proto == TCP {
		str += fmt.Sprintf(";interleaved=%d-%d", transport.interleaved[0], transport.interleaved[1])
	} else {
		if transport.client_ports[0] != 0 {
			str += fmt.Sprintf(";client_port=%d-%d", transport.client_ports[0], transport.client_ports[1])
		}
		if transport.server_ports[0] != 0 {
			str += fmt.Sprintf(";server_port=%d-%d", transport.server_ports[0], transport.server_ports[1])
		}
	}

	if transport.mode == MODE_PLAY {
		str += ";mode=PLAY"
	} else if transport.mode == MODE_RECORD {
		str += ";mode=RECORD"
	}
	return str
}
