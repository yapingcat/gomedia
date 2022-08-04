package rtsp

import "github.com/yapingcat/gomedia/go-rtsp/rtp"

type RtspTrack struct {
	TrackName string //video/audio/application
	Codec     RtspCodec
	Local     RtspTransport
	Remote    RtspTransport
	output    OutputFunc
	pack      rtp.Packer
	unpack    rtp.UnPacker
}

type OutputFunc func([]byte)

func (track *RtspTrack) SetOutput(f OutputFunc) {
	track.output = f
}

func (track *RtspTrack) WriteSample(sample []byte, timestamp uint32) {

}

func (track *RtspTrack) Input(data []byte) {

}
