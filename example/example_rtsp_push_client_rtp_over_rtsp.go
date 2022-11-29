package main

import (
	"context"
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

type RtspRecordSession struct {
	once          sync.Once
	c             net.Conn
	handShakeDone chan struct{}
}

func NewRtspRecordSession(c net.Conn) *RtspRecordSession {
	return &RtspRecordSession{c: c, handShakeDone: make(chan struct{})}
}

func (cli *RtspRecordSession) Destory() {
	cli.once.Do(func() {
		cli.c.Close()
	})
}

func (cli *RtspRecordSession) HandleOption(res rtsp.RtspResponse, public []string) error {
	fmt.Println("rtsp server public ", public)
	return nil
}

func (cli *RtspRecordSession) HandleDescribe(res rtsp.RtspResponse, sdp *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
	return nil
}

func (cli *RtspRecordSession) HandleSetup(res rtsp.RtspResponse, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
	fmt.Println("HandleSetup sessionid:", sessionId, " timeout:", timeout)
	return nil
}

func (cli *RtspRecordSession) HandleAnnounce(res rtsp.RtspResponse) error {
	fmt.Println("Handle Announce", res.StatusCode)
	return nil
}

func (cli *RtspRecordSession) HandlePlay(res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
	return nil
}

func (cli *RtspRecordSession) HandlePause(res rtsp.RtspResponse) error {
	return nil
}

func (cli *RtspRecordSession) HandleTeardown(res rtsp.RtspResponse) error {
	return nil
}

func (cli *RtspRecordSession) HandleGetParameter(res rtsp.RtspResponse) error {
	return nil
}

func (cli *RtspRecordSession) HandleSetParameter(res rtsp.RtspResponse) error {
	return nil
}

func (cli *RtspRecordSession) HandleRedirect(req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
	return nil
}

func (cli *RtspRecordSession) HandleRecord(res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
	fmt.Println("hand Record ", res.StatusCode)
	close(cli.handShakeDone)
	return nil
}

func (cli *RtspRecordSession) HandleRequest(req rtsp.RtspRequest) error {
	return nil
}

func loopSend(ctx context.Context, channel chan []byte, c net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-channel:
			_, sendError = c.Write(b)
			if sendError != nil {
				return
			}
		}
	}
}

func readFlv(ctx context.Context, fileName string, done chan struct{}, track *rtsp.RtspTrack) {
	select {
	case <-done:
		break
	case <-ctx.Done():
		return
	}
	flvfilereader, _ := os.Open(fileName)
	defer flvfilereader.Close()
	fr := flv.CreateFlvReader()
	fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
		if ci == codec.CODECID_VIDEO_H264 {
			err := track.WriteSample(rtsp.RtspSample{Sample: b, Timestamp: pts})
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

	sess := NewRtspRecordSession(c)
	client, _ := rtsp.NewRtspClient(os.Args[1], sess, rtsp.WithEnableRecord())
	videoTrack := rtsp.NewVideoTrack(rtsp.RtspCodec{Cid: rtsp.RTSP_CODEC_H264, PayloadType: 96, SampleRate: 90000})
	client.AddTrack(videoTrack)
	sc := make(chan []byte, 100)
	ctx, cancel := context.WithCancel(context.Background())
	client.SetOutput(func(b []byte) error {
		if sendError != nil {
			return sendError
		}
		sc <- b
		return nil
	})
	go loopSend(ctx, sc, c)
	go readFlv(ctx, os.Args[2], sess.handShakeDone, videoTrack)
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

	cancel()
	sess.Destory()
}
