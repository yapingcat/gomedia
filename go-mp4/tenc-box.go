package mp4
import "io"

func decodeTencBox(demuxer *MovDemuxer, size uint32) (err error) {
	buf := make([]byte, size-BasicBoxLen)
	if _, err = io.ReadFull(demuxer.reader, buf); err != nil {
		return
	}
	n := 6
	track := demuxer.tracks[len(demuxer.tracks)-1]
	track.defaultIsProtected = buf[n]
	n += 1
	track.defaultPerSampleIVSize = buf[n]
	n += 1
	copy(track.defaultKID[:], buf[n:])
	return nil
}
