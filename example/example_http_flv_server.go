package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
)

func onHttpFlv(w http.ResponseWriter, r *http.Request) {
	fmt.Println("on http flv")
	w.Header().Set("Content-Type", "video/x-flv")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	streamName := r.URL.Path
	streamName = strings.TrimLeft(streamName, "/live/")
	fmt.Println("start go routine", streamName)
	reader := flv.CreateFlvReader()
	writer := flv.CreateFlvWriter(w)
	writer.WriteFlvHeader()
	reader.OnFrame = func(cid codec.CodecID, frame []byte, pts, dts uint32) {
		if cid == codec.CODECID_AUDIO_AAC {
			writer.WriteAAC(frame, pts, dts)
		} else if cid == codec.CODECID_VIDEO_H264 {
			fmt.Println(len(frame))
			if codec.IsH264VCLNaluType(codec.H264NaluType(frame)) {
				time.Sleep(time.Millisecond * 40)
			}
			writer.WriteH264(frame, pts, dts)
		}
	}

	fileReader, err := os.Open(streamName)
	defer fileReader.Close()
	fmt.Println(err)
	cache := make([]byte, 4096)
	for {
		n, err := fileReader.Read(cache)
		if err != nil {
			fmt.Println(err)
			break
		}
		reader.Input(cache[0:n])
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/live/", onHttpFlv)
	server := http.Server{
		Addr:         ":19999",
		Handler:      mux,
		ReadTimeout:  time.Second * 1200,
		WriteTimeout: time.Second * 1200,
	}
	fmt.Println("server.listen")
	fmt.Println(server.ListenAndServe())
	select {}
}
