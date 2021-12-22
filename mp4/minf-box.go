package mp4

func makeMinfBox(track *mp4track) []byte {
    var mhdbox []byte
    switch track.cid {
    case MP4_CODEC_H264:
        fallthrough
    case MP4_CODEC_H265:
        mhdbox = makeVmhdBox()
    case MP4_CODEC_G711A:
        fallthrough
    case MP4_CODEC_G711U:
        fallthrough
    case MP4_CODEC_AAC:
        mhdbox = makeSmhdBox()
    default:
        panic("unsupport codec id")
    }
    dinfbox := makeDefaultDinfBox()
    stblbox := makeStblBox(track)
    MINF.Size = 8 + uint64(len(mhdbox)+len(dinfbox)+len(stblbox))
    offset, minfbox := MINF.Encode()
    copy(minfbox[offset:], mhdbox)
    offset += len(mhdbox)
    copy(minfbox[offset:], dinfbox)
    offset += len(dinfbox)
    copy(minfbox[offset:], stblbox)
    offset += len(stblbox)
    return minfbox
}
