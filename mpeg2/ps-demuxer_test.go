package mpeg2

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/yapingcat/gomedia/mpeg"
)

func Test_Input(t *testing.T) {
	psfile := "../example/source.200kbps.768x320.flv.ps"
	rfd, _ := os.Open(psfile)
	defer rfd.Close()
	buf, _ := ioutil.ReadAll(rfd)
	fmt.Printf("read %d size\n", len(buf))
	fd, err := os.OpenFile("1.h264", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fd.Close()
	fd2, err := os.OpenFile("4.aac", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fd.Close()
	demuxer := NewPSDemuxer()
	demuxer.OnFrame = func(frame []byte, cid PS_STREAM_TYPE, pts uint64, dts uint64) {
		if cid == PS_STREAM_H264 {
			if mpeg.H264NaluType(frame) == 9 {
				return
			}
			//fmt.Println(len(frame))
			n, err := fd.Write(frame)
			if err != nil || n != len(frame) {
				fmt.Println(err)
			}
		} else if cid == PS_STREAM_AAC {
			n, err := fd2.Write(frame)
			if err != nil || n != len(frame) {
				fmt.Println(err)
			}
		}
	}
	fmt.Println(demuxer.Input(buf))
}
