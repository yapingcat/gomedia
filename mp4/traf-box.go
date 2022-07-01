package mp4

func makeTraf(track *mp4track, moofOffset uint64) []byte {
	tfhd := makeTfhdBox(track, moofOffset)
	tfdt := makeTfdtBox(track)

	traf := BasicBox{Type: [4]byte{'t', 'r', 'a', 'f'}}
	traf.Size = 8 + uint64(len(tfhd)+len(tfdt))
	offset, boxData := traf.Encode()
	copy(boxData[offset:], tfhd)
	offset += len(tfhd)
	copy(boxData[offset:], tfdt)
	offset += len(tfdt)
	return boxData
}
