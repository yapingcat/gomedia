package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

func main() {
	tsFileName := os.Args[1]
	mp4FileName := os.Args[2]

	tsFd, err := os.Open(tsFileName)
	if err != nil {
		panic(err)
	}
	defer tsFd.Close()

	hasAudio := false
	hasVideo := false
	var atid uint32 = 0
	var vtid uint32 = 0

	mp4file, err := os.OpenFile(mp4FileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer mp4file.Close()

	muxer, err := mp4.CreateMp4Muxer(mp4file)
	if err != nil {
		panic(err)
	}

	demuxer := mpeg2.NewTSDemuxer()
	demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
		if cid == mpeg2.TS_STREAM_H264 {
			if !hasVideo {
				vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
				hasVideo = true
			}
			err = muxer.Write(vtid, frame, uint64(pts), uint64(dts))
			if err != nil {
				panic(err)
			}
		} else if cid == mpeg2.TS_STREAM_AAC {
			if !hasAudio {
				atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
				hasAudio = true
			}
			err = muxer.Write(atid, frame, uint64(pts), uint64(dts))
			if err != nil {
				panic(err)
			}
		} else if cid == mpeg2.TS_STREAM_AUDIO_MPEG1 || cid == mpeg2.TS_STREAM_AUDIO_MPEG2 {
			if !hasAudio {
				atid = muxer.AddAudioTrack(mp4.MP4_CODEC_MP3)
				hasAudio = true
			}
			err = muxer.Write(atid, frame, uint64(pts), uint64(dts))
			if err != nil {
				panic(err)
			}
		}
	}

	buf, err := ioutil.ReadAll(tsFd)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %d size\n", len(buf))
	err = demuxer.Input(bytes.NewReader(buf))
	if err != nil {
		panic(err)
	}

	err = muxer.WriteTrailer()
	if err != nil {
		panic(err)
	}
}
