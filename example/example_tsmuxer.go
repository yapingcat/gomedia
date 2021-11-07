package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"time"

	"../mpeg"
	"../mpeg2"
)

func main() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	tsfilename := os.Args[2]
	tsf, err := os.OpenFile(tsfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	muxer := mpeg2.NewTSMuxer()
	muxer.OnPacket = func(pkg []byte) {
		fmt.Println("write packet")
		//mpeg.ShowPacketHexdump(pkg)
		tsf.Write(pkg)
		//os.Stdin.Read(make([]byte, 1))
	}

	cpuout, err := os.OpenFile("cpu.out", os.O_RDWR|os.O_CREATE, 0666)
	pprof.StartCPUProfile(cpuout)

	go func() {
		timeout := time.NewTicker(time.Second * 30)
		c := timeout.C
		<-c
		pprof.StopCPUProfile()
	}()

	pid := muxer.AddStream(mpeg2.TS_STREAM_H264)
	h264, _ := ioutil.ReadAll(f)
	var pts uint64 = 0
	var dts uint64 = 0
	mpeg.SplitFrameWithStartCode(h264, func(nalu []byte) bool {
		//fmt.Println("wtite nalu")
		if mpeg.H264NaluType(nalu) <= mpeg.H264_NAL_I_SLICE {
			pts += 40
			dts += 40
			fmt.Println(dts)
		}
		fmt.Println(muxer.Write(pid, nalu, pts, dts))
		return true
	})

}
