package mp4

func makeStblBox(track *mp4track) []byte {
	var stsdbox []byte
	var sttsbox []byte
	var cttsbox []byte
	var stscbox []byte
	var stszbox []byte
	var stcobox []byte
	stsdbox = makeStsd(track, getHandlerType(track.cid))
	if track.stbltable.stts != nil {
		sttsbox = makeStts(track.stbltable.stts)
	}
	if track.stbltable.ctts != nil {
		cttsbox = makeCtts(track.stbltable.ctts)
	}
	if track.stbltable.stsc != nil {
		stscbox = makeStsc(track.stbltable.stsc)
	}
	if track.stbltable.stsz != nil {
		stszbox = makeStsz(track.stbltable.stsz)
	}
	if track.stbltable.stco != nil {
		stcobox = makeStco(track.stbltable.stco)
	}

	STBL.Size = uint64(8 + len(stsdbox) + len(sttsbox) + len(cttsbox) + len(stscbox) + len(stszbox) + len(stcobox))
	offset, stblbox := STBL.Encode()
	copy(stblbox[offset:], stsdbox)
	offset += len(stsdbox)
	copy(stblbox[offset:], sttsbox)
	offset += len(sttsbox)
	copy(stblbox[offset:], cttsbox)
	offset += len(cttsbox)
	copy(stblbox[offset:], stscbox)
	offset += len(stscbox)
	copy(stblbox[offset:], stszbox)
	offset += len(stszbox)
	copy(stblbox[offset:], stcobox)
	offset += len(stcobox)
	return stblbox
}
