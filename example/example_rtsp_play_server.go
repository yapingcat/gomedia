package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/yapingcat/gomedia/go-codec"
    "github.com/yapingcat/gomedia/go-flv"
    "github.com/yapingcat/gomedia/go-rtsp"
)

type RtspUdpSender struct {
    rtpPort        int
    rtcpPort       int
    remoteRtpPort  int
    remoteRtcpPort int
    rtp            *net.UDPConn
    rtcp           *net.UDPConn
    track          *rtsp.RtspTrack
}

type RtspPlaySeverSession struct {
    startUdpPort int
    c            net.Conn
    remoteAddr   string
    tracks       map[string]*rtsp.RtspTrack
    senders      map[string]*RtspUdpSender
    die          sync.Once
    completed    chan struct{}
}

func NewRtspPlayServerSession(conn net.Conn) *RtspPlaySeverSession {
    return &RtspPlaySeverSession{
        c:            conn,
        startUdpPort: 20000,
        tracks:       make(map[string]*rtsp.RtspTrack),
        senders:      make(map[string]*RtspUdpSender),
        completed:    make(chan struct{}),
    }
}

func (server *RtspPlaySeverSession) Start() {
    svr := rtsp.NewRtspServer(server)
    svr.SetOutput(func(b []byte) (err error) {
        _, err = server.c.Write(b)
        return
    })
    defer server.Stop()
    server.remoteAddr = server.c.RemoteAddr().String()
    buf := make([]byte, 65535)
    for {
        n, err := server.c.Read(buf)
        if err != nil {
            fmt.Println(err)
            break
        }
        svr.Input(buf[:n])
    }
    return
}

func (server *RtspPlaySeverSession) Stop() {
    server.die.Do(func() {
        server.c.Close()
        close(server.completed)
    })
}

func (server *RtspPlaySeverSession) HandleOption(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {
    fmt.Println("handle option")
}

func (server *RtspPlaySeverSession) HandleDescribe(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {
    fmt.Println("handle describe")
    fmt.Println("add video track")
    videoTrack := rtsp.NewVideoTrack(rtsp.RtspCodec{Cid: rtsp.RTSP_CODEC_H264, PayloadType: 96, SampleRate: 90000})
    svr.AddTrack(videoTrack)
    server.tracks["video"] = videoTrack
    server.senders["video"] = &RtspUdpSender{rtpPort: server.startUdpPort, rtcpPort: server.startUdpPort + 1}
    server.startUdpPort += 2
}

func (server *RtspPlaySeverSession) HandleSetup(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse, transport *rtsp.RtspTransport, track *rtsp.RtspTrack) {
    fmt.Println("handle setup", *transport)
    if transport.Proto == rtsp.UDP {
        transport.Server_ports[0] = uint16(server.senders[track.TrackName].rtpPort)
        transport.Server_ports[1] = uint16(server.senders[track.TrackName].rtcpPort)
        server.senders[track.TrackName].remoteRtpPort = int(transport.Client_ports[0])
        server.senders[track.TrackName].remoteRtcpPort = int(transport.Client_ports[1])

        srcAddr := net.UDPAddr{IP: net.IPv4zero, Port: server.senders[track.TrackName].rtpPort}
        srcAddr2 := net.UDPAddr{IP: net.IPv4zero, Port: server.senders[track.TrackName].rtcpPort}
        dstAddr := net.UDPAddr{IP: net.ParseIP(server.remoteAddr), Port: server.senders[track.TrackName].remoteRtpPort}
        dstAddr2 := net.UDPAddr{IP: net.ParseIP(server.remoteAddr), Port: server.senders[track.TrackName].remoteRtcpPort}
        server.senders[track.TrackName].rtp, _ = net.DialUDP("udp4", &srcAddr, &dstAddr)
        server.senders[track.TrackName].rtcp, _ = net.DialUDP("udp4", &srcAddr2, &dstAddr2)
        track.OpenTrack()
        track.OnPacket(func(b []byte, isRtcp bool) (err error) {
            if isRtcp {
                fmt.Println("send rtcp packet")
                _, err = server.senders[track.TrackName].rtcp.Write(b)
            } else {
                _, err = server.senders[track.TrackName].rtp.Write(b)
            }
            return
        })
        server.senders[track.TrackName].track = track
        return
    } else {
        res.StatusCode = rtsp.Unsupported_Transport
    }
}

func (server *RtspPlaySeverSession) HandlePlay(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse, timeRange *rtsp.RangeTime, info []*rtsp.RtpInfo) {
    fmt.Println("handle play")
    streamName := req.Uri[strings.LastIndex(req.Uri, "/")+1:]
    fileName := streamName + ".flv"
    go func() {
        flvfilereader, _ := os.Open(fileName)
        defer flvfilereader.Close()
        fr := flv.CreateFlvReader()
        fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
            if ci == codec.CODECID_VIDEO_H264 {
                err := server.senders["video"].track.WriteSample(rtsp.RtspSample{Sample: b, Timestamp: pts * 90})
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
            err = fr.Input(cache[0:n])
            if err != nil {
                break
            }
        }
        server.Stop()
    }()

    go func() {
        rtcpTimer := time.NewTicker(time.Duration(time.Second * 3))
        defer rtcpTimer.Stop()
        for {
            select {
            case <-rtcpTimer.C:
                for _, sender := range server.senders {
                    err := sender.track.SendReport()
                    fmt.Println("send report")
                    if err != nil {
                        fmt.Println(err)
                        return
                    }
                }
            case <-server.completed:
                return
            }
        }
    }()

    for _, sender := range server.senders {
        go func() {
            buf := make([]byte, 4096)
            for {
                n, err := sender.rtcp.Read(buf)
                if err != nil {
                    fmt.Println(err)
                    break
                }
                fmt.Println("read rtcp packet ", n)
                sender.track.Input(buf[:n], true)
            }
        }()
    }
}

func (server *RtspPlaySeverSession) HandleAnnounce(svr *rtsp.RtspServer, req rtsp.RtspRequest, tracks map[string]*rtsp.RtspTrack) {
}

func (server *RtspPlaySeverSession) HandlePause(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {

}

func (server *RtspPlaySeverSession) HandleTeardown(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {
}

func (server *RtspPlaySeverSession) HandleGetParameter(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {

}

func (server *RtspPlaySeverSession) HandleSetParameter(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse) {

}

func (server *RtspPlaySeverSession) HandleRecord(svr *rtsp.RtspServer, req rtsp.RtspRequest, res *rtsp.RtspResponse, timeRange *rtsp.RangeTime, info []*rtsp.RtpInfo) {

}

func (server *RtspPlaySeverSession) HandleResponse(svr *rtsp.RtspServer, res rtsp.RtspResponse) {

}

func main() {
    addr := "0.0.0.0:554"
    listen, _ := net.Listen("tcp4", addr)
    for {
        conn, _ := listen.Accept()
        sess := NewRtspPlayServerSession(conn)
        go sess.Start()
    }
}
