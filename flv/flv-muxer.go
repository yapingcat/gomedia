package flv

import (
    "bytes"
    "errors"

    "github.com/yapingcat/gomedia/mpeg"
)

func WriteAudioTag(data []byte, cid FLV_SOUND_FORMAT, isSequenceHeader bool) []byte {
    var atag AudioTag
    atag.SoundFormat = uint8(cid)
    if cid == FLV_AAC {
        atag.SoundRate = uint8(FLV_SAMPLE_44000)
    } else if cid == FLV_G711A || cid == FLV_G711U {
        atag.SoundRate = uint8(FLV_SAMPLE_5500)
    }
    atag.SoundSize = 1
    atag.SoundType = 1
    if isSequenceHeader {
        atag.AACPacketType = 0
    } else {
        atag.AACPacketType = 1
    }
    tagData := atag.Encode()
    tagData = append(tagData, data...)
    return tagData
}

func WriteVideoTag(data []byte, cid FLV_VIDEO_CODEC_ID, cts int32, isSequenceHeader bool) []byte {
    var vtag VideoTag
    vtag.CodecId = uint8(cid)
    vtag.CompositionTime = cts
    if cid == FLV_AVC {
        nalType := mpeg.H264NaluTypeWithoutStartCode(data[4:])
        if nalType == mpeg.H264_NAL_I_SLICE {
            vtag.FrameType = uint8(KEY_FRAME)
        } else {
            vtag.FrameType = uint8(INTER_FRAME)
        }
    }
    if isSequenceHeader {
        vtag.AVCPacketType = uint8(AVC_SEQUENCE_HEADER)
    } else {
        vtag.AVCPacketType = uint8(AVC_NALU)
    }
    tagData := vtag.Encode()
    tagData = append(tagData, data...)
    return tagData
}

type AVTagMuxer interface {
    Write(frames []byte, pts uint32, dts uint32) [][]byte
}

type AVCMuxer struct {
    spsset map[uint64][]byte
    ppsset map[uint64][]byte
    cache  []byte
}

func (muxer *AVCMuxer) Write(frames []byte, pts uint32, dts uint32) [][]byte {
    var updateSequence bool = false
    var vcl bool = false
    mpeg.SplitFrameWithStartCode(frames, func(nalu []byte) bool {
        start, sc := mpeg.FindStarCode(nalu, 0)
        naltype := mpeg.H264NaluTypeWithoutStartCode(nalu[start+int(sc):])
        bs := mpeg.NewBitStream(nalu[start+int(sc):])
        switch naltype {
        case mpeg.H264_NAL_SPS:
            var sps mpeg.SPS
            sps.Decode(bs)
            s, found := muxer.spsset[sps.Seq_parameter_set_id]
            if !found || !bytes.Equal(s, nalu) {
                naluCopy := make([]byte, len(nalu))
                copy(naluCopy, nalu)
                muxer.spsset[sps.Seq_parameter_set_id] = naluCopy
                updateSequence = true
            }
            muxer.cache = append(muxer.cache, mpeg.ConvertAnnexBToAVCC(nalu)...)
        case mpeg.H264_NAL_PPS:
            var pps mpeg.PPS
            pps.Decode(bs)
            muxer.ppsset[pps.Pic_parameter_set_id] = nalu
            s, found := muxer.ppsset[pps.Pic_parameter_set_id]
            if !found || !bytes.Equal(s, nalu) {
                naluCopy := make([]byte, len(nalu))
                copy(naluCopy, nalu)
                muxer.ppsset[pps.Pic_parameter_set_id] = naluCopy
                updateSequence = true
            }
            muxer.cache = append(muxer.cache, mpeg.ConvertAnnexBToAVCC(nalu)...)
        default:
            if naltype <= mpeg.H264_NAL_I_SLICE {
                vcl = true
            }
            muxer.cache = append(muxer.cache, mpeg.ConvertAnnexBToAVCC(nalu)...)
        }
        return true
    })
    var tags [][]byte
    if updateSequence && len(muxer.ppsset) > 0 && len(muxer.spsset) > 0 {
        spss := make([][]byte, len(muxer.spsset))
        idx := 0
        for _, sps := range muxer.spsset {
            spss[idx] = sps
            idx++
        }
        idx = 0
        ppss := make([][]byte, len(muxer.ppsset))
        for _, pps := range muxer.ppsset {
            ppss[idx] = pps
            idx++
        }
        extraData := mpeg.CreateH264AVCCExtradata(spss, ppss)
        tags = append(tags, WriteVideoTag(extraData, FLV_AVC, 0, true))
    }
    if vcl {
        tags = append(tags, WriteVideoTag(muxer.cache, FLV_AVC, int32(pts-dts), false))
        muxer.cache = muxer.cache[:0]
    }
    return tags
}

type HevcMuxer struct {
}

func CreateVideoMuxer(cid FLV_VIDEO_CODEC_ID) AVTagMuxer {
    if cid == FLV_AVC {
        return &AVCMuxer{
            spsset: make(map[uint64][]byte),
            ppsset: make(map[uint64][]byte),
            cache:  make([]byte, 0, 1024),
        }
    } else if cid == FLV_HEVC {
        //TODO FLV Support H265
        return nil
    }
    return nil
}

type AACMuxer struct {
    updateSequence bool
}

func (muxer *AACMuxer) Write(frames []byte, pts uint32, dts uint32) [][]byte {
    var tags [][]byte
    mpeg.SplitAACFrame(frames, func(aac []byte) {
        hdr := mpeg.NewAdtsFrameHeader()
        hdr.Decode(aac)
        if muxer.updateSequence {
            asc, _ := mpeg.ConvertADTSToASC(aac)
            tags = append(tags, WriteAudioTag(asc, FLV_AAC, true))
            muxer.updateSequence = false
        }
        tags = append(tags, WriteAudioTag(aac[7:], FLV_AAC, false))
    })
    return tags
}

type G711AMuxer struct {
}

func (muxer *G711AMuxer) Write(frames []byte, pts uint32, dts uint32) [][]byte {
    tags := make([][]byte, 1)
    tags[0] = WriteAudioTag(frames, FLV_G711A, true)
    return tags
}

type G711UMuxer struct {
}

func (muxer *G711UMuxer) Write(frames []byte, pts uint32, dts uint32) [][]byte {
    tags := make([][]byte, 1)
    tags[0] = WriteAudioTag(frames, FLV_G711U, true)
    return tags
}

func CreateAudioMuxer(cid FLV_SOUND_FORMAT) AVTagMuxer {
    if cid == FLV_AAC {
        return &AACMuxer{updateSequence: true}
    } else if cid == FLV_G711A {
        return new(G711AMuxer)
    } else if cid == FLV_G711U {
        return new(G711UMuxer)
    } else {
        return nil
    }
}

type FlvMuxer struct {
    videoMuxer AVTagMuxer
    audioMuxer AVTagMuxer
}

func NewFlvMuxer(vid FLV_VIDEO_CODEC_ID, aid FLV_SOUND_FORMAT) *FlvMuxer {
    return &FlvMuxer{
        videoMuxer: CreateVideoMuxer(vid),
        audioMuxer: CreateAudioMuxer(aid),
    }
}

func (muxer *FlvMuxer) SetVideoCodeId(cid FLV_VIDEO_CODEC_ID) {
    muxer.videoMuxer = CreateVideoMuxer(cid)
}

func (muxer *FlvMuxer) SetAudioCodeId(cid FLV_SOUND_FORMAT) {
    muxer.audioMuxer = CreateAudioMuxer(cid)
}

func (muxer *FlvMuxer) WriteVideo(frames []byte, pts uint32, dts uint32) ([][]byte, error) {
    if muxer.videoMuxer == nil {
        return nil, errors.New("video Muxer is Nil")
    }
    return muxer.WriteFrames(VIDEO_TAG, frames, pts, dts)
}

func (muxer *FlvMuxer) WriteAudio(frames []byte, pts uint32, dts uint32) ([][]byte, error) {
    if muxer.audioMuxer == nil {
        return nil, errors.New("audio Muxer is Nil")
    }
    return muxer.WriteFrames(AUDIO_TAG, frames, pts, dts)
}

func (muxer *FlvMuxer) WriteFrames(frameType TagType, frames []byte, pts uint32, dts uint32) ([][]byte, error) {

    var ftag FlvTag
    var tags [][]byte
    if frameType == AUDIO_TAG {
        ftag.TagType = uint8(AUDIO_TAG)
        tags = muxer.audioMuxer.Write(frames, pts, dts)
    } else if frameType == VIDEO_TAG {
        ftag.TagType = uint8(VIDEO_TAG)
        tags = muxer.videoMuxer.Write(frames, pts, dts)
    } else {
        return nil, errors.New("unsupport Frame Type")
    }
    ftag.Timestamp = dts & 0x0FFF
    ftag.TimestampExtended = uint8(dts >> 24 & 0xFF)

    tmptags := make([][]byte, 0, 1)
    for _, tag := range tags {
        ftag.DataSize = uint32(len(tag))
        vtag := ftag.Encode()
        vtag = append(vtag, tag...)
        tmptags = append(tmptags, vtag)
    }
    return tmptags, nil
}
