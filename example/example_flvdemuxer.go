package main

import (
	"fmt"
	"os"

	"github.com/yapingcat/gomedia/flv"
	"github.com/yapingcat/gomedia/mpeg"
)

func main() {

	flvfilereader, _ := os.Open(os.Args[1])
	fr := flv.CreateFlvReader(flvfilereader)
	firstAudio := true
	var audiof *os.File
	firstVideo := true
	var videof *os.File
	fr.OnFrame = func(ci mpeg.CodecID, b []byte, u1, u2 uint32) {
		if ci == mpeg.CODECID_AUDIO_AAC {
			if firstAudio {
				audiof, _ = os.OpenFile("audio.aac", os.O_CREATE|os.O_RDWR, 0666)
				firstAudio = false
			}
			audiof.Write(b)
		} else if ci == mpeg.CODECID_VIDEO_H264 {
			if firstVideo {
				videof, _ = os.OpenFile("video.h264", os.O_CREATE|os.O_RDWR, 0666)
				firstVideo = false
			}
			videof.Write(b)
		}
	}
	fmt.Println(fr.LoopRead())
}
