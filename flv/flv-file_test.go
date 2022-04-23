package flv

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/yapingcat/gomedia/mpeg"
)

func TestFlvReader_Input(t *testing.T) {
	t.Run("test_1", func(t *testing.T) {
		videoFile, _ := os.OpenFile("v.h264", os.O_CREATE|os.O_RDWR, 0666)
		defer videoFile.Close()
		audioFile, _ := os.OpenFile("a.aac", os.O_CREATE|os.O_RDWR, 0666)
		defer audioFile.Close()
		f := CreateFlvReader()
		f.OnFrame = func(cid mpeg.CodecID, frame []byte, pts, dts uint32) {
			if cid == mpeg.CODECID_VIDEO_H264 {
				videoFile.Write(frame)
			} else if cid == mpeg.CODECID_AUDIO_AAC {
				audioFile.Write(frame)
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

		// content, _ := ioutil.ReadAll(fd)
		// if err := f.Input(content); err != nil {
		// 	t.Errorf("FlvReader.Input() error = %v", err)
		// }
	})
}

func TestFlvWriter_Write(t *testing.T) {

	t.Run("test_2", func(t *testing.T) {
		newflv, _ := os.OpenFile("new.flv", os.O_CREATE|os.O_RDWR, 0666)
		defer newflv.Close()
		wf := CreateFlvWriter(newflv)
		wf.WriteFlvHeader()
		rf := CreateFlvReader()
		rf.OnFrame = func(cid mpeg.CodecID, frame []byte, pts, dts uint32) {
			if cid == mpeg.CODECID_VIDEO_H264 {
				if err := wf.WriteH264(frame, pts, dts); err != nil {
					fmt.Println(err)
				}
			} else if cid == mpeg.CODECID_AUDIO_AAC {
				if err := wf.WriteAAC(frame, pts, dts); err != nil {
					fmt.Println(err)
				}
			}
		}
		fd, _ := os.Open("source.200kbps.768x320.flv")
		defer fd.Close()
		content, _ := ioutil.ReadAll(fd)
		if err := rf.Input(content); err != nil {
			t.Errorf("FlvReader.Input() error = %v", err)
		}
	})
}

func TestFlvWriter_WriteHevc(t *testing.T) {

	t.Run("test_3", func(t *testing.T) {
		newflv, _ := os.OpenFile("h265.flv", os.O_CREATE|os.O_RDWR, 0666)
		defer newflv.Close()
		wf := CreateFlvWriter(newflv)
		wf.WriteFlvHeader()
		var pts uint32 = 0
		var dts uint32 = 0
		rawh265, err := os.Open("1.h265")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rawh265.Close()
		buf, _ := ioutil.ReadAll(rawh265)
		mpeg.SplitFrameWithStartCode(buf, func(nalu []byte) bool {
			fmt.Printf("%x %x %x %x %x\n", nalu[0], nalu[1], nalu[2], nalu[3], nalu[4])
			fmt.Printf("nalu size %d\n", len(nalu))
			if err := wf.WriteH265(nalu, pts, dts); err != nil {
				fmt.Println(err)
			}
			pts += 40
			dts += 40
			return true
		})
	})
}

func TestFlvReadH265(t *testing.T) {

	t.Run("test_4", func(t *testing.T) {
		videoFile, _ := os.OpenFile("v2.h265", os.O_CREATE|os.O_RDWR, 0666)
		defer videoFile.Close()
		f := CreateFlvReader()
		f.OnFrame = func(cid mpeg.CodecID, frame []byte, pts, dts uint32) {
			if cid == mpeg.CODECID_VIDEO_H265 {
				videoFile.Write(frame)
			}
		}
		fd, _ := os.Open("l.flv")
		defer fd.Close()
		content, _ := ioutil.ReadAll(fd)
		if err := f.Input(content); err != nil {
			t.Errorf("FlvReader.Input() error = %v", err)
		}
	})
}
