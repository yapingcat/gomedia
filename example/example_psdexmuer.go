package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/mpeg"
	"github.com/yapingcat/gomedia/mpeg2"
)

func main() {
	psfile := os.Args[1]
	rfd, _ := os.Open(psfile)
	defer rfd.Close()
	buf, _ := ioutil.ReadAll(rfd)
	fmt.Printf("read %d size\n", len(buf))
	fd, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fd.Close()
	fd2, err := os.OpenFile(os.Args[3], os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fd.Close()
	demuxer := mpeg2.NewPSDemuxer()
	demuxer.OnFrame = func(frame []byte, cid mpeg2.PS_STREAM_TYPE, pts uint64, dts uint64) {
		if cid == mpeg2.PS_STREAM_H264 {
			if mpeg.H264NaluType(frame) == 9 {
				return
			}
			//fmt.Printf("write h264 frame:%d\n", len(frame))
			n, err := fd.Write(frame)
			if err != nil || n != len(frame) {
				fmt.Println(err)
			}
		} else if cid == mpeg2.PS_STREAM_AAC {
			n, err := fd2.Write(frame)
			if err != nil || n != len(frame) {
				fmt.Println(err)
			}
		}
	}
	fmt.Println(demuxer.Input(buf))
	demuxer.Flush()
}
