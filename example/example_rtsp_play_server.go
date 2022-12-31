package main

import (
    "fmt"
    "net"
    "os"
    "strings"
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
}

func NewRtspPlayServerSession(conn net.Conn) *RtspPlaySeverSession {
    return &RtspPlaySeverSession{
        c:            conn,
        startUdpPort: 20000,
        tracks:       make(map[string]*rtsp.RtspTrack),
        senders:      make(map[string]*RtspUdpSender),
    }
}

func (server *RtspPlaySeverSession) Start() {
    svr := rtsp.NewRtspServer(server)
    svr.SetOutput(func(b []byte) (err error) {
        _, err = server.c.Write(b)
        return
    })
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
    server.c.Close()
    return
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
        laddr, _ := net.ResolveUDPAddr("udp4", srcAddr.String())
        srcAddr2 := net.UDPAddr{IP: net.IPv4zero, Port: server.senders[track.TrackName].rtcpPort}
        laddr2, _ := net.ResolveUDPAddr("udp4", srcAddr2.String())
        dstAddr := net.UDPAddr{IP: net.ParseIP(server.remoteAddr), Port: server.senders[track.TrackName].remoteRtpPort}
        raddr, _ := net.ResolveUDPAddr("udp4", dstAddr.String())
        dstAddr2 := net.UDPAddr{IP: net.ParseIP(server.remoteAddr), Port: server.senders[track.TrackName].remoteRtcpPort}
        raddr2, _ := net.ResolveUDPAddr("udp4", dstAddr2.String())
        server.senders[track.TrackName].rtp, _ = net.DialUDP("udp4", laddr, raddr)
        server.senders[track.TrackName].rtcp, _ = net.DialUDP("udp4", laddr2, raddr2)
        track.OpenTrack()
        track.OnPacket(func(b []byte, isRtcp bool) (err error) {
            if isRtcp {
                _, err = server.senders[track.TrackName].rtcp.Write(b)
            } else {
                _, err = server.senders[track.TrackName].rtp.Write(b)
            }
            return
        })
        server.senders[track.TrackName].track = track
        fmt.Println(server.senders[track.TrackName])
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
    }()

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
