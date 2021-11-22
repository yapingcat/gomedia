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
	newflv, _ := os.OpenFile(os.Args[1]+"3.flv", os.O_CREATE|os.O_RDWR, 0666)
	defer newflv.Close()
	fw := flv.CreateFlvWriter(newflv)
	fw.WriteFlvHeader()
	fr.OnFrame = func(ci mpeg.CodecID, b []byte, u1, u2 uint32) {
		if ci == mpeg.CODECID_AUDIO_AAC {
			fw.WriteAAC(b, u1, u2)
		} else if ci == mpeg.CODECID_VIDEO_H264 {
			fw.WriteH264(b, u1, u2)
		}
	}
	fmt.Println(fr.LoopRead())
}
