package main

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "os"

    "github.com/yapingcat/gomedia/go-mp4"
    "github.com/yapingcat/gomedia/go-mpeg2"
)

func main() {
    tsfile := os.Args[1]
    tsFd, err := os.Open(tsfile)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer tsFd.Close()

    hasAudio := false
    hasVideo := false
    var atid uint32 = 0
    var vtid uint32 = 0

    mp4filename := "convert_ts.mp4"
    mp4file, err := os.OpenFile(mp4filename, os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer mp4file.Close()

    muxer, err := mp4.CreateMp4Muxer(mp4file)
    if err != nil {
        fmt.Println(err)
        return
    }

    demuxer := mpeg2.NewTSDemuxer()
    demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
        if cid == mpeg2.TS_STREAM_H264 {
            if !hasVideo {
                vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
                hasVideo = true
            }
            err := muxer.Write(vtid, frame, uint64(pts), uint64(dts))
            if err != nil {
                fmt.Println(err)
            }
        } else if cid == mpeg2.TS_STREAM_AAC {
            if !hasAudio {
                atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
                hasAudio = true
            }
            err := muxer.Write(atid, frame, uint64(pts), uint64(dts))
            if err != nil {
                fmt.Println(err)
            }
        } else if cid == mpeg2.TS_STREAM_AUDIO_MPEG1 || cid == mpeg2.TS_STREAM_AUDIO_MPEG2 {
            if !hasAudio {
                atid = muxer.AddAudioTrack(mp4.MP4_CODEC_MP3)
                hasAudio = true
            }
            err := muxer.Write(atid, frame, uint64(pts), uint64(dts))
            if err != nil {
                fmt.Println(err)
            }
        }
    }

    buf, _ := ioutil.ReadAll(tsFd)
    fmt.Printf("read %d size\n", len(buf))
    fmt.Println(demuxer.Input(bytes.NewReader(buf)))

    muxer.WriteTrailer()
}
