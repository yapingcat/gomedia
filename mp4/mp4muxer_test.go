package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/yapingcat/gomedia/mpeg"
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

func TestCreateMp4Reader(t *testing.T) {
	f, err := os.Open("jellyfish-3-mbps-hd.h264.mp4")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	for err == nil {
		nn := int64(0)
		size := make([]byte, 4)
		_, err = io.ReadFull(f, size)
		if err != nil {
			break
		}
		nn += 4
		boxtype := make([]byte, 4)
		_, err = io.ReadFull(f, boxtype)
		if err != nil {
			break
		}
		nn += 4
		var isize uint64 = uint64(binary.BigEndian.Uint32(size))
		if isize == 1 {
			size := make([]byte, 8)
			_, err = io.ReadFull(f, size)
			if err != nil {
				break
			}
			isize = binary.BigEndian.Uint64(size)
			nn += 8
		}
		fmt.Printf("Read Box(%s) size:%d\n", boxtype, isize)
		f.Seek(int64(isize)-nn, 1)
	}
}

func TestCreateMp4Muxer(t *testing.T) {

	f, err := os.Open("jellyfish-3-mbps-hd.h265")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	mp4filename := "jellyfish-3-mbps-hd.h265.mp4"
	mp4file, err := os.OpenFile(mp4filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mp4file.Close()

	buf, _ := ioutil.ReadAll(f)
	pts := uint64(0)
	dts := uint64(0)
	ii := [3]uint64{33, 33, 34}
	idx := 0

	type args struct {
		wh Writer
	}
	tests := []struct {
		name string
		args args
		want *Movmuxer
	}{
		{name: "muxer h264", args: args{wh: newmymp4writer(mp4file)}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			muxer := CreateMp4Muxer(tt.args.wh)
			tid := muxer.AddVideoTrack(MP4_CODEC_H265)
			cache := make([]byte, 0)
			mpeg.SplitFrameWithStartCode(buf, func(nalu []byte) bool {
				ntype := mpeg.H265NaluType(nalu)
				if !mpeg.IsH265VCLNaluType(ntype) {
					cache = append(cache, nalu...)
					return true
				}
				if len(cache) > 0 {
					cache = append(cache, nalu...)
					muxer.Write(tid, cache, pts, dts)
					cache = cache[:0]
				} else {
					muxer.Write(tid, nalu, pts, dts)
				}
				pts += ii[idx]
				dts += ii[idx]
				idx++
				idx = idx % 3
				return true
			})
			fmt.Printf("last dts %d\n", dts)
			muxer.Writetrailer()
		})
	}
}
