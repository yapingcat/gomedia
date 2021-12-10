package mp4

type TrackReferenceBox struct {
}

type TrackReferenceTypeBox struct {
}

type TrackBox struct {
	Tkhd *TrackHeaderBox
	Tref *TrackReferenceBox
	Trtb *TrackReferenceTypeBox
}
