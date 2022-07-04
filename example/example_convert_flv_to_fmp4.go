package main

import (
    "fmt"
    "os"

    "github.com/yapingcat/gomedia/codec"
    "github.com/yapingcat/gomedia/flv"
    "github.com/yapingcat/gomedia/mp4"
)

func main() {
    mp4filename := "test2_fmp4.mp4"
    mp4file, err := os.OpenFile(mp4filename, os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer mp4file.Close()

    muxer, err := mp4.CreateMp4Muxer(mp4file, mp4.WithMp4Flag(mp4.MP4_FLAG_FRAGMENT))
    if err != nil {
        fmt.Println(err)
        return
    }
    vtid := muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
    atid := muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)

    flvfilereader, _ := os.Open(os.Args[1])
    defer flvfilereader.Close()
    fr := flv.CreateFlvReader()

    fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
        if ci == codec.CODECID_AUDIO_AAC {
            err := muxer.Write(atid, b, uint64(pts), uint64(dts))
            if err != nil {
                fmt.Println(err)
            }
        } else if ci == codec.CODECID_VIDEO_H264 {
            err := muxer.Write(vtid, b, uint64(pts), uint64(dts))
            if err != nil {
                fmt.Println(err)
            }
        }
    }

    cache := make([]byte, 4096)
    for {
        n, err := flvfilereader.Read(cache)
        if err != nil {
            fmt.Println(err)
            break
        }
        fr.Input(cache[0:n])
    }
    muxer.WriteTrailer()
}
