package flv

import "github.com/yapingcat/gomedia/mpeg"

func PutUint24(b []byte, v uint32) {
    _ = b[2]
    b[0] = byte(v >> 16)
    b[1] = byte(v >> 8)
    b[2] = byte(v)
}

func GetUint24(b []byte) (v uint32) {
    _ = b[2]
    v = uint32(b[0])
    v = (v << 8) | uint32(b[1])
    v = (v << 8) | uint32(b[2])
    return v
}

func CovertFlvVideoCodecId2MpegCodecId(cid FLV_VIDEO_CODEC_ID) mpeg.CodecID {
    if cid == FLV_AVC {
        return mpeg.CODECID_VIDEO_H264
    } else if cid == FLV_HEVC {
        return mpeg.CODECID_VIDEO_H265
    }
    return mpeg.CODECID_UNRECOGNIZED
}

func CovertFlvAudioCodecId2MpegCodecId(cid FLV_SOUND_FORMAT) mpeg.CodecID {
    if cid == FLV_AAC {
        return mpeg.CODECID_AUDIO_AAC
    } else if cid == FLV_G711A {
        return mpeg.CODECID_AUDIO_G711A
    } else if cid == FLV_G711U {
        return mpeg.CODECID_AUDIO_G711U
    }
    return mpeg.CODECID_UNRECOGNIZED
}

func GetTagLenByAudioCodec(cid FLV_SOUND_FORMAT) int {
    if cid == FLV_AAC {
        return 2
    } else {
        return 1
    }
}

func GetTagLenByVideoCodec(cid FLV_VIDEO_CODEC_ID) int {
    if cid == FLV_AVC || cid == FLV_HEVC {
        return 5
    } else {
        return 1
    }
}
