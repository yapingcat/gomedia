package mp4

func makeTraf(track *mp4track, moofOffset uint64, moofSize uint64) []byte {
	tfhd := makeTfhdBox(track, moofOffset)
	trun := makeTrunBoxes(track, moofSize)
	tfdt := makeTfdtBox(track)

	traf := BasicBox{Type: [4]byte{'t', 'r', 'a', 'f'}}
	traf.Size = 8 + uint64(len(tfhd)+len(tfdt)+len(trun))
	offset, boxData := traf.Encode()
	copy(boxData[offset:], tfhd)
	offset += len(tfhd)
	copy(boxData[offset:], trun)
	offset += len(trun)
	copy(boxData[offset:], tfdt)
	offset += len(tfdt)
	return boxData
}
