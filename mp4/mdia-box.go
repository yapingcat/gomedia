package mp4

func makeMdiaBox(track *mp4track) []byte {
    mdhdbox := makeMdhdBox()
    hdlrbox := makeHdlrBox(getHandlerType(track.cid))
    minfbox := makeMinfBox(track)
    MDIA.Size = 8 + uint64(len(mdhdbox)+len(hdlrbox)+len(minfbox))
    offset, mdiabox := MDIA.Encode()
    copy(mdiabox[offset:], mdhdbox)
    offset += len(mdhdbox)
    copy(mdiabox[offset:], hdlrbox)
    offset += len(hdlrbox)
    copy(mdiabox[offset:], minfbox)
    offset += len(minfbox)
    return mdiabox
}
