package mp4

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
