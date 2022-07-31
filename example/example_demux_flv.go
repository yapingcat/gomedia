package main

import (
	"fmt"
	"os"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
)

func main() {

	flvfilereader, _ := os.Open(os.Args[1])
	defer flvfilereader.Close()
	fr := flv.CreateFlvReader()
	firstAudio := true
	var audiof *os.File
	firstVideo := true
	var videof *os.File
	fr.OnFrame = func(ci codec.CodecID, b []byte, u1, u2 uint32) {
		if ci == codec.CODECID_AUDIO_AAC {
			if firstAudio {
				audiof, _ = os.OpenFile("audio.aac", os.O_CREATE|os.O_RDWR, 0666)
				firstAudio = false
			}
			audiof.Write(b)
		} else if ci == codec.CODECID_VIDEO_H264 {
			if firstVideo {
				videof, _ = os.OpenFile("video.h264", os.O_CREATE|os.O_RDWR, 0666)
				firstVideo = false
			}
			videof.Write(b)
		} else if ci == codec.CODECID_VIDEO_H265 {
			if firstVideo {
				videof, _ = os.OpenFile("video.h265", os.O_CREATE|os.O_RDWR, 0666)
				firstVideo = false
			}
			videof.Write(b)
		}
	}

	cache := make([]byte, 4096)
	for {
		n, err := flvfilereader.Read(cache)
		if err != nil {
			fmt.Println(err)
			break
		}
		fr.Input(cache[0:n])
	}

}
