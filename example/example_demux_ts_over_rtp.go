package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/yapingcat/gomedia/go-mpeg2"
)

const RTP_FIX_HEAD_LEN = 12

type RtpReceiver struct {
	conn  *net.UDPConn
	buf   []byte
	cache *bytes.Buffer
}

func NewRtpReceiver(conn *net.UDPConn) *RtpReceiver {
	return &RtpReceiver{
		conn:  conn,
		buf:   make([]byte, 4096),
		cache: bytes.NewBuffer(make([]byte, 0, 1500)),
	}
}

func (r *RtpReceiver) Read(p []byte) (n int, err error) {

	if r.cache.Len() > 0 {
		return r.cache.Read(p)
	}
	for {
		readBytes := 0
		err = r.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		if err != nil {
			return 0, err
		}
		readBytes, err = r.conn.Read(r.buf)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		if readBytes <= RTP_FIX_HEAD_LEN {
			fmt.Println("rtp payload length == 0")
			continue
		}
		//filter out rtp head
		if readBytes-RTP_FIX_HEAD_LEN <= len(p) {
			n = copy(p, r.buf[RTP_FIX_HEAD_LEN:readBytes])
			return
		}
		n = copy(p, r.buf[RTP_FIX_HEAD_LEN:RTP_FIX_HEAD_LEN+len(p)])
		r.cache.Write(r.buf[RTP_FIX_HEAD_LEN+len(p) : readBytes])
		return
	}
}

var videoFile = flag.String("videofile", "v.h264", "export raw video data to the videofile")
var audioFile = flag.String("audiofile", "a.aac", "export raw audio data to the audiofile")

//use ffmpeg commad to test this example
//ffmpeg -re -i <media file> -vcodec copy -acodec copy -f rtp_mpegts rtp://127.0.0.1:19999
func main() {
	flag.Parse()

	var v *os.File = nil
	var a *os.File = nil

	defer func() {
		if v != nil {
			v.Close()
		}

		if a != nil {
			a.Close()
		}

	}()

	localAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:19999")
	c, _ := net.ListenUDP("udp4", localAddr)
	demuxer := mpeg2.NewTSDemuxer()
	demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts, dts uint64) {
		if cid == mpeg2.TS_STREAM_H264 {
			if v == nil {
				v, _ = os.OpenFile(*videoFile, os.O_CREATE|os.O_RDWR, 0666)
			}
			fmt.Println("Got H264 Frame:", "pts:", pts, "dts:", dts, "Frame len:", len(frame))
			v.Write(frame)
		} else if cid == mpeg2.TS_STREAM_AAC {
			if a == nil {
				a, _ = os.OpenFile(*audioFile, os.O_CREATE|os.O_RDWR, 0666)
			}
			a.Write(frame)
		}
	}
	fmt.Println(demuxer.Input(NewRtpReceiver(c)))
}
