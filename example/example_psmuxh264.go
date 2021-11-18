package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/mpeg"
	"github.com/yapingcat/gomedia/mpeg2"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	psfilename := os.Args[1] + ".ps"

	ps, err := os.OpenFile(psfilename, os.O_CREATE|os.O_RDWR, 666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ps.Close()

	muxer := mpeg2.NewPsMuxer()
	muxer.OnPacket = func(pkg []byte) {
		ps.Write(pkg)
	}
	pid := muxer.AddStream(mpeg2.PS_STREAM_H264)
	buf, _ := ioutil.ReadAll(f)
	pts := uint64(0)
	dts := uint64(0)
	mpeg.SplitFrameWithStartCode(buf, func(nalu []byte) bool {
		muxer.Write(pid, nalu, pts*90, dts*90)
		if mpeg.H264NaluType(nalu) <= mpeg.H264_NAL_I_SLICE {
			pts += 40
			dts += 40
		}
		return true
	})

}
