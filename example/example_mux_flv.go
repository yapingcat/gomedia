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
	newflv, _ := os.OpenFile(os.Args[1]+"_new.flv", os.O_CREATE|os.O_RDWR, 0666)
	defer newflv.Close()
	fw := flv.CreateFlvWriter(newflv)
	fw.WriteFlvHeader()
	fr.OnFrame = func(ci codec.CodecID, b []byte, pts uint32, dts uint32) {
		if ci == codec.CODECID_AUDIO_AAC {
			fw.WriteAAC(b, pts, dts)
		} else if ci == codec.CODECID_AUDIO_MP3 {
			fmt.Println("write mp3 frame")
			fw.WriteMp3(b, pts, dts)
		} else if ci == codec.CODECID_VIDEO_H264 {
			fmt.Println("write H264 frame")
			fw.WriteH264(b, pts, dts)
		} else if ci == codec.CODECID_VIDEO_H265 {
			fw.WriteH265(b, pts, dts)
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
