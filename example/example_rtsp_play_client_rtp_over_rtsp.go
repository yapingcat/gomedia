package main

import (
    "fmt"
    "net"
    "net/url"
    "os"
    "sync"
    "time"

    "github.com/yapingcat/gomedia/go-rtsp"
    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

const (
    Init      = 0
    HandShake = 1
    Playing   = 2
    Teardown  = 3
)

type RtspPlaySession struct {
    videoFile *os.File
    audioFile *os.File
    timeout   int
    state     int
    once      sync.Once
    c         net.Conn
}

func NewRtspPlaySession(c net.Conn) *RtspPlaySession {
    return &RtspPlaySession{state: Init, c: c}
}

func (cli *RtspPlaySession) Destory() {
    cli.once.Do(func() {
        if cli.videoFile != nil {
            cli.videoFile.Close()
        }
        if cli.audioFile != nil {
            cli.audioFile.Close()
        }
        cli.c.Close()
    })
}

func (cli *RtspPlaySession) HandleOption(res rtsp.RtspResponse, public []string) error {
    fmt.Println("rtsp server public ", public)
    return nil
}

func (cli *RtspPlaySession) HandleDescribe(res rtsp.RtspResponse, sdp *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
    fmt.Println("handle describe ", res.StatusCode, res.Reason)
    cli.state = HandShake
    for k, t := range tracks {
        if t == nil {
            continue
        }
        fmt.Println("Got ", k, " track")
        if t.Codec.Cid == rtsp.RTSP_CODEC_H264 {
            if cli.videoFile == nil {
                cli.videoFile, _ = os.OpenFile("video.h264", os.O_CREATE|os.O_RDWR, 0666)
            }
            t.OnSample(func(sample rtsp.RtspSample) {
                fmt.Println("Got H264 Frame size:", len(sample.Sample), " timestamp:", sample.Timestamp)
                cli.videoFile.Write(sample.Sample)
            })
        } else if t.Codec.Cid == rtsp.RTSP_CODEC_AAC {
            if cli.audioFile == nil {
                cli.audioFile, _ = os.OpenFile("audio.aac", os.O_CREATE|os.O_RDWR, 0666)
            }
            t.OnSample(func(sample rtsp.RtspSample) {
                fmt.Println("Got AAC Frame size:", len(sample.Sample), " timestamp:", sample.Timestamp)
                cli.audioFile.Write(sample.Sample)
            })
        }
    }
    return nil
}

func (cli *RtspPlaySession) HandleSetup(res rtsp.RtspResponse, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
    fmt.Println("HandleSetup sessionid:", sessionId, " timeout:", timeout)
    cli.timeout = timeout
    return nil
}

func (cli *RtspPlaySession) HandleAnnounce(res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandlePlay(res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    if res.StatusCode != 200 {
        fmt.Println("play failed ", res.StatusCode, res.Reason)
        return nil
    }
    cli.state = Playing
    return nil
}

func (cli *RtspPlaySession) HandlePause(res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleTeardown(res rtsp.RtspResponse) error {
    cli.state = Teardown
    return nil
}

func (cli *RtspPlaySession) HandleGetParameter(res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleSetParameter(res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleRedirect(req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
    return nil
}

func (cli *RtspPlaySession) HandleRecord(res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    return nil
}

func (cli *RtspPlaySession) HandleRequest(req rtsp.RtspRequest) error {
    return nil
}

func sendInLoop(sendChan chan []byte, quit chan struct{}, c net.Conn) {
    for {
        select {
        case b := <-sendChan:
            _, err := c.Write(b)
            if err != nil {
                c.Close()
                fmt.Println("quit send in loop")
                return
            }
        case <-quit:
            fmt.Println("quit send in loop")
            return
        }
    }
}

func main() {
    u, err := url.Parse(os.Args[1])
    if err != nil {
        panic(err)
    }
    host := u.Host
    if u.Port() == "" {
        host += ":554"
    }
    c, err := net.Dial("tcp4", host)
    if err != nil {
        fmt.Println(err)
        return
    }
    sess := NewRtspPlaySession(c)
    client, _ := rtsp.NewRtspClient(os.Args[1], sess)
    client.SetOutput(func(b []byte) error {
        c.Write(b)
        return nil
    })
    client.Start()
    buf := make([]byte, 4096)
    beg := time.Now()
    for {
        n, err := c.Read(buf)
        if err != nil {
            fmt.Println(err)
            break
        }
        if err = client.Input(buf[:n]); err != nil {
            fmt.Println(err)
            break
        }

        //rtsp keepalive
        switch sess.state {
        case Playing:
            if time.Now().After(beg.Add(time.Duration(sess.timeout/2) * time.Second)) {
                beg = time.Now()
                client.KeepAlive(rtsp.OPTIONS)
            }
        case Teardown:
            break
        default:
            beg = time.Now()
        }
    }
    sess.Destory()
}
