package mp4

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
    cid        MOV_CODEC_TYPE
    stbltable  *movstbl
    samplelist []sampleEntry
}

func newmp4track(cid MOV_CODEC_TYPE) *mp4track {
    return &mp4track{
        cid:        cid,
        stbltable:  nil,
        samplelist: make([]sampleEntry, 0),
    }
}

type Movmuxer struct {
    writer Writer
    tracks []*mp4track
}

func (muxer *Movmuxer) AddAudioTrack(cid int, channelcount uint8, sampleBits uint8, sampleRate uint) int {
    return 1
}

func (muxer *Movmuxer) AddVideoTrack(cid int) int {
    return 0
}

func (muxer *Movmuxer) makeStblTable() {
    for _, track := range muxer.tracks {
        if track.stbltable == nil {
            track.stbltable = new(movstbl)
        }
        sameSize := true
        stts := new(movstts)
        stts.entrys = make([]sttsEntry, 1)
        movchunks := make([]movchunk, 0)
        ckn := uint32(1)
        for i, sample := range track.samplelist {
            sttsEntry := sttsEntry{sampleCount: 1, sampleDelta: 0}
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
    }
}
