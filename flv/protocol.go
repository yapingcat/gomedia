package flv

type FLVSAMPLEINDEX int

const (
    FLV_SAMPLE_5500 FLVSAMPLEINDEX = iota
    FLV_SAMPLE_11000
    FLV_SAMPLE_22000
    FLV_SAMPLE_44000
)

type TagType int

const (
    AUDIO_TAG  TagType = 8
    VIDEO_TAG  TagType = 9
    SCRIPT_TAG TagType = 18
)

type FLV_VIDEO_FRAME_TYPE int

const (
    KEY_FRAME FLV_VIDEO_FRAME_TYPE = 1
    INTER_FRAME
)

type FLV_VIDEO_CODEC_ID int

const (
    FLV_AVC  FLV_VIDEO_CODEC_ID = 7
    FLV_HEVC FLV_VIDEO_CODEC_ID = 12
)

const (
    AVC_SEQUENCE_HEADER = 0
    AVC_NALU            = 1
)

const (
    AAC_SEQUENCE_HEADER = 0
    AAC_RAW             = 1
)

type FLV_SOUND_FORMAT int

const (
    FLV_MP3   FLV_SOUND_FORMAT = 2
    FLV_G711A FLV_SOUND_FORMAT = 7
    FLV_G711U FLV_SOUND_FORMAT = 8
    FLV_AAC   FLV_SOUND_FORMAT = 10
)
