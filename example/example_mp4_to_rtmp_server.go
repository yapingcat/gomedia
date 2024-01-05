package main

import (
	"fmt"
	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-rtmp"
	"io"
	"net"
	"os"
	"time"
)

// Will push the last file under mp4sPath to the specified rtmp server
func main() {
	var (
		mp4sPath = "your_mp4_dir" //like ./mp4/
		rtmpUrl  = "rtmpUrl"      //like rtmp://127.0.0.1:1935/live/test110
	)
	c, err := net.Dial("tcp4", "127.0.0.1:1935")
	if err != nil {
		fmt.Println(err)
	}
	cli := rtmp.NewRtmpClient(rtmp.WithComplexHandshake(),
		rtmp.WithComplexHandshakeSchema(rtmp.HANDSHAKE_COMPLEX_SCHEMA1),
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
		fmt.Println("start to read mp4 file")
		ptsbase := uint32(0)
		for {
			filees, err := os.ReadDir(mp4sPath)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(mp4sPath + filees[len(filees)-1].Name())
			ptsbase = PushRtmp(mp4sPath+filees[len(filees)-1].Name(), cli, ptsbase)
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
		cli.Input(buf[:n])
	}
	fmt.Println(err)
}

func PushRtmp(fileName string, cli *rtmp.RtmpClient, ptsbase uint32) uint32 {
	mp4File, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return 0
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
			time.Sleep(10 * time.Millisecond)
			cli.WriteVideo(codec.CODECID_VIDEO_H264, pkg.Data, uint32(pkg.Pts)+ptsbase, uint32(pkg.Dts))
			ptsbase = uint32(pkg.Pts) + ptsbase
		} else if pkg.Cid == mp4.MP4_CODEC_AAC {
			cli.WriteAudio(codec.CODECID_AUDIO_AAC, pkg.Data, uint32(pkg.Pts), uint32(pkg.Dts))
		} else if pkg.Cid == mp4.MP4_CODEC_MP3 {
			cli.WriteAudio(codec.CODECID_AUDIO_MP3, pkg.Data, uint32(pkg.Pts), uint32(pkg.Dts))
		}
	}
	fmt.Println(ptsbase)
	return ptsbase
}
