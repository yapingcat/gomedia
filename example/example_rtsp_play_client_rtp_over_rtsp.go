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
    tsFile *os.File
    timeout   int
    once      sync.Once
    die       chan struct{}
    c         net.Conn
    lastError error
}

func NewRtspPlaySession(c net.Conn) *RtspPlaySession {
    return &RtspPlaySession{die: make(chan struct{}), c: c}
}

func (cli *RtspPlaySession) Destory() {
    cli.once.Do(func() {
        if cli.videoFile != nil {
            cli.videoFile.Close()
        }
        if cli.audioFile != nil {
            cli.audioFile.Close()
        }
        if cli.tsFile != nil {
            cli.tsFile.Close()
        }
        cli.c.Close()
        close(cli.die)
    })
}

func (cli *RtspPlaySession) HandleOption(client *rtsp.RtspClient, res rtsp.RtspResponse, public []string) error {
    fmt.Println("rtsp server public ", public)
    return nil
}

func (cli *RtspPlaySession) HandleDescribe(client *rtsp.RtspClient, res rtsp.RtspResponse, sdp *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
    fmt.Println("handle describe ", res.StatusCode, res.Reason)
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
                //fmt.Println("Got H264 Frame size:", len(sample.Sample), " timestamp:", sample.Timestamp)
                cli.videoFile.Write(sample.Sample)
            })
        } else if t.Codec.Cid == rtsp.RTSP_CODEC_AAC {
            if cli.audioFile == nil {
                cli.audioFile, _ = os.OpenFile("audio.aac", os.O_CREATE|os.O_RDWR, 0666)
            }
            t.OnSample(func(sample rtsp.RtspSample) {
                //fmt.Println("Got AAC Frame size:", len(sample.Sample), " timestamp:", sample.Timestamp)
                cli.audioFile.Write(sample.Sample)
            })
        } else if t.Codec.Cid == rtsp.RTSP_CODEC_TS {
            if cli.tsFile == nil {
                cli.tsFile, _ = os.OpenFile("mp2t.ts", os.O_CREATE|os.O_RDWR, 0666)
            }
            t.OnSample(func(sample rtsp.RtspSample) {
                cli.tsFile.Write(sample.Sample)
            })
        }
    }
    return nil
}

func (cli *RtspPlaySession) HandleSetup(client *rtsp.RtspClient, res rtsp.RtspResponse, track *rtsp.RtspTrack, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
    fmt.Println("HandleSetup sessionid:", sessionId, " timeout:", timeout)
    cli.timeout = timeout
    return nil
}

func (cli *RtspPlaySession) HandleAnnounce(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandlePlay(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    if res.StatusCode != 200 {
        fmt.Println("play failed ", res.StatusCode, res.Reason)
        return nil
    }
    go func() {
        //rtsp keepalive
        to := time.NewTicker(time.Duration(cli.timeout/2) * time.Second)
        defer to.Stop()
        for {
            select {
            case <-to.C:
                client.KeepAlive(rtsp.OPTIONS)
            case <-cli.die:
                return
            }
        }
    }()
    return nil
}

func (cli *RtspPlaySession) HandlePause(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleTeardown(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleGetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleSetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspPlaySession) HandleRedirect(client *rtsp.RtspClient, req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
    return nil
}

func (cli *RtspPlaySession) HandleRecord(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    return nil
}

func (cli *RtspPlaySession) HandleRequest(client *rtsp.RtspClient, req rtsp.RtspRequest) error {
    return nil
}

func (cli *RtspPlaySession) sendInLoop(sendChan chan []byte) {
    for {
        select {
        case b := <-sendChan:
            _, err := cli.c.Write(b)
            if err != nil {
                cli.Destory()
                cli.lastError = err
                fmt.Println("quit send in loop")
                return
            }

        case <-cli.die:
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

    sc := make(chan []byte, 100)
    sess := NewRtspPlaySession(c)

    go sess.sendInLoop(sc)
    client, _ := rtsp.NewRtspClient(os.Args[1], sess)
    client.SetOutput(func(b []byte) error {
        if sess.lastError != nil {
            return sess.lastError
        }
        sc <- b
        return nil
    })
    client.Start()
    buf := make([]byte, 4096)
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
    }
    sess.Destory()
}
