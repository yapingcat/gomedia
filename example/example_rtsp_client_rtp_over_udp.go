package main

import (
    "errors"
    "fmt"
    "net"
    "net/url"
    "os"
    "sync"
    "time"

    "github.com/yapingcat/gomedia/go-rtsp"
    "github.com/yapingcat/gomedia/go-rtsp/sdp"
)

type UdpPairSession struct {
    rtpSess  *net.UDPConn
    rtcpSess *net.UDPConn
}

type RtspUdpPlaySession struct {
    udpport    uint16
    videoFile  *os.File
    audioFile  *os.File
    timeout    int
    once       sync.Once
    die        chan struct{}
    c          net.Conn
    lastError  error
    sesss      map[string]*UdpPairSession
    remoteAddr string
}

func NewRtspUdpPlaySession(c net.Conn) *RtspUdpPlaySession {
    return &RtspUdpPlaySession{udpport: 30000, die: make(chan struct{}), c: c, sesss: make(map[string]*UdpPairSession)}
}

func (cli *RtspUdpPlaySession) Destory() {
    cli.once.Do(func() {
        if cli.videoFile != nil {
            cli.videoFile.Close()
        }
        if cli.audioFile != nil {
            cli.audioFile.Close()
        }
        cli.c.Close()
        close(cli.die)
    })
}

func (cli *RtspUdpPlaySession) HandleOption(client *rtsp.RtspClient, res rtsp.RtspResponse, public []string) error {
    fmt.Println("rtsp server public ", public)
    return nil
}

func (cli *RtspUdpPlaySession) HandleDescribe(client *rtsp.RtspClient, res rtsp.RtspResponse, sdp *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
    fmt.Println("handle describe ", res.StatusCode, res.Reason)
    for k, t := range tracks {
        if t == nil {
            continue
        }
        fmt.Println("Got ", k, " track")
        transport := rtsp.NewRtspTransport(rtsp.WithEnableUdp(), rtsp.WithClientUdpPort(cli.udpport, cli.udpport+1), rtsp.WithMode(rtsp.MODE_PLAY))
        t.SetTransport(transport)
        t.OpenTrack()
        cli.udpport += 2
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

func (cli *RtspUdpPlaySession) HandleSetup(client *rtsp.RtspClient, res rtsp.RtspResponse, track *rtsp.RtspTrack, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
    fmt.Println("HandleSetup sessionid:", sessionId, " timeout:", timeout)
    if res.StatusCode == rtsp.Unsupported_Transport {
        return errors.New("unsupport udp transport")
    }
    srcAddr := net.UDPAddr{IP: net.IPv4zero, Port: int(track.GetTransport().Client_ports[0])}
    laddr, _ := net.ResolveUDPAddr("udp4", srcAddr.String())
    srcAddr2 := net.UDPAddr{IP: net.IPv4zero, Port: int(track.GetTransport().Client_ports[1])}
    laddr2, _ := net.ResolveUDPAddr("udp4", srcAddr2.String())
    dstAddr := net.UDPAddr{IP: net.ParseIP(cli.c.RemoteAddr().String()), Port: int(track.GetTransport().Server_ports[0])}
    raddr, _ := net.ResolveUDPAddr("udp4", dstAddr.String())
    dstAddr2 := net.UDPAddr{IP: net.ParseIP(cli.c.RemoteAddr().String()), Port: int(track.GetTransport().Server_ports[1])}
    raddr2, _ := net.ResolveUDPAddr("udp4", dstAddr2.String())
    rtpUdpsess, _ := net.DialUDP("udp4", laddr, raddr)
    rtcpUdpsess, _ := net.DialUDP("udp4", laddr2, raddr2)
    cli.sesss[track.TrackName] = &UdpPairSession{rtpSess: rtpUdpsess, rtcpSess: rtcpUdpsess}
    track.OnPacket(func(b []byte, isRtcp bool) (err error) {
        if isRtcp {
            _, err = rtcpUdpsess.Write(b)
        }
        return
    })
    go func() {
        buf := make([]byte, 1500)
        for {
            r, err := rtpUdpsess.Read(buf)
            if err != nil {
                fmt.Println(err)
                break
            }
            err = track.Input(buf[:r], false)
            if err != nil {
                fmt.Println(err)
                break
            }
        }
        cli.Destory()
    }()

    go func() {
        buf := make([]byte, 1500)
        for {
            r, err := rtcpUdpsess.Read(buf)
            if err != nil {
                fmt.Println(err)
                break
            }
            fmt.Println("read rtcp")
            err = track.Input(buf[:r], true)
            if err != nil {
                fmt.Println(err)
                break
            }
        }
        cli.Destory()
    }()

    cli.timeout = timeout
    return nil
}

func (cli *RtspUdpPlaySession) HandleAnnounce(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandlePlay(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
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

func (cli *RtspUdpPlaySession) HandlePause(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleTeardown(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleGetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleSetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleRedirect(client *rtsp.RtspClient, req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleRecord(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
    return nil
}

func (cli *RtspUdpPlaySession) HandleRequest(client *rtsp.RtspClient, req rtsp.RtspRequest) error {
    return nil
}

func (cli *RtspUdpPlaySession) sendInLoop(sendChan chan []byte) {
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
    sess := NewRtspUdpPlaySession(c)
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
