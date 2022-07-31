package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yapingcat/gomedia/go-mp4"
)

var mp4filename = flag.String("mp4file", "test.mp4", "mp4 file you want to decode")
var rawvideo = flag.String("videofile", "v.h264", "export raw video data to the videofile")
var rawaudio = flag.String("audiofile", "a.aac", "export raw audio data to the audiofile")

func main() {
	flag.Parse()
	f, err := os.Open(*mp4filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	vfile, err := os.OpenFile(*rawvideo, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer vfile.Close()
	afile, err := os.OpenFile(*rawaudio, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer afile.Close()
	demuxer := mp4.CreateMp4Demuxer(f)
	if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", infos)
	}
	mp4info := demuxer.GetMp4Info()
	fmt.Printf("%+v\n", mp4info)
	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
		if pkg.Cid == mp4.MP4_CODEC_H264 {
			vfile.Write(pkg.Data)
		} else if pkg.Cid == mp4.MP4_CODEC_AAC {
			afile.Write(pkg.Data)
		}
	}

}
