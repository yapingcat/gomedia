package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/yapingcat/gomedia/go-mp4"
)

func mov_tag(tag [4]byte) uint32 {
	return binary.LittleEndian.Uint32(tag[:])
}

func main() {
	mp4FilePath := os.Args[1]
	newTime, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	timeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(timeBuf, uint32(newTime))

	mp4Fd, err := os.OpenFile(mp4FilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer mp4Fd.Close()

Loop:
	for err == nil {
		basebox := mp4.BasicBox{}
		_, err = basebox.Decode(mp4Fd)
		if err != nil {
			break
		}
		if basebox.Size < mp4.BasicBoxLen {
			err = errors.New("mp4 Parser error")
			break
		}
		fmt.Println(string(basebox.Type[:]))
		tagName := mov_tag(basebox.Type)
		switch tagName {
		case mov_tag([4]byte{'m', 'o', 'o', 'v'}):
			fmt.Println("Got moov box")
		case mov_tag([4]byte{'m', 'v', 'h', 'd'}):
			fmt.Println("Got mvhd box")
			offset, _ := mp4Fd.Seek(0, io.SeekCurrent)
			mvhd := mp4.MovieHeaderBox{Box: new(mp4.FullBox)}
			if _, err = mvhd.Decode(mp4Fd); err != nil {
				break Loop
			}
			offset2, _ := mp4Fd.Seek(0, io.SeekCurrent)
			if mvhd.Box.Version == 0 {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write(timeBuf) //modify time
			} else {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //modify time
			}
			mp4Fd.Seek(offset2, io.SeekStart)
		case mov_tag([4]byte{'t', 'r', 'a', 'k'}):
			fmt.Println("Got trak box")
		case mov_tag([4]byte{'m', 'd', 'i', 'a'}):
			fmt.Println("Got mdia box")
		case mov_tag([4]byte{'m', 'd', 'h', 'd'}):
			fmt.Println("Got mdhd box")
			offset, _ := mp4Fd.Seek(0, io.SeekCurrent)
			mdhd := mp4.MediaHeaderBox{Box: new(mp4.FullBox)}
			if _, err = mdhd.Decode(mp4Fd); err != nil {
				break Loop
			}
			offset2, _ := mp4Fd.Seek(0, io.SeekCurrent)
			if mdhd.Box.Version == 0 {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write(timeBuf) //modify time
			} else {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //modify time
			}
			mp4Fd.Seek(offset2, io.SeekStart)
		case mov_tag([4]byte{'t', 'k', 'h', 'd'}):
			fmt.Println("Got tkhd box")
			offset, _ := mp4Fd.Seek(0, io.SeekCurrent)
			tkhd := mp4.TrackHeaderBox{Box: new(mp4.FullBox)}
			if _, err = tkhd.Decode(mp4Fd); err != nil {
				break Loop
			}
			offset2, _ := mp4Fd.Seek(0, io.SeekCurrent)
			if tkhd.Box.Version == 0 {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write(timeBuf) //modify time
			} else {
				mp4Fd.Seek(offset+4, io.SeekStart)
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //create time
				mp4Fd.Write([]byte{0x00, 0x00, 0x00, 0x00})
				mp4Fd.Write(timeBuf) //modify time
			}
			mp4Fd.Seek(offset2, io.SeekStart)
		default:
			_, err = mp4Fd.Seek(int64(basebox.Size)-mp4.BasicBoxLen, io.SeekCurrent)
		}
	}
	if err != io.EOF {
		panic(err)
	}
	return
}
