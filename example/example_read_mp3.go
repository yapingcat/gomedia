package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yapingcat/gomedia/go-codec"
)

func main() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	mp3, _ := ioutil.ReadAll(f)
	fmt.Println("Get Mp3 file size", len(mp3))
	codec.SplitMp3Frames(mp3, func(head *codec.MP3FrameHead, frame []byte) {
		fmt.Println("Get mp3 Frame")
		fmt.Printf("mp3 frame head %+v\n", head)
		fmt.Printf("mp3 bitrate:%d,samplerate:%d,channelcount:%d\n", head.GetBitRate(), head.GetSampleRate(), head.GetChannelCount())
	})

}
