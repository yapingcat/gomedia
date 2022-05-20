package rtmp

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/yapingcat/gomedia/codec"
	"github.com/yapingcat/gomedia/flv"
)

func TestRtmpServerHandle_Play(t *testing.T) {

	t.Run("server play", func(t *testing.T) {
		listen, _ := net.Listen("tcp4", "0.0.0.0:1935")
		conn, _ := listen.Accept()

		go func() {

			ready := make(chan struct{})

			handle := NewRtmpServerHandle()
			handle.onPlay = func(app, streamName string, start, duration float64, reset bool) StatusCode {
				fmt.Println("onplay", app, streamName, start, duration)
				return NETSTREAM_PLAY_START
			}

			handle.OnStateChange(func(newstate RtmpState) {
				if newstate == STATE_RTMP_PLAY_START {
					close(ready)
				}
			})

			handle.SetOutput(func(b []byte) {
				conn.Write(b)
			})

			go func() {
				<-ready
				f := flv.CreateFlvReader()
				f.OnFrame = func(cid codec.CodecID, frame []byte, pts, dts uint32) {
					if cid == codec.CODECID_VIDEO_H264 {
						fmt.Println("write video frame", pts, dts)
						handle.WriteVideo(cid, frame, pts, dts)
						time.Sleep(time.Millisecond * 33)
					} else if cid == codec.CODECID_AUDIO_AAC {
						handle.WriteAudio(cid, frame, pts, dts)
					}
				}
				fd, _ := os.Open("source.200kbps.768x320.flv")
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
			}()

			buf := make([]byte, 60000)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					fmt.Println(err)
					break
				}
				err = handle.Input(buf[0:n])
				if err != nil {
					fmt.Println(err)
					break
				}
			}
			conn.Close()
		}()

		select {}
	})
}

func TestRtmpServerHandle_Pub(t *testing.T) {

	t.Run("server pub", func(t *testing.T) {
		listen, _ := net.Listen("tcp4", "0.0.0.0:1935")
		conn, _ := listen.Accept()

		go func() {

			var videoFile *os.File
			handle := NewRtmpServerHandle()
			handle.OnPublish(func(app, streamName string) StatusCode {
				videoFile, _ = os.OpenFile(streamName+".h264", os.O_CREATE|os.O_RDWR, 0666)
				return NETSTREAM_PUBLISH_START
			})

			handle.SetOutput(func(b []byte) {
				conn.Write(b)
			})

			handle.OnFrame(func(cid codec.CodecID, pts, dts uint32, frame []byte) {
				if cid == codec.CODECID_VIDEO_H264 {
					fmt.Println("H264, length:", len(frame), "pts:", pts, "dts:", dts)
					videoFile.Write(frame)
				}
			})

			buf := make([]byte, 60000)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					fmt.Println(err)
					break
				}
				err = handle.Input(buf[0:n])
				if err != nil {
					fmt.Println(err)
					break
				}
			}
			conn.Close()
		}()

		select {}
	})
}
