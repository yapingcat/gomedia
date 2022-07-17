package main

import (
	"fmt"
	"os"

	"github.com/yapingcat/gomedia/mp4"
)

func main() {
	fmp4File := os.Args[1]
	fmp4, err := os.Open(fmp4File)
	if err != nil {
		fmt.Println(err)
		return
	}
	videof, _ := os.OpenFile("fmp4.h264", os.O_CREATE|os.O_RDWR, 0666)
	audiof, _ := os.OpenFile("fmp4.aac", os.O_CREATE|os.O_RDWR, 0666)
	demuxer := mp4.CreateMp4Demuxer(fmp4)
	infos, _ := demuxer.ReadHead()
	for _, info := range infos {
		fmt.Printf("%v\n", info)
	}

	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			break
		}
		fmt.Println(pkg.Cid, pkg.Pts, pkg.Dts, len(pkg.Data))
		if pkg.Cid == mp4.MP4_CODEC_H264 {
			videof.Write(pkg.Data)
		} else if pkg.Cid == mp4.MP4_CODEC_AAC {
			audiof.Write(pkg.Data)
		}
	}
}
