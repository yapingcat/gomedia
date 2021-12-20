package mp4

type MOV_CODEC_TYPE int

const (
    MOV_CODEC_H264 MOV_CODEC_TYPE = iota + 1
    MOV_CODEC_H265

    MOV_CODEC_AAC MOV_CODEC_TYPE = iota + 100
    MOV_CODEC_G711A
    MOV_CODEC_G711U
)

func TranscodeCidToCodecName(cid MOV_CODEC_TYPE) [4]byte {
    switch cid {
    case MOV_CODEC_H264:
        return [4]byte{'a', 'v', 'c', '1'}
    case MOV_CODEC_H265:
        return [4]byte{'h', 'v', 'c', '1'}
    case MOV_CODEC_AAC:
        return [4]byte{'m', 'p', '4', 'a'}
    case MOV_CODEC_G711A:
        return [4]byte{'a', 'l', 'a', 'w'}
    case MOV_CODEC_G711U:
        return [4]byte{'u', 'l', 'a', 'w'}
    default:
        panic("unsupport codec id")
    }
}
