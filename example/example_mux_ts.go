package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

func main() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	tsfilename := os.Args[2]
	tsf, err := os.OpenFile(tsfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tsf.Close()

	muxer := mpeg2.NewTSMuxer()
	muxer.OnPacket = func(pkg []byte) {
		fmt.Println("write packet")
		//mpeg.ShowPacketHexdump(pkg)
		tsf.Write(pkg)
		//os.Stdin.Read(make([]byte, 1))
	}

	pid := muxer.AddStream(mpeg2.TS_STREAM_H264)
	h264, _ := ioutil.ReadAll(f)
	var pts uint64 = 0
	var dts uint64 = 0
	codec.SplitFrameWithStartCode(h264, func(nalu []byte) bool {
		//fmt.Println("wtite nalu")
		if codec.H264NaluType(nalu) <= codec.H264_NAL_I_SLICE {
			pts += 40
			dts += 40
			fmt.Println(dts)
		}
		fmt.Println(muxer.Write(pid, nalu, pts, dts))
		return true
	})

}
