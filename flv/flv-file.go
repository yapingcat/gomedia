package flv

import (
    "errors"
    "io"
    "os"

    "github.com/yapingcat/gomedia/mpeg"
)

//  FLV File
//  TheFLVheader
//  An FLV file shall begin with the FLV header:
//  -------------------------------------------------------------------------------------------------------------
//  FLV header
//  Field                       Type                            Comment
//  -------------------------------------------------------------------------------------------------------------
//  Signature                   UI8                            Signature byte always 'F' (0x46)
//
//  Signature                   UI8                            Signature byte always 'L' (0x4C)
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
//  PreviousTagSize0            UI32                        Always 0
//  Tag1                        FLVTAG                      First tag
//  PreviousTagSize1            UI32                        Size of previous tag, including its header, in bytes. For FLV version 1,
//                                                          this value is 11 plus the DataSize of the previous tag
//  Tag2                        FLVTAG                      Second tag
//
//  ....
//
//  PreviousTagSizeN-1          UI32                        Size of second-to-last tag, including its header, in bytes.
//  ---------------------------------------------------------------------------------------------------------------

type FlvFileReader struct {
    fd      *os.File
    asc     []byte
    OnFrame func(mpeg.CodecID, []byte, uint32, uint32)
    OnTag   func(ftag FlvTag, tag interface{})
}

func CreateFlvFileReader() *FlvFileReader {
    flvFile := &FlvFileReader{
        fd:      nil,
        OnFrame: nil,
        OnTag:   nil,
    }
    return flvFile
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
        var vtag VideoTag
        var atag AudioTag
        var taglen int = 0
        var vcid FLV_VIDEO_CODEC_ID
        var acid FLV_SOUND_FORMAT
        var packtype uint8
        ftag.Decode(tag[:11])
        if ftag.TagType == uint8(AUDIO_TAG) {
            if _, err := f.fd.Read(tag[11:12]); err != nil {
                return err
            }
            if (tag[11]&0xF0)>>4 == byte(FLV_AAC) {
                if _, err := f.fd.Read(tag[12:13]); err != nil {
                    return err
                } else {
                    atag.Decode(tag[11:13])
                    acid = FLV_SOUND_FORMAT(atag.SoundFormat)
                    packtype = atag.AACPacketType
                    taglen = 2
                }
            } else {
                taglen = 1
                atag.Decode(tag[11:12])
                acid = FLV_SOUND_FORMAT(atag.SoundFormat)
            }
            if f.OnTag != nil {
                f.OnTag(ftag, atag)
            }
        } else if ftag.TagType == uint8(VIDEO_TAG) {
            if _, err := f.fd.Read(tag[11:12]); err != nil {
                return err
            }
            if tag[11]&0x0F == byte(FLV_AVC) {
                if _, err := f.fd.Read(tag[12:16]); err != nil {
                    return err
                } else {
                    vtag.Decode(tag[11:16])
                    taglen = 5
                    vcid = FLV_VIDEO_CODEC_ID(vtag.CodecId)
                    packtype = vtag.AVCPacketType
                }
            } else {
                taglen = 1
                vtag.Decode(tag[11:12])
                vcid = FLV_VIDEO_CODEC_ID(vtag.CodecId)
            }
            if f.OnTag != nil {
                f.OnTag(ftag, vtag)
            }
        } else if ftag.TagType == uint8(SCRIPT_TAG) {
            if f.OnTag != nil {
                f.OnTag(ftag, nil)
            }
        }
        pts := uint32(ftag.TimestampExtended)<<24 | ftag.Timestamp
        data := make([]byte, ftag.DataSize-uint32(taglen))
        f.fd.Read(data)
        if ftag.TagType == uint8(AUDIO_TAG) {
            dts := pts
            f.demuxAudio(acid, packtype, data, pts, dts)
        } else if ftag.TagType == uint8(VIDEO_TAG) {
            if vcid == FLV_AVC || vcid == FLV_HEVC {
                dts := pts
                if vtag.CompositionTime < 0 {
                    dts -= uint32(-1 * vtag.CompositionTime)
                } else {
                    dts = uint32(vtag.CompositionTime)
                }
                f.demuxAudio(acid, packtype, data, pts, dts)
            }
        }
        f.fd.Read(tag[:4])
    }
}

func (f *FlvFileReader) demuxAudio(cid FLV_SOUND_FORMAT, packetType uint8, data []byte, pts uint32, dts uint32) {
    var audioFrame []byte
    if cid == FLV_AAC {
        if packetType == AAC_SEQUENCE_HEADER {
            f.asc = data
        } else {
            audioFrame = mpeg.ConvertASCToADTS(f.asc, len(data)+7)
            audioFrame = append(audioFrame, data...)
        }
    } else {
        audioFrame = data
    }
    if f.OnFrame != nil {
        f.OnFrame(CovertFlvAudioCodecId2MpegCodecId(cid), audioFrame, pts, dts)
    }
}

func (f *FlvFileReader) demuxVideo(cid FLV_VIDEO_CODEC_ID, packetType uint8, data []byte, pts uint32, dts uint32) {
    return
}

type FlvFileWriter struct {
    fd    *os.File
    muxer *FlvMuxer
}

func CreateFlvFileWriter() *FlvFileWriter {
    flvFile := &FlvFileWriter{
        fd:    nil,
        muxer: new(FlvMuxer),
    }
    return flvFile
}

func (f *FlvFileWriter) Open(path string) (err error) {

    if f.fd, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666); err != nil {
        return err
    }
    var flvhdr [9]byte
    flvhdr[0] = 'F'
    flvhdr[1] = 'L'
    flvhdr[2] = 'V'
    flvhdr[3] = 0x01
    flvhdr[4] = 0x05
    flvhdr[5] = 0
    flvhdr[6] = 0
    flvhdr[7] = 0
    flvhdr[8] = 9

    f.fd.Write(flvhdr[:])

    var previousTagSize0 [4]byte
    previousTagSize0[0] = 0
    previousTagSize0[1] = 0
    previousTagSize0[2] = 0
    previousTagSize0[3] = 0
    f.fd.Write(previousTagSize0[:])

    return nil
}

//adts aac frame
func (f *FlvFileWriter) WriteAAC(data []byte, pts uint32, dts uint32) error {
    if f.muxer.audioMuxer == nil {
        f.muxer.SetAudioCodeId(FLV_AAC)
    } else {
        if _, ok := f.muxer.audioMuxer.(*AACMuxer); !ok {
            panic("audio codec change")
        }
    }
    if tags, err := f.muxer.WriteAudio(data, pts, dts); err != nil {
        return err
    } else {
        for _, tag := range tags {
            f.fd.Write(tag)
        }
    }
    return nil
}

//H264 Frame with startcode 0x0000001
func (f *FlvFileWriter) WriteH264(data []byte, pts uint32, dts uint32) error {
    if f.muxer.videoMuxer == nil {
        f.muxer.SetVideoCodeId(FLV_AVC)
    } else {
        if _, ok := f.muxer.audioMuxer.(*AVCMuxer); !ok {
            panic("video codec change")
        }
    }
    if tags, err := f.muxer.WriteVideo(data, pts, dts); err != nil {
        return err
    } else {
        for _, tag := range tags {
            f.fd.Write(tag)
        }
    }
    return nil
}
