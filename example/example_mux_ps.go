package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "github.com/yapingcat/gomedia/mpeg2"
)

func main() {
    filename := os.Args[1]
    f2, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f2.Close()
    psfilename := os.Args[2]
    psf, err := os.OpenFile(psfilename, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer psf.Close()

    muxer := mpeg2.NewPsMuxer()
    muxer.OnPacket = func(pkg []byte) {
        psf.Write(pkg)
    }

    pid := muxer.AddStream(mpeg2.PS_STREAM_H265)
    audioId := muxer.AddStream(mpeg2.PS_STREAM_AAC)
    buf, _ := ioutil.ReadAll(f2)
    demuxer := mpeg2.NewPSDemuxer()
    demuxer.OnFrame = func(frame []byte, cid mpeg2.PS_STREAM_TYPE, pts uint64, dts uint64) {
        if cid == mpeg2.PS_STREAM_H265 {
            muxer.Write(pid, frame, pts, dts)
            // fmt.Println("write h264")
        } else if cid == mpeg2.PS_STREAM_AAC {
            muxer.Write(audioId, frame, pts, dts)
        }
    }
    demuxer.Input(buf)
    demuxer.Flush()

}
