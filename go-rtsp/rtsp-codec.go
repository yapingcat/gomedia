package rtsp

import (
    "strings"
)

type RTSP_CODEC_ID int

const (
    RTSP_CODEC_H264 RTSP_CODEC_ID = iota
    RTSP_CODEC_H265
    RTSP_CODEC_AAC
    RTSP_CODEC_G711A
    RTSP_CODEC_G711U
    RTSP_CODEC_PS
    RTSP_CODEC_TS
)

type RtspCodec struct {
    Cid          RTSP_CODEC_ID //H264,H265,PCMU,PCMA...
    PayloadType  uint8
    SampleRate   uint32
    ChannelCount uint8
}

func GetCodecIdByEncodeName(name string) RTSP_CODEC_ID {
    lowName := strings.ToLower(name)
    switch lowName {
    case "h264":
        return RTSP_CODEC_H264
    case "h265":
        return RTSP_CODEC_H265
    case "mpeg4-generic", "mpeg4-latm":
        return RTSP_CODEC_AAC
    case "pcmu":
        return RTSP_CODEC_G711A
    case "pcma":
        return RTSP_CODEC_G711U
    case "mp2t":
        return RTSP_CODEC_TS
    }
    panic("unsupport codec")
}

func GetEncodeNameByCodecId(cid RTSP_CODEC_ID) string {
    switch cid {
    case RTSP_CODEC_H264:
        return "H264"
    case RTSP_CODEC_H265:
        return "H265"
    case RTSP_CODEC_AAC:
        return "mpeg4-generic"
    case RTSP_CODEC_G711A:
        return "pcma"
    case RTSP_CODEC_G711U:
        return "pcmu"
    case RTSP_CODEC_PS:
        return "MP2P"
    case RTSP_CODEC_TS:
        return "MP2T"
    default:
        panic("unsupport rtsp codec id")
    }
}

func NewCodec(name string, pt uint8, sampleRate uint32, channel uint8) RtspCodec {
    return RtspCodec{Cid: GetCodecIdByEncodeName(name), PayloadType: pt, SampleRate: sampleRate, ChannelCount: channel}
}

func NewVideoCodec(name string, pt uint8, sampleRate uint32) RtspCodec {
    return RtspCodec{Cid: GetCodecIdByEncodeName(name), PayloadType: pt, SampleRate: sampleRate}
}

func NewAudioCodec(name string, pt uint8, sampleRate uint32, channelCount int) RtspCodec {
    return RtspCodec{Cid: GetCodecIdByEncodeName(name), PayloadType: pt, SampleRate: sampleRate, ChannelCount: uint8(channelCount)}
}

func NewApplicatioCodec(name string, pt uint8) RtspCodec {
    return RtspCodec{Cid: GetCodecIdByEncodeName(name), PayloadType: pt}
}
