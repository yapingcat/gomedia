package rtsp

import (
    "fmt"
    "math/rand"
    "time"

    "github.com/yapingcat/gomedia/go-codec"
    "github.com/yapingcat/gomedia/go-rtsp/rtp"
    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

func init() {
    rand.Seed(time.Now().Unix())
}

type RtspSample struct {
    Cid       RTSP_CODEC_ID
    Sample    []byte
    Timestamp uint32 //in milliseconds
    Completed bool
}

type OnSampleCallBack func(sample RtspSample)

type RtspTrack struct {
    TrackName    string //video/audio/application
    Codec        RtspCodec
    transport    *RtspTransport
    onSample     OnSampleCallBack
    uri          string
    onPacket     PacketCallBack
    pack         rtp.Packer
    unpack       rtp.UnPacker
    isOpen       bool
    extra        interface{}
    paramHandler sdp.FmtpCodecParamParser
    initSequence uint16
}

type PacketCallBack func(b []byte, isRtcp bool) error
type TrackOption func(t *RtspTrack)

func WithCodecParamHandler(handler sdp.FmtpCodecParamParser) TrackOption {
    return func(t *RtspTrack) {
        t.paramHandler = handler
    }
}

func NewVideoTrack(codec RtspCodec, opt ...TrackOption) *RtspTrack {
    return newTrack("video", codec, opt...)

}
func NewAudioTrack(codec RtspCodec, opt ...TrackOption) *RtspTrack {
    return newTrack("audio", codec, opt...)
}

func NewMetaTrack(codec RtspCodec, opt ...TrackOption) *RtspTrack {
    return newTrack("application", codec, opt...)
}

func newTrack(name string, codec RtspCodec, opt ...TrackOption) *RtspTrack {
    track := &RtspTrack{
        TrackName:    name,
        Codec:        codec,
        initSequence: uint16(rand.Uint32()),
    }
    for _, o := range opt {
        o(track)
    }
    track.unpack = track.createUnpacker()
    track.pack = track.createPacker()
    return track
}

func (track *RtspTrack) EnableTCP() {
    track.transport = NewRtspTransport()
}

func (track *RtspTrack) SetCodecParamHandle(handle sdp.FmtpCodecParamParser) {
    track.paramHandler = handle
}

func (track *RtspTrack) SetTransport(transport *RtspTransport) {
    track.transport = transport
}

func (track *RtspTrack) SetExtraData(extra interface{}) {
    track.extra = extra
}

func (track *RtspTrack) OnSample(onsample OnSampleCallBack) {
    track.onSample = onsample
    hasSps := false
    hasPps := false
    hasVps := false
    track.unpack.OnFrame(func(frame []byte, timestamp uint32, lost bool) {
        sample := RtspSample{
            Cid:       track.Codec.Cid,
            Sample:    frame,
            Timestamp: timestamp * 1000 / track.Codec.SampleRate,
            Completed: !lost,
        }
        if sample.Cid == RTSP_CODEC_H264 {
            nalu_type := codec.H264NaluType(frame)
            switch nalu_type {
            case codec.H264_NAL_SPS:
                hasSps = true
            case codec.H264_NAL_PPS:
                hasPps = true
            }
            if nalu_type == codec.H264_NAL_I_SLICE && (!hasPps || !hasSps) {
                if h264Param, ok := track.paramHandler.(*sdp.H264FmtpParam); ok {
                    sps, pps := h264Param.GetSpsPps()
                    if len(sps) > 0 && len(pps) > 0 {
                        tmpSample := make([]byte, 0, len(sps)+len(pps)+8+len(frame))
                        tmpSample = append(tmpSample, []byte{0x00, 0x00, 0x00, 0x01}...)
                        tmpSample = append(tmpSample, sps...)
                        tmpSample = append(tmpSample, []byte{0x00, 0x00, 0x00, 0x01}...)
                        tmpSample = append(tmpSample, pps...)
                        tmpSample = append(tmpSample, frame...)
                        sample.Sample = tmpSample
                        track.onSample(sample)
                        return
                    }
                }
            }
        } else {
            nalu_type := codec.H265NaluType(frame)
            switch nalu_type {
            case codec.H265_NAL_PPS:
                hasPps = true
            case codec.H265_NAL_SPS:
                hasSps = true
            case codec.H265_NAL_VPS:
                hasVps = true
            }
            if nalu_type >= 16 && nalu_type <= 21 && (!hasPps || !hasSps || !hasVps) {
                if h265Param, ok := track.paramHandler.(*sdp.H265FmtpParam); ok {
                    vps, sps, pps := h265Param.GetVpsSpsPps()
                    if len(vps) > 0 && len(sps) > 0 && len(pps) > 0 {
                        tmpSample := make([]byte, 0, len(vps)+len(sps)+len(pps)+12+len(frame))
                        tmpSample = append(tmpSample, []byte{0x00, 0x00, 0x00, 0x01}...)
                        tmpSample = append(tmpSample, vps...)
                        tmpSample = append(tmpSample, []byte{0x00, 0x00, 0x00, 0x01}...)
                        tmpSample = append(tmpSample, sps...)
                        tmpSample = append(tmpSample, []byte{0x00, 0x00, 0x00, 0x01}...)
                        tmpSample = append(tmpSample, pps...)
                        tmpSample = append(tmpSample, frame...)
                        sample.Sample = tmpSample
                        track.onSample(sample)
                        return
                    }
                }
            }
        }
        track.onSample(sample)
    })
}

func (track *RtspTrack) OnPacket(f PacketCallBack) {
    track.onPacket = f
    track.pack.OnPacket(func(pkt []byte) error {
        return track.onPacket(pkt, false)
    })
}

func (track *RtspTrack) WriteSample(sample RtspSample) error {
    return track.pack.Pack(sample.Sample, sample.Timestamp*track.Codec.SampleRate/1000)

}

func (track *RtspTrack) OpenTrack() {
    track.isOpen = true
}

func (track *RtspTrack) mediaDescripe() string {
    md := fmt.Sprintf("m=%s 0 RTP/AVP %d\r\n", track.TrackName, track.Codec.PayloadType)
    md += fmt.Sprintf("a=control:%s\r\n", track.uri)
    if track.TrackName != "audio" {
        md += fmt.Sprintf("a=rtpmap:%d %s/%d\r\n", track.Codec.PayloadType, GetEncodeNameByCodecId(track.Codec.Cid), track.Codec.SampleRate)
    } else {
        md += fmt.Sprintf("a=rtpmap:%d %s/%d/%d\r\n", track.Codec.PayloadType, GetEncodeNameByCodecId(track.Codec.Cid), track.Codec.SampleRate, track.Codec.ChannelCount)
    }
    if track.paramHandler != nil {
        md += fmt.Sprintf("a=fmtp:%d %s\r\n", track.Codec.PayloadType, track.paramHandler.Save())
    }
    return md
}

func (track *RtspTrack) input(data []byte, isRtcp bool) error {
    //TODO
    if isRtcp {
        return nil
    }
    return track.unpack.UnPack(data)
}

func (track *RtspTrack) createUnpacker() rtp.UnPacker {

    switch track.Codec.Cid {
    case RTSP_CODEC_H264:
        return rtp.NewH264UnPacker()
    case RTSP_CODEC_H265:
        return rtp.NewH265UnPacker()
    case RTSP_CODEC_AAC:
        if aacFmtp, ok := track.paramHandler.(*sdp.AACFmtpParam); ok {
            return rtp.NewAACUnPacker(aacFmtp.SizeLength(), aacFmtp.IndexLength(), aacFmtp.AudioSpecificConfig())
        } else {
            return rtp.NewAACUnPacker(13, 3, nil)
        }
    case RTSP_CODEC_G711A, RTSP_CODEC_G711U:
        return rtp.NewG711UnPacker()
    }
    return nil
}

func (track *RtspTrack) createPacker() rtp.Packer {
    switch track.Codec.Cid {
    case RTSP_CODEC_AAC:
        return rtp.NewAACPacker(track.Codec.PayloadType, rand.Uint32(), track.initSequence, 1400)
    case RTSP_CODEC_H264:
        return rtp.NewH264Packer(track.Codec.PayloadType, rand.Uint32(), track.initSequence, 1400)
    case RTSP_CODEC_H265:
        return rtp.NewH265Packer(track.Codec.PayloadType, rand.Uint32(), track.initSequence, 1400)
    case RTSP_CODEC_G711U, RTSP_CODEC_G711A:
        return rtp.NewG711Packer(track.Codec.PayloadType, rand.Uint32(), track.initSequence, 1400)
    case RTSP_CODEC_PS:
        return nil
    case RTSP_CODEC_TS:
        return nil
    default:
        return nil
    }
}
