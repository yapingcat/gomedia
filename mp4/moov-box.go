package mp4

type MoovBox struct {
    mvhd *MovieHeaderBox
    trak *TrackHeaderBox
}
