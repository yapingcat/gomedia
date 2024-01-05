package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-rtmp"
)

type TimestampAdjust struct {
	lastTimeStamp    int64
	adjust_timestamp int64
}

func newTimestampAdjust() *TimestampAdjust {
	return &TimestampAdjust{
		lastTimeStamp:    -1,
		adjust_timestamp: 0,
	}
}

// timestamp in millisecond
func (adjust *TimestampAdjust) adjust(timestamp int64) int64 {
	if adjust.lastTimeStamp == -1 {
		adjust.adjust_timestamp = timestamp
		adjust.lastTimeStamp = timestamp
		return adjust.adjust_timestamp
	}

	delta := timestamp - adjust.lastTimeStamp
	if delta < -1000 || delta > 1000 {
		adjust.adjust_timestamp = adjust.adjust_timestamp + 1
	} else {
		adjust.adjust_timestamp = adjust.adjust_timestamp + delta
	}
	adjust.lastTimeStamp = timestamp
	return adjust.adjust_timestamp
}

var video_pts_adjust *TimestampAdjust = newTimestampAdjust()
var video_dts_adjust *TimestampAdjust = newTimestampAdjust()
var audio_ts_adjust *TimestampAdjust = newTimestampAdjust()

// Will push the last file under mp4sPath to the specified rtmp server
func main() {
	var (
		mp4Path = "your_mp4_dir" //like ./mp4/
		rtmpUrl = "rtmpUrl"      //like rtmp://127.0.0.1:1935/live/test110
	)
	c, err := net.Dial("tcp4", "127.0.0.1:1935")
	if err != nil {
		fmt.Println(err)
	}
	cli := rtmp.NewRtmpClient(rtmp.WithComplexHandshake(),
		rtmp.WithComplexHandshakeSchema(rtmp.HANDSHAKE_COMPLEX_SCHEMA0),
		rtmp.WithEnablePublish())
	cli.OnError(func(code, describe string) {
		fmt.Printf("rtmp code:%s ,describe:%s\n", code, describe)
	})
	isReady := make(chan struct{})
	cli.OnStatus(func(code, level, describe string) {
		fmt.Printf("rtmp onstatus code:%s ,level %s ,describe:%s\n", code, describe)
	})
	cli.OnStateChange(func(newState rtmp.RtmpState) {
		if newState == rtmp.STATE_RTMP_PUBLISH_START {
			fmt.Println("ready for publish")
			close(isReady)
		}
	})
	cli.SetOutput(func(bytes []byte) error {
		_, err := c.Write(bytes)
		return err
	})
	go func() {
		<-isReady
		fmt.Println("start to read file")
		for {
			filees, err := os.ReadDir(mp4Path)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(mp4Path + filees[len(filees)-1].Name())
			PushRtmp(mp4Path+filees[len(filees)-1].Name(), cli)
		}
	}()

	cli.Start(rtmpUrl)
	buf := make([]byte, 4096)
	n := 0
	for err == nil {
		n, err = c.Read(buf)
		if err != nil {
			continue
		}
		fmt.Println("read byte", n)
		cli.Input(buf[:n])
	}
	fmt.Println(err)
}

func PushRtmp(fileName string, cli *rtmp.RtmpClient) {
	mp4File, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mp4File.Close()
	demuxer := mp4.CreateMp4Demuxer(mp4File)
	if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", infos)
	}
	mp4info := demuxer.GetMp4Info()
	fmt.Printf("%+v\n", mp4info)

	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			fmt.Println(err)
			break
		}
		if pkg.Cid == mp4.MP4_CODEC_H264 {
			time.Sleep(20 * time.Millisecond)
			pts := video_pts_adjust.adjust(int64(pkg.Pts))
			dts := video_dts_adjust.adjust(int64(pkg.Dts))
			cli.WriteVideo(codec.CODECID_VIDEO_H264, pkg.Data, uint32(pts), uint32(dts))
		} else if pkg.Cid == mp4.MP4_CODEC_AAC {
			pts := audio_ts_adjust.adjust(int64(pkg.Pts))
			dts := video_dts_adjust.adjust(int64(pkg.Dts))
			cli.WriteAudio(codec.CODECID_AUDIO_AAC, pkg.Data, uint32(pts), uint32(dts))
		} else if pkg.Cid == mp4.MP4_CODEC_MP3 {
			pts := audio_ts_adjust.adjust(int64(pkg.Pts))
			dts := video_dts_adjust.adjust(int64(pkg.Dts))
			cli.WriteAudio(codec.CODECID_AUDIO_MP3, pkg.Data, uint32(pts), uint32(dts))
		}

	}
	return
}
