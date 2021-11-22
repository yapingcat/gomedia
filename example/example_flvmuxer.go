package main

import (
	"fmt"
	"os"

	"github.com/yapingcat/gomedia/flv"
	"github.com/yapingcat/gomedia/mpeg"
)

func main() {
	flvfilereader, _ := os.Open(os.Args[1])
	defer flvfilereader.Close()
	fr := flv.CreateFlvReader(flvfilereader)
	newflv, _ := os.OpenFile(os.Args[1]+"4.flv", os.O_CREATE|os.O_RDWR, 0666)
	defer newflv.Close()
	fw := flv.CreateFlvWriter(newflv)
	fw.WriteFlvHeader()
	fr.OnFrame = func(ci mpeg.CodecID, b []byte, pts uint32, dts uint32) {
		if ci == mpeg.CODECID_AUDIO_AAC {
			fw.WriteAAC(b, pts, dts)
		} else if ci == mpeg.CODECID_VIDEO_H264 {
			fw.WriteH264(b, pts, dts)
		}
	}
	fmt.Println(fr.LoopRead())
}
