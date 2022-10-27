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
		fmt.Println("write aac packet")
		tsf.Write(pkg)
	}

	pid := muxer.AddStream(mpeg2.TS_STREAM_AAC)
	aac, _ := ioutil.ReadAll(f)
	var pts uint64 = 0
	var dts uint64 = 0
	var i int = 0
	codec.SplitAACFrame(aac, func(aac []byte) {

		if i < 3 {
			pts += 23
			dts += 23
			i++
		} else {
			pts += 24
			dts += 24
			i = 0
		}
		fmt.Println(muxer.Write(pid, aac, pts, dts))
	})

}
