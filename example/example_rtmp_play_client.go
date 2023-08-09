package main

import (
    "flag"
    "fmt"
    "net"
    "net/url"
    "os"

    "github.com/yapingcat/gomedia/go-codec"
    "github.com/yapingcat/gomedia/go-rtmp"
)

var rtmpUrl = flag.String("url", "rtmp://127.0.0.1/live/test", "play rtmp url")
var video = flag.String("video", "v.h264", "safe video data to the file")
var audio = flag.String("audio", "a.aac", "safe audio data to the file")

func main() {
    flag.Parse()

    u, err := url.Parse(*rtmpUrl)
    if err != nil {
        panic(err)
    }
    host := u.Host
    if u.Port() == "" {
        host += ":1935"
    }

    //connect to remote rtmp server
    c, err := net.Dial("tcp4", host)
    if err != nil {
        fmt.Println(err)
        return
    }

    videoFd, _ := os.OpenFile(*video, os.O_CREATE|os.O_RDWR, 0666)
    audioFd, _ := os.OpenFile(*audio, os.O_CREATE|os.O_RDWR, 0666)
    defer func() {
        if c != nil {
            c.Close()
        }
        if videoFd != nil {
            videoFd.Close()
        }
        if audioFd != nil {
            audioFd.Close()
        }
    }()

    client := rtmp.NewRtmpClient(rtmp.WithChunkSize(6000), rtmp.WithComplexHandshake())
    client.OnFrame(func(cid codec.CodecID, pts, dts uint32, frame []byte) {
        if cid == codec.CODECID_VIDEO_H264 || cid == codec.CODECID_VIDEO_H265 {
            videoFd.Write(frame)
        } else if cid == codec.CODECID_AUDIO_AAC {
            audioFd.Write(frame)
        }
    })

    //must set output callback
    client.SetOutput(func(b []byte) error {
        _, err := c.Write(b)
        return err
    })

    client.Start(*rtmpUrl)
    buf := make([]byte, 4096)
    n := 0
    for {
        n, err = c.Read(buf)
        if err != nil {
            break
        }
        err = client.Input(buf[:n])
        if err != nil {
            break
        }
    }
    fmt.Println(err)
}
