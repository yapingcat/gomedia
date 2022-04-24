package main

import (
	"fmt"
	"os"

	"github.com/yapingcat/gomedia/codec"
	"github.com/yapingcat/gomedia/flv"
	"github.com/yapingcat/gomedia/mp4"
)

type mymp4writer struct {
	fp *os.File
}

func newmymp4writer(f *os.File) *mymp4writer {
	return &mymp4writer{
		fp: f,
	}
}

func (mp4w *mymp4writer) Write(p []byte) (n int, err error) {
	return mp4w.fp.Write(p)
}
func (mp4w *mymp4writer) Seek(offset int64, whence int) (int64, error) {
	return mp4w.fp.Seek(offset, whence)
}
func (mp4w *mymp4writer) Tell() (offset int64) {
	offset, _ = mp4w.fp.Seek(0, 1)
	return
}

func main() {
	mp4filename := "test2.mp4"
	mp4file, err := os.OpenFile(mp4filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mp4file.Close()

	muxer := mp4.CreateMp4Muxer(newmymp4writer(mp4file))
	vtid := muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
	atid := muxer.AddAudioTrack(mp4.MP4_CODEC_AAC, 0, 16, 44100)

	flvfilereader, _ := os.Open(os.Args[1])
	defer flvfilereader.Close()
	fr := flv.CreateFlvReader()

	fr.OnFrame = func(ci codec.CodecID, b []byte, pts, dts uint32) {
		if ci == codec.CODECID_AUDIO_AAC {
			err := muxer.Write(atid, b, uint64(pts), uint64(dts))
			if err != nil {
				fmt.Println()
			}

		} else if ci == codec.CODECID_VIDEO_H264 {
			err := muxer.Write(vtid, b, uint64(pts), uint64(dts))
			if err != nil {
				fmt.Println()
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
	muxer.Writetrailer()
}
