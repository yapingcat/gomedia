package main

import (
    "fmt"
    "net"
    "net/url"
    "os"
    "sync"
    "time"

    "github.com/yapingcat/gomedia/go-codec"
    "github.com/yapingcat/gomedia/go-flv"
    "github.com/yapingcat/gomedia/go-rtsp"
    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

var sendError error
var flvFileName string

type RtspRecordSession struct {
    once       sync.Once
    c          net.Conn
    die        chan struct{}
    eof        chan struct{}
    waitSend   sync.WaitGroup
    sendChanel chan []byte
}

func NewRtspRecordSession(c net.Conn) *RtspRecordSession {
    return &RtspRecordSession{c: c, die: make(chan struct{}), eof: make(chan struct{}), sendChanel: make(chan []byte, 100)}
}

func (cli *RtspRecordSession) Destory() {
    cli.once.Do(func() {
        close(cli.die)
        cli.waitSend.Wait()
        for b := range cli.sendChanel {
            _, err := cli.c.Write(b)
            if err != nil {
                break
            }
        }
        cli.c.Close()
    })
}

func (cli *RtspRecordSession) HandleOption(client *rtsp.RtspClient, res rtsp.RtspResponse, public []string) error {
    fmt.Println("rtsp server public ", public)
    return nil
}

func (cli *RtspRecordSession) HandleDescribe(client *rtsp.RtspClient, res rtsp.RtspResponse, sdp *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
    return nil
}

func (cli *RtspRecordSession) HandleSetup(client *rtsp.RtspClient, res rtsp.RtspResponse, track *rtsp.RtspTrack, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
    fmt.Println("HandleSetup sessionid:", sessionId, " timeout:", timeout)
    return nil
}

func (cli *RtspRecordSession) HandleAnnounce(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    fmt.Println("Handle Announce", res.StatusCode)
    return nil
}

func (cli *RtspRecordSession) HandlePlay(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    return nil
}

func (cli *RtspRecordSession) HandlePause(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspRecordSession) HandleTeardown(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspRecordSession) HandleGetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspRecordSession) HandleSetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspRecordSession) HandleRedirect(client *rtsp.RtspClient, req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
    return nil
}

func (cli *RtspRecordSession) HandleRecord(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    fmt.Println("hand Record ", res.StatusCode)
    videoTrack, _ := client.GetTrack("video")
    go func() {
        flvfilereader, _ := os.Open(flvFileName)
        defer flvfilereader.Close()
        fr := flv.CreateFlvReader()
        fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
            if ci == codec.CODECID_VIDEO_H264 {
                err := videoTrack.WriteSample(rtsp.RtspSample{Sample: b, Timestamp: pts * 90})
                if err != nil {
                    fmt.Println(err)
                }
                time.Sleep(time.Millisecond * 20)
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
        cli.Destory()
    }()

    //send rtcp sr packet
    //it's not necessary in sometimes
    go func() {
        rtcpTimer := time.NewTicker(time.Duration(time.Second * 3))
        defer rtcpTimer.Stop()
        for {
            select {
            case <-rtcpTimer.C:
                videoTrack.SendReport()
            case <-cli.die:
                return
            }
        }
    }()
    return nil
}

func (cli *RtspRecordSession) HandleRequest(client *rtsp.RtspClient, req rtsp.RtspRequest) error {
    return nil
}

func (cli *RtspRecordSession) loopSend() {
    cli.waitSend.Add(1)
    defer cli.waitSend.Done()
    for {
        select {
        case <-cli.die:
            return
        case b := <-cli.sendChanel:
            _, sendError = cli.c.Write(b)
            if sendError != nil {
                return
            }
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
    flvFileName = os.Args[2]

    c, err := net.Dial("tcp4", host)
    if err != nil {
        fmt.Println(err)
        return
    }

    sess := NewRtspRecordSession(c)
    client, _ := rtsp.NewRtspClient(os.Args[1], sess, rtsp.WithEnableRecord())
    videoTrack := rtsp.NewVideoTrack(rtsp.RtspCodec{Cid: rtsp.RTSP_CODEC_H264, PayloadType: 96, SampleRate: 90000})
    client.AddTrack(videoTrack)
    client.SetOutput(func(b []byte) error {
        if sendError != nil {
            return sendError
        }
        sess.sendChanel <- b
        return nil
    })

    go sess.loopSend()
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
