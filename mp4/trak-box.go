package mp4

func makeTrak(track *mp4track) []byte {
    track.makeStblTable()
    tkhd := makeTkhdBox(track)
    edts := makeEdtsBox(track)
    mdia := makeMdiaBox(track)

    trak := BasicBox{Type: [4]byte{'t', 'r', 'a', 'k'}}
    trak.Size = 8 + uint64(len(tkhd)+len(edts)+len(mdia))
    offset, trakBox := trak.Encode()
    copy(trakBox[offset:], tkhd)
    offset += len(tkhd)
    copy(trakBox[offset:], edts)
    offset += len(edts)
    copy(trakBox[offset:], mdia)
    return trakBox
}
