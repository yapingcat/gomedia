package mp4

import (
    "github.com/yapingcat/gomedia/codec"
)

type MP4_CODEC_TYPE int

const (
    MP4_CODEC_H264 MP4_CODEC_TYPE = iota + 1
    MP4_CODEC_H265

    MP4_CODEC_AAC MP4_CODEC_TYPE = iota + 100
    MP4_CODEC_G711A
    MP4_CODEC_G711U
)

func getCodecNameWithCodecId(cid MP4_CODEC_TYPE) [4]byte {
    switch cid {
    case MP4_CODEC_H264:
        return [4]byte{'a', 'v', 'c', '1'}
    case MP4_CODEC_H265:
        return [4]byte{'h', 'v', 'c', '1'}
    case MP4_CODEC_AAC:
        return [4]byte{'m', 'p', '4', 'a'}
    case MP4_CODEC_G711A:
        return [4]byte{'a', 'l', 'a', 'w'}
    case MP4_CODEC_G711U:
        return [4]byte{'u', 'l', 'a', 'w'}
    default:
        panic("unsupport codec id")
    }
}

//ffmpeg isom.c const AVCodecTag ff_mp4_obj_type[]
func getBojecttypeWithCodecId(cid MP4_CODEC_TYPE) uint8 {
    switch cid {
    case MP4_CODEC_H264:
        return 0x21
    case MP4_CODEC_H265:
        return 0x23
    case MP4_CODEC_AAC:
        return 0x40
    case MP4_CODEC_G711A:
        return 0xfd
    case MP4_CODEC_G711U:
        return 0xfe
    default:
        panic("unsupport codec id")
    }
}

func isH264NewAccessUnit(nalu []byte) bool {
    nalu_type := codec.H264NaluType(nalu)
    switch nalu_type {
    case codec.H264_NAL_AUD, codec.H264_NAL_SPS,
        codec.H264_NAL_PPS, codec.H264_NAL_SEI:
        return true
    case codec.H264_NAL_I_SLICE, codec.H264_NAL_P_SLICE,
        codec.H264_NAL_SLICE_A, codec.H264_NAL_SLICE_B, codec.H264_NAL_SLICE_C:
        firstMbInSlice := codec.GetH264FirstMbInSlice(nalu)
        if firstMbInSlice == 0 {
            return true
        }
    }
    return false
}

func isH265NewAccessUnit(nalu []byte) bool {
    nalu_type := codec.H265NaluType(nalu)
    switch nalu_type {
    case codec.H265_NAL_AUD, codec.H265_NAL_SPS,
        codec.H265_NAL_PPS, codec.H265_NAL_SEI, codec.H265_NAL_VPS:
        return true
    case codec.H265_NAL_Slice_TRAIL_N, codec.H265_NAL_LICE_TRAIL_R,
        codec.H265_NAL_SLICE_TSA_N, codec.H265_NAL_SLICE_TSA_R,
        codec.H265_NAL_SLICE_STSA_N, codec.H265_NAL_SLICE_STSA_R,
        codec.H265_NAL_SLICE_RADL_N, codec.H265_NAL_SLICE_RADL_R,
        codec.H265_NAL_SLICE_RASL_N, codec.H265_NAL_SLICE_RASL_R,
        codec.H265_NAL_SLICE_BLA_W_LP, codec.H265_NAL_SLICE_BLA_W_RADL,
        codec.H265_NAL_SLICE_BLA_N_LP, codec.H265_NAL_SLICE_IDR_W_RADL,
        codec.H265_NAL_SLICE_IDR_N_LP, codec.H265_NAL_SLICE_CRA:
        firstMbInSlice := codec.GetH265FirstMbInSlice(nalu)
        if firstMbInSlice == 0 {
            return true
        }
    }
    return false
}
