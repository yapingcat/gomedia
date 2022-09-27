package rtsp

import (
    "fmt"
    "strconv"

    "github.com/yapingcat/gomedia/go-rtsp/rtp"
    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

type RtspSample struct {
    Sample    []byte
    Timestamp uint32
}

type OnSampleCallBack func(sample RtspSample)

type RtspTrack struct {
    TrackName    string //video/audio/application
    Codec        RtspCodec
    transport    *RtspTransport
    onSample     OnSampleCallBack
    uri          string
    output       OutputFunc
    pack         rtp.Packer
    unpack       rtp.UnPacker
    isOpen       bool
    paramHandler sdp.CodecParamHandler
}

type OutputFunc func([]byte)

type TrackOption func(t *RtspTrack)

func WithCodecParamHandler(handler sdp.CodecParamHandler) TrackOption {
    return func(t *RtspTrack) {
        t.paramHandler = handler
    }
}

func NewTrack(name string, codec RtspCodec, opt ...TrackOption) *RtspTrack {
    track := &RtspTrack{TrackName: name, Codec: codec}
    for _, o := range opt {
        o(track)
    }
    return track
}

func (track *RtspTrack) SetTransport(transport *RtspTransport) {
    track.transport = transport
}

func (track *RtspTrack) OnSample(onsample OnSampleCallBack) {
    track.onSample = onsample
}

func (track *RtspTrack) SetOutput(f OutputFunc) {
    track.output = f
}

func (track *RtspTrack) WriteSample(sample RtspSample) error {
    switch track.Codec.EncodeName {
    case RTSP_CODEC_H264:
        return track.writeH264(sample)
    case RTSP_CODEC_H265:
        return track.writeH265(sample)
    case RTSP_CODEC_PCMA:
        fallthrough
    case RTSP_CODEC_PCMU:
        return track.writeG711(sample)
    }
    return nil
}

func (track *RtspTrack) OpenTrack() {
    track.isOpen = true
}

func (track *RtspTrack) writeH264(sample RtspSample) error {

}

func (track *RtspTrack) writeH265(sample RtspSample) error {

}

func (track *RtspTrack) writeG711(sample RtspSample) error {

}

func (track *RtspTrack) mediaDescripe() string {
    md := ""
    switch track.Codec.EncodeName {
    case RTSP_CODEC_H264:
        md = fmt.Sprintf("m=video 0 RTP/AVP %d", strconv.Itoa(int(track.Codec.PayloadType)))
        md += fmt.Sprintf("rtpmap:%d H264/%d", track.Codec.PayloadType, track.Codec.ClockRate)
    case RTSP_CODEC_H265:
        md = fmt.Sprintf("m=video 0 RTP/AVP %d", strconv.Itoa(int(track.Codec.PayloadType)))
        md += fmt.Sprintf("rtpmap:%d H265/%d", track.Codec.PayloadType, track.Codec.ClockRate)
    case RTSP_CODEC_PCMA:
        md = fmt.Sprintf("m=audio 0 RTP/AVP %d", strconv.Itoa(int(track.Codec.PayloadType)))
        md += fmt.Sprintf("rtpmap:%d PCMA/%d/%d", track.Codec.PayloadType, track.Codec.ClockRate, track.Codec.ChannelCount)
    case RTSP_CODEC_PCMU:
        md = fmt.Sprintf("m=audio 0 RTP/AVP %d", strconv.Itoa(int(track.Codec.PayloadType)))
        md += fmt.Sprintf("rtpmap:%d PCMU/%d/%d", track.Codec.PayloadType, track.Codec.ClockRate, track.Codec.ChannelCount)
    }
    md += track.paramHandler.Encode()
    md += "a=control:" + track.uri
    return md
}

func (track *RtspTrack) input(data []byte) error {

}
