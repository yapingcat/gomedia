package flv

import (
	"errors"
	"io"
	"os"

	"../mpeg"
)

//	FLV File
//	TheFLVheader
//	An FLV file shall begin with the FLV header:
//	-------------------------------------------------------------------------------------------------------------
//	FLV header
//	Field                       Type                            Comment
//	-------------------------------------------------------------------------------------------------------------
//	Signature                   UI8							Signature byte always 'F' (0x46)
//
//	Signature                   UI8		                    Signature byte always 'L' (0x4C)
//
//  Signature                   UI8                         Signature byte always 'V' (0x56)
//
//  Version                     UI8                         File version (for example, 0x01 for FLV version 1)
//
//  TypeFlagsReserved           UB[5]                       Shall be 0
//
//  TypeFlagsAudio              UB[1]                       1 = Audio tags are present
//
//  TypeFlagsReserved           UB[1]                       Shall be 0
//
//  TypeFlagsVideo              UB[1]                       1 = Video tags are present
//
//  DataOffset                  UI32                        The length of this header in bytes
//  -------------------------------------------------------------------------------------------------------------
//
//  TheFLVFileBody
//  After the FLV header, the remainder of an FLV file shall consist of alternating back-pointers and tags.
//  They interleave as shown in the following table:
//  -------------------------------------------------------------------------------------------------------------
//  FLV File Body
//  Field                       Type                        Comment
//  -------------------------------------------------------------------------------------------------------------
//	PreviousTagSize0            UI32                        Always 0
//  Tag1                        FLVTAG                      First tag
//  PreviousTagSize1            UI32                        Size of previous tag, including its header, in bytes. For FLV version 1,
//                                                          this value is 11 plus the DataSize of the previous tag
//  Tag2                        FLVTAG                      Second tag
//
//  ....
//
//  PreviousTagSizeN-1          UI32                        Size of second-to-last tag, including its header, in bytes.
//  ---------------------------------------------------------------------------------------------------------------

type FlvFileWriter struct {
	//vcodecid mpeg.CodecID
	//acodecid mpeg.CodecID
	firstAudio bool
	firstVideo bool
	fd         *os.File
}

type FlvFileReader struct {
	fd      *os.File
	onFrame func([]byte, uint32, mpeg.CodecID)
	onTag   func(ftag FlvTag, tag interface{})
}

func CreateFlvFileReader() *FlvFileReader {
	flvFile := &FlvFileReader{
		fd:      nil,
		onFrame: nil,
		onTag:   nil,
	}
	return flvFile
}

func (f *FlvFileReader) SetOnFrame(onframe func([]byte, uint32, mpeg.CodecID)) {
	f.onFrame = onframe
}

func (f *FlvFileReader) SetOnTag(ontag func(ftag FlvTag, tag interface{})) {
	f.onTag = ontag
}

func (f *FlvFileReader) Open(filepath string) (err error) {

	if f.fd, err = os.Open(filepath); err != nil {
		return err
	}
	var flvhdr [9]byte
	if count, err := f.fd.Read(flvhdr[:]); err != nil {
		return err
	} else if count < 9 {
		err = errors.New("flv File Header < 9 Bytes")
		return err
	} else {
		if flvhdr[0] != 'F' || flvhdr[1] != 'L' || flvhdr[2] != 'V' {
			err = errors.New("this file Is Not FLV File")
			return err
		}
	}

	f.fd.Read(flvhdr[:4])
	return nil
}

func (f *FlvFileReader) DeMuxFile() error {
	for {
		var tag [16]byte
		if count, err := f.fd.Read(tag[:11]); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		} else if count < 11 {
			return errors.New("FLV File Header < 11 Bytes")
		}
		var ftag FlvTag
		var taglen int = 0
		ftag.Decode(tag[:11])
		if ftag.TagType == 8 {
			if _, err := f.fd.Read(tag[11:12]); err != nil {
				return err
			}
			var atag AudioTag
			if (tag[11]&0xF0)>>4 == 10 {
				if _, err := f.fd.Read(tag[12:13]); err != nil {
					return err
				} else {
					atag.Decode(tag[11:13])
					taglen = 2
				}
			} else {
				taglen = 1
				atag.Decode(tag[11:12])
			}
			if f.onTag != nil {
				f.onTag(ftag, atag)
			}
		} else if ftag.TagType == 9 {
			if _, err := f.fd.Read(tag[11:12]); err != nil {
				return err
			}
			var vtag VideoTag
			if tag[11]&0x0F == 7 {
				if _, err := f.fd.Read(tag[12:16]); err != nil {
					return err
				} else {
					vtag.Decode(tag[11:16])
					taglen = 5
				}
			} else {
				taglen = 1
				vtag.Decode(tag[11:12])
			}
			if f.onTag != nil {
				f.onTag(ftag, vtag)
			}
		} else if ftag.TagType == 18 {
			if f.onTag != nil {
				f.onTag(ftag, nil)
			}
		}
		data := make([]byte, ftag.DataSize-uint32(taglen))
		f.fd.Read(data)
		f.fd.Read(tag[:4])
	}
}

type FileWriter struct {
	fd *os.File
}

func CreateFlvFileWriter() *FlvFileWriter {
	flvFile := &FlvFileWriter{
		fd:         nil,
		firstAudio: true,
		firstVideo: true,
	}
	return flvFile
}

func (f *FileWriter) Open(path string) (err error) {

	if f.fd, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666); err != nil {
		return err
	}
	var flvhdr [9]byte
	flvhdr[0] = 'F'
	flvhdr[1] = 'L'
	flvhdr[2] = 'V'
	flvhdr[3] = 0x01
	flvhdr[4] = 0
	flvhdr[5] = 0
	flvhdr[6] = 0
	flvhdr[7] = 0
	flvhdr[8] = 0

	f.fd.Write(flvhdr[:])

	var previousTagSize0 [4]byte
	previousTagSize0[0] = 0
	previousTagSize0[1] = 0
	previousTagSize0[2] = 0
	previousTagSize0[3] = 0
	f.fd.Write(previousTagSize0[:])

	return nil
}

func (f *FlvFileWriter) MuxerAudio(data []byte, cid mpeg.CodecID, pts uint32, dts uint32, sample int, bitsPersample int) error {
	flvtag := FlvTag{
		StreamID:          0,
		TimestampExtended: 0,
		DataSize:          0,
	}
	flvtag.TagType = uint8(AUDIO_TAG)
	if pts > 0x00FFFFFF {
		flvtag.Timestamp = 0x00FFFFFF
		flvtag.TimestampExtended = uint8(pts >> 24)
	} else {
		flvtag.Timestamp = pts
	}
	flvtag.DataSize = uint32(len(data))
	count, err := f.fd.Write(flvtag.Encode())
	if err != nil {
		return err
	}
	if count != 11 {
		return errors.New("write to file < 11 bytes")
	}

	atag := AudioTag{
		SoundFormat: 0,
		SoundSize:   0,
		SoundType:   0,
		SoundRate:   0,
	}

	switch cid {
	case mpeg.CODECID_AUDIO_AAC:
		atag.SoundFormat = FLV_AAC
		atag.SoundRate = uint8(FLV_SAMPLE_44000)
		atag.SoundSize = 1
		atag.SoundType = 1
		if f.firstAudio {
			atag.AACPacketType = 0
		}
		tag := atag.Encode()

		c, err := f.fd.Write(tag)
		if err != nil {
			return err
		}

		if c != len(tag) {
			return errors.New("write to file too least bytes")
		}
		return f.muxerAAC(data)
	default:
		return errors.New("unsupport audio codec")
	}

}

func (f *FlvFileWriter) MuxerVideo() error {
	return nil
}

func (f *FlvFileWriter) muxerAAC(aac []byte) error {
	if f.firstAudio {
		f.firstAudio = false
	}
	return nil
}
