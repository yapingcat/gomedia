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
		tsf.Write(pkg)
	}

	pid := muxer.AddStream(mpeg2.TS_STREAM_AUDIO_MPEG1)
	mp3, _ := ioutil.ReadAll(f)
	var pts uint64 = 0
	var dts uint64 = 0
	codec.SplitMp3Frames(mp3, func(head *codec.MP3FrameHead, frame []byte) {
		sampleSize := head.SampleSize
		sampleRate := head.GetSampleRate()
		delta := sampleSize * 1000 / sampleRate
		muxer.Write(pid, frame, pts, dts)
		fmt.Println("write pts:", pts, "dts:", dts)
		pts += uint64(delta)
		dts = pts
	})

}
