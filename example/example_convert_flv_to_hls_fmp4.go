package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yapingcat/gomedia/codec"
	"github.com/yapingcat/gomedia/flv"
	"github.com/yapingcat/gomedia/mp4"
)

type hlsSegment struct {
	duration float32
	uri      string
}

type hlsMuxer struct {
	initUri  string
	segments []hlsSegment
}

func (muxer *hlsMuxer) makeM3u8() string {
	buf := make([]byte, 0, 4096)
	m3u := bytes.NewBuffer(buf)
	maxDuration := 0
	for _, seg := range muxer.segments {
		if maxDuration < int(math.Ceil(float64(seg.duration))) {
			maxDuration = int(math.Ceil(float64(seg.duration)))
		}
	}

	m3u.WriteString("#EXTM3U\n")
	m3u.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", maxDuration))
	m3u.WriteString("#EXT-X-VERSION:7\n")
	m3u.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
	m3u.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	m3u.WriteString(fmt.Sprintf("#EXT-X-MAP:URI=\"%s\"\n", muxer.initUri))

	for _, seg := range muxer.segments {
		m3u.WriteString(fmt.Sprintf("#EXTINF:%.3f,%s\n", seg.duration, "no desc"))
		m3u.WriteString(seg.uri + "\n")
	}
	m3u.WriteString("#EXT-X-ENDLIST\n")
	return m3u.String()
}

func generateM3U8(flvFile string) {
	hls := &hlsMuxer{}
	var muxer *mp4.Movmuxer = nil
	var vtid uint32
	var atid uint32
	i := 0
	filename := fmt.Sprintf("stream-%d.mp4", i)
	mp4file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	muxer, err = mp4.CreateMp4Muxer(mp4file, mp4.WithMp4Flag(mp4.MP4_FLAG_DASH))
	if err != nil {
		fmt.Println(err)
		return
	}
	muxer.OnNewFragment(func(duration uint32, firstPts, firstDts uint64) {
		fmt.Println("on segment", duration)
		hls.segments = append(hls.segments, hlsSegment{
			uri:      filename,
			duration: float32(duration) / 1000,
		})

		mp4file.Close()
		if i == 0 {
			initFile, _ := os.OpenFile("init.mp4", os.O_CREATE|os.O_RDWR, 0666)
			muxer.WriteInitSegment(initFile)
			initFile.Close()
			hls.initUri = "init.mp4"
		}

		i++
		filename = fmt.Sprintf("stream-%d.mp4", i)
		mp4file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		muxer.ReBindWriter(mp4file)
	})
	vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
	atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)

	flvfilereader, _ := os.Open(flvFile)
	defer flvfilereader.Close()
	fr := flv.CreateFlvReader()

	fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
		if ci == codec.CODECID_AUDIO_AAC {
			err := muxer.Write(atid, b, uint64(pts), uint64(dts))
			if err != nil {
				fmt.Println(err)
			}
		} else if ci == codec.CODECID_VIDEO_H264 {
			err := muxer.Write(vtid, b, uint64(pts), uint64(dts))
			if err != nil {
				fmt.Println(err)
			}
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
	muxer.FlushFragment()
	m3u8Name := "test.m3u8"
	m3u8, _ := os.OpenFile(m3u8Name, os.O_CREATE|os.O_RDWR, 0666)
	m3u8.WriteString(hls.makeM3u8())
}

func onHLSVod(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	if strings.LastIndex(r.URL.Path, "m3u8") != -1 {
		fmt.Println("request m3u8", r.URL.Path)
		m3u8, err := os.Open("test.m3u8")
		if err != nil {
			return
		}
		defer m3u8.Close()
		b, _ := ioutil.ReadAll(m3u8)
		buf.Write(b)
		w.Header().Add("Content-Type", "application/vnd.apple.mpegurl")
	} else {
		fmt.Println("request fmp4", r.URL.Path)
		fmp4File := strings.TrimLeft(r.URL.Path, "/vod/")
		fmp4, err := os.Open(fmp4File)
		if err != nil {
			return
		}
		defer fmp4.Close()
		b, _ := ioutil.ReadAll(fmp4)
		buf.Write(b)
		w.Header().Set("Content-Type", "video/mp4")
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Write(buf.Bytes())
}

//http://127.0.0.1:19999/vod/test.m3u8
func main() {
	generateM3U8(os.Args[1])

	mux := http.NewServeMux()
	mux.HandleFunc("/vod/", onHLSVod)
	server := http.Server{
		Addr:         ":19999",
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	fmt.Println("server.listen")
	fmt.Println(server.ListenAndServe())

}
