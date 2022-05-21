package main

import (
    "flag"
    "fmt"
    "net"
    "net/url"
    "os"
    "time"

    "github.com/yapingcat/gomedia/codec"
    "github.com/yapingcat/gomedia/flv"
    "github.com/yapingcat/gomedia/rtmp"
)

var rtmpUrl = flag.String("url", "rtmp://127.0.0.1/live/test", "publish rtmp url")
var flvFile = flag.String("flv", "test.flv", "push flv file to server")

func publish(fileName string, cli *rtmp.RtmpClient) {
    f := flv.CreateFlvReader()
    f.OnFrame = func(cid codec.CodecID, frame []byte, pts, dts uint32) {
        if cid == codec.CODECID_VIDEO_H264 {
            cli.WriteVideo(cid, frame, pts, dts)
            time.Sleep(time.Millisecond * 33)
        } else if cid == codec.CODECID_AUDIO_AAC {
            cli.WriteAudio(cid, frame, pts, dts)
        }
    }
    fd, _ := os.Open(fileName)
    defer fd.Close()
    cache := make([]byte, 4096)
    for {
        n, err := fd.Read(cache)
        if err != nil {
            fmt.Println(err)
            break
        }
        f.Input(cache[0:n])
    }
}

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
        fmt.Println("connect failed", err)
        return
    }

    isReady := make(chan struct{})

    //创建rtmp client 使能复杂握手,rtmp推流
    cli := rtmp.NewRtmpClient(rtmp.WithComplexHandshake(), rtmp.WithEnablePublish())

    //监听状态变化,STATE_RTMP_PUBLISH_START 状态通知推流
    cli.OnStateChange(func(newState rtmp.RtmpState) {
        if newState == rtmp.STATE_RTMP_PUBLISH_START {
            fmt.Println("ready for publish")
            close(isReady)
        }
    })

    cli.SetOutput(func(data []byte) error {
        _, err := c.Write(data)
        return err
    })

    go func() {
        <-isReady
        fmt.Println("start to read flv")
        publish(*flvFile, cli)
    }()

    cli.Start(*rtmpUrl)
    buf := make([]byte, 4096)
    n := 0
    for err == nil {
        n, err = c.Read(buf)
        if err != nil {
            continue
        }
        cli.Input(buf[:n])
    }
    fmt.Println(err)
}
