package mp4

import (
    "errors"

    "github.com/yapingcat/gomedia/mpeg"
)

type sampleEntry struct {
    pts                    uint64
    dts                    uint64
    offset                 uint64
    size                   uint64
    SampleDescriptionIndex uint32 //alway should be 1
}

type movchunk struct {
    chunknum    uint32
    samplenum   uint32
    chunkoffset uint64
}

type mp4track struct {
    cid         MOV_CODEC_TYPE
    trackId     uint32
    stbltable   *movstbl
    duration    uint32
    timescale   uint32
    width       uint32
    height      uint32
    sampleRate  uint32
    sampleBits  uint8
    chanelCount uint8
    samplelist  []sampleEntry
    extra       extraData
}

func newmp4track(cid MOV_CODEC_TYPE) *mp4track {
    track := &mp4track{
        cid:        cid,
        stbltable:  nil,
        samplelist: make([]sampleEntry, 0),
    }
    if cid == MOV_CODEC_H264 {
        track.extra = new(h264ExtraData)
    } else if cid == MOV_CODEC_H265 {
        track.extra = new(h265ExtraData)
    } else if cid == MOV_CODEC_AAC {
        track.extra = new(aacExtraData)
    }
    return track
}

type extraData interface {
    export() []byte
    load(data []byte)
}

type h264ExtraData struct {
    spss [][]byte
    ppss [][]byte
}

func (extra *h264ExtraData) export() []byte {
    return mpeg.CreateH264AVCCExtradata(extra.spss, extra.ppss)
}

func (extra *h264ExtraData) load(data []byte) {
    extra.spss, extra.ppss = mpeg.CovertExtradata(data)
}

type h265ExtraData struct {
    hvccExtra mpeg.HEVCRecordConfiguration
}

func (extra *h265ExtraData) export() []byte {
    return extra.hvccExtra.Encode()
}

func (extra *h265ExtraData) load(data []byte) {
    hevcExtra := mpeg.HEVCRecordConfiguration{}
    hevcExtra.Decode(data)
}

type aacExtraData struct {
    asc []byte
}

func (extra *aacExtraData) export() []byte {
    return extra.asc
}

func (extra *aacExtraData) load(data []byte) {
    extra.asc = make([]byte, len(data))
    copy(extra.asc, data)
}

type Movmuxer struct {
    writerHandler Writer
    nextTrackId   uint32
    mdatOffset    uint32
    tracks        map[uint32]*mp4track
    width         uint32
    height        uint32
}

func CreateMp4Muxer(wh Writer) *Movmuxer {
    muxer := &Movmuxer{
        writerHandler: wh,
        nextTrackId:   0,
        tracks:        make(map[uint32]*mp4track),
    }
    ftyp := NewFileTypeBox()
    ftyp.Major_brand = mov_tag(isom)
    ftyp.Minor_version = 0x200
    ftyp.Compatible_brands = make([]uint32, 4)
    ftyp.Compatible_brands[0] = mov_tag(isom)
    ftyp.Compatible_brands[1] = mov_tag(iso2)
    ftyp.Compatible_brands[2] = mov_tag(avc1)
    ftyp.Compatible_brands[3] = mov_tag(mp41)
    len, boxdata := ftyp.Encode()
    muxer.writerHandler.Write(boxdata[0:len])
    free := NewFreeBox()
    freelen, freeboxdata := free.Encode()
    muxer.writerHandler.Write(freeboxdata[0:freelen])
    muxer.mdatOffset = uint32(muxer.writerHandler.Tell())
    MDAT.Size = 0
    mdatlen, mdat := MDAT.Encode()
    muxer.writerHandler.Write(mdat[0:mdatlen])
    return muxer
}

func (muxer *Movmuxer) AddAudioTrack(cid MOV_CODEC_TYPE, channelcount uint8, sampleBits uint8, sampleRate uint) uint32 {
    track := &mp4track{
        cid:         cid,
        trackId:     muxer.nextTrackId,
        sampleRate:  uint32(sampleRate),
        sampleBits:  sampleBits,
        chanelCount: channelcount,
    }
    muxer.tracks[muxer.nextTrackId] = track
    muxer.nextTrackId++
    return track.trackId
}

func (muxer *Movmuxer) AddVideoTrack(cid MOV_CODEC_TYPE) uint32 {
    track := &mp4track{
        cid:     cid,
        trackId: muxer.nextTrackId,
    }
    muxer.tracks[muxer.nextTrackId] = track
    muxer.nextTrackId++
    return track.trackId
}

func (muxer *Movmuxer) Write(track uint32, data []byte, pts uint64, dts uint64) error {
    mp4track := muxer.tracks[track]
    entry := sampleEntry{
        pts:                    pts,
        dts:                    dts,
        size:                   uint64(len(data)),
        SampleDescriptionIndex: 1,
        offset:                 uint64(muxer.writerHandler.Tell()),
    }
    mp4track.samplelist = append(mp4track.samplelist, entry)
    mp4track.duration += uint32(dts)
    if mp4track.cid == MOV_CODEC_H264 {
        return muxer.writeH264(mp4track, data)
    } else if mp4track.cid == MOV_CODEC_H265 {
        return muxer.writeH265(mp4track, data)
    } else if mp4track.cid == MOV_CODEC_AAC {
        return muxer.writeAAC(mp4track, data)
    } else if mp4track.cid == MOV_CODEC_G711A || mp4track.cid == MOV_CODEC_G711U {
        return muxer.writeG711(mp4track, data)
    } else {
        return errors.New("UnSupport Codec")
    }
}

func (muxer *Movmuxer) Writetrailer() (err error) {
    currentoffset := muxer.writerHandler.Tell()
    datalen := currentoffset - int64(muxer.mdatOffset)
    if datalen > 0xFFFFFFFF {
        MDAT.Size = uint64(datalen)
        len, mdata := MDAT.Encode()
        if _, err = muxer.writerHandler.Seek(int64(muxer.mdatOffset)-8, 0); err != nil {
            return err
        }
        if _, err = muxer.writerHandler.Write(mdata[0:len]); err != nil {
            return err
        }
        if _, err = muxer.writerHandler.Seek(currentoffset, 0); err != nil {
            return err
        }
    }

    mvhd := NewMovieHeaderBox()
    mvhd.Next_track_ID = muxer.nextTrackId
    _, moov := mvhd.Encode()
    moovsize := 0
    for _, track := range muxer.tracks {
        muxer.makeStblTable()

        if track.cid == MOV_CODEC_H264 || track.cid == MOV_CODEC_H265 {
            tkhd := NewTrackHeaderBox()
            tkhd.Track_ID = track.trackId
            tkhd.Duration = uint64(track.duration)
            tkhd.Width = muxer.width
            tkhd.Height = muxer.height
            s1, tkhdbytes := tkhd.Encode()

        }
    }

}

func (muxer *Movmuxer) writeH264(track *mp4track, h264 []byte) (err error) {
    h264extra, ok := track.extra.(*h264ExtraData)
    if !ok {
        panic("must init h264ExtraData first")
    }
    mpeg.SplitFrameWithStartCode(h264, func(nalu []byte) bool {
        nalu_type := mpeg.H264NaluType(nalu)
        switch nalu_type {
        case mpeg.H264_NAL_SPS:
            spsid := mpeg.GetSPSIdWithStartCode(nalu)
            for _, sps := range h264extra.spss {
                if spsid == mpeg.GetSPSId(sps) {
                    return true
                }
            }
            tmp := make([]byte, len(nalu))
            copy(tmp, nalu)
            h264extra.spss = append(h264extra.spss, tmp)
            if muxer.width == 0 || muxer.height == 0 {
                muxer.width, muxer.height = mpeg.GetH264Resolution(h264extra.spss[0])
                track.width = muxer.width
                track.height = muxer.height
            }
        case mpeg.H264_NAL_PPS:
            ppsid := mpeg.GetSPSIdWithStartCode(nalu)
            for _, pps := range h264extra.ppss {
                if ppsid == mpeg.GetPPSId(pps) {
                    return true
                }
            }
            tmp := make([]byte, len(nalu))
            copy(tmp, nalu)
            h264extra.ppss = append(h264extra.ppss, tmp)
        }
        avcc := mpeg.ConvertAnnexBToAVCC(nalu)
        if _, err = muxer.writerHandler.Write(avcc); err != nil {
            return false
        }
        return true
    })
    return
}

func (muxer *Movmuxer) writeH265(track *mp4track, h265 []byte) (err error) {
    h265extra, ok := track.extra.(*h265ExtraData)
    if !ok {
        panic("must init h265ExtraData first")
    }
    mpeg.SplitFrameWithStartCode(h265, func(nalu []byte) bool {
        nalu_type := mpeg.H265NaluType(nalu)
        switch nalu_type {
        case mpeg.H265_NAL_SPS:
            h265extra.hvccExtra.UpdateSPS(nalu)
            if muxer.width == 0 || muxer.height == 0 {
                muxer.width, muxer.height = mpeg.GetH265Resolution(nalu)
                track.width = muxer.width
                track.height = muxer.height
            }
        case mpeg.H265_NAL_PPS:
            h265extra.hvccExtra.UpdatePPS(nalu)
        case mpeg.H265_NAL_VPS:
            h265extra.hvccExtra.UpdateVPS(nalu)
        }
        avcc := mpeg.ConvertAnnexBToAVCC(nalu)
        if _, err = muxer.writerHandler.Write(avcc); err != nil {
            return false
        }
        return true
    })
    return
}

func (muxer *Movmuxer) writeAAC(track *mp4track, aacframes []byte) (err error) {
    aacextra, ok := track.extra.(*aacExtraData)
    if !ok {
        panic("must init aacExtraData first")
    }
    if aacextra.asc == nil || len(aacextra.asc) <= 0 {
        asc, err := mpeg.ConvertADTSToASC(aacframes)
        if err != nil {
            return err
        }
        copy(aacextra.asc, asc)
    }
    _, err = muxer.writerHandler.Write(aacframes[7:])
    return
}

func (muxer *Movmuxer) writeG711(track *mp4track, g711 []byte) (err error) {
    _, err = muxer.writerHandler.Write(g711)
    return
}

func (muxer *Movmuxer) makeStblTable() {
    for _, track := range muxer.tracks {
        if track.stbltable == nil {
            track.stbltable = new(movstbl)
        }
        sameSize := true
        stts := new(movstts)
        stts.entrys = make([]sttsEntry, 0)
        movchunks := make([]movchunk, 0)
        ctts := new(movctts)
        ctts.entrys = make([]cttsEntry, 0)
        ckn := uint32(1)
        for i, sample := range track.samplelist {
            sttsEntry := sttsEntry{sampleCount: 1, sampleDelta: 0}
            cttsEntry := cttsEntry{sampleCount: 1, sampleOffset: uint32(sample.pts) - uint32(sample.dts)}
            if i == len(track.samplelist)-1 {
                stts.entrys = append(stts.entrys, sttsEntry)
            } else {
                delta := track.samplelist[i+1].dts - sample.dts
                if len(stts.entrys) > 0 && delta == uint64(stts.entrys[len(stts.entrys)-1].sampleDelta) {
                    stts.entrys[len(stts.entrys)-1].sampleCount++
                } else {
                    sttsEntry.sampleDelta = uint32(delta)
                    stts.entrys = append(stts.entrys, sttsEntry)
                }
            }

            if len(ctts.entrys) == 0 {
                ctts.entrys = append(ctts.entrys, cttsEntry)
            } else {
                if ctts.entrys[len(ctts.entrys)-1].sampleOffset == cttsEntry.sampleOffset {
                    ctts.entrys[len(ctts.entrys)-1].sampleCount++
                } else {
                    ctts.entrys = append(ctts.entrys, cttsEntry)
                }
            }

            if sameSize && track.samplelist[i+1].size != track.samplelist[i].size {
                sameSize = false
            }
            if i > 0 && sample.offset == track.samplelist[i-1].offset+track.samplelist[i-1].size {
                movchunks[ckn-1].samplenum++
            } else {
                ck := movchunk{chunknum: ckn, samplenum: 1, chunkoffset: sample.offset}
                movchunks = append(movchunks, ck)
                ckn++
            }
        }
        stsz := &movstsz{
            sampleSize:  0,
            sampleCount: uint32(len(track.samplelist)),
        }
        if sameSize {
            stsz.sampleSize = uint32(track.samplelist[0].size)
        } else {
            stsz.entrySizelist = make([]uint32, stsz.sampleCount)
            for i := 0; i < len(stsz.entrySizelist); i++ {
                stsz.entrySizelist[i] = uint32(track.samplelist[i].size)
            }
        }

        stsc := &movstsc{
            entrys:     make([]stscEntry, len(movchunks)),
            entryCount: 0,
        }

        for i, chunk := range movchunks {
            if i == 0 || chunk.samplenum == movchunks[i-1].samplenum {
                stsc.entrys[stsc.entryCount].firstChunk = chunk.chunknum
                stsc.entrys[stsc.entryCount].sampleDescriptionIndex = 1
                stsc.entrys[stsc.entryCount].samplesPerChunk = chunk.samplenum
                stsc.entryCount++
            }
        }

        stco := &movstco{entryCount: ckn, chunkOffsetlist: make([]uint64, ckn)}
        for i := 0; i < int(stco.entryCount); i++ {
            stco.chunkOffsetlist[i] = movchunks[i].chunkoffset
        }

        track.stbltable.stts = stts
        track.stbltable.stsc = stsc
        track.stbltable.stco = stco
        track.stbltable.stsz = stsz
        track.stbltable.ctts = ctts
    }
}
