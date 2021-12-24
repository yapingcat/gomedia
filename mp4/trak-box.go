package mp4

func makeTrak(track *mp4track) []byte {
    track.makeStblTable()
    tkhd := makeTkhdBox(track)
    mdia := makeMdiaBox(track)

    TRAK.Size = 8 + uint64(len(tkhd)+len(mdia))
    offset, trak := TRAK.Encode()
    copy(trak[offset:], tkhd)
    offset += len(tkhd)
    copy(trak[offset:], mdia)
    return trak
}
