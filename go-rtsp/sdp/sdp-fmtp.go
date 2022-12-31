package sdp

import (
    "encoding/base64"
    "encoding/hex"
    "fmt"
    "strconv"
    "strings"

    "github.com/yapingcat/gomedia/go-codec"
)

type FmtpCodecParamParser interface {
    Load(fmtp string)
    Save() string
}

func CreateFmtpParamParser(name string) FmtpCodecParamParser {
    tmp := strings.ToLower(name)
    switch tmp {
    case "h264":
        return NewH264FmtpParam()
    case "h265":
        return NewH265FmtpParam()
    case "mpeg4-generic":
        return NewAACFmtpParam()
    }
    return nil
}

type H264ExtraOption func(param *H264FmtpParam)

type H264FmtpParam struct {
    packetizationMode int
    profileLevelId    []byte
    sps               []byte
    pps               []byte
}

func WithPacketizationMode(mode int) H264ExtraOption {
    return func(param *H264FmtpParam) {
        param.packetizationMode = mode
    }
}

func WithProfileLevelId(profileLevel []byte) H264ExtraOption {
    return func(param *H264FmtpParam) {
        param.profileLevelId = make([]byte, len(profileLevel))
        copy(param.profileLevelId, profileLevel)
    }
}

func WithH264SPS(sps []byte) H264ExtraOption {
    return func(param *H264FmtpParam) {
        idx, sc := codec.FindStartCode(sps, 0)
        if idx == -1 {
            param.sps = make([]byte, len(sps))
            copy(param.sps, sps)
        } else {
            param.sps = make([]byte, len(sps)-idx-int(sc))
            copy(param.sps, sps[idx+int(sc):])
        }
    }
}

func WithH264PPS(pps []byte) H264ExtraOption {
    return func(param *H264FmtpParam) {
        idx, sc := codec.FindStartCode(pps, 0)
        if idx == -1 {
            param.pps = make([]byte, len(pps))
            copy(param.pps, pps)
        } else {
            param.pps = make([]byte, len(pps)-idx-int(sc))
            copy(param.pps, pps[idx+int(sc):])
        }
    }
}

// a=fmtp:98 profile-level-id=42A01E;
//           packetization-mode=1;
//           sprop-parameter-sets=<parameter sets data>
func NewH264FmtpParam(opt ...H264ExtraOption) *H264FmtpParam {
    param := &H264FmtpParam{packetizationMode: 1}
    for _, o := range opt {
        o(param)
    }
    return param
}

func (param *H264FmtpParam) GetSpsPps() ([]byte, []byte) {
    return param.sps, param.pps
}

func (param *H264FmtpParam) Load(fmtp string) {
    items := strings.SplitN(fmtp, " ", 2)
    if len(items) < 2 {
        return
    }

    codecParam := strings.Split(items[1], ";")
    for _, p := range codecParam {
        kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
        if len(kv) < 2 {
            continue
        }
        switch kv[0] {
        case "packetization-mode":
            param.packetizationMode, _ = strconv.Atoi(kv[1])
        case "sprop-parameter-sets":
            spspps := strings.Split(kv[1], ",")
            param.sps, _ = base64.StdEncoding.DecodeString(spspps[0])
            param.pps, _ = base64.StdEncoding.DecodeString(spspps[1])
        case "profile-level-id":
            param.profileLevelId = make([]byte, 3)
            fmt.Sscanf(kv[1], "%02x%02x%02x", &param.profileLevelId[0], &param.profileLevelId[1], &param.profileLevelId[2])
        }
    }
}

func (param *H264FmtpParam) Save() string {
    paramStr := ""
    if len(param.profileLevelId) > 0 {
        paramStr += fmt.Sprintf("profile-level-id=%02x%02x%02x;", param.profileLevelId[0], param.profileLevelId[1], param.profileLevelId[2])
    }
    paramStr += fmt.Sprintf("packetization-mode=%d", param.packetizationMode)
    if len(param.sps) > 0 && len(param.pps) > 0 {
        paramStr += fmt.Sprintf(";sprop-parameter-sets=%s,%s", base64.StdEncoding.EncodeToString(param.sps), base64.StdEncoding.EncodeToString(param.pps))
    }
    return paramStr
}

type H265FmtpParam struct {
    sps []byte
    pps []byte
    vps []byte
}
type H265FmtpPramOption func(extra *H265FmtpParam)

func WithH265SPS(sps []byte) H265FmtpPramOption {
    return func(extra *H265FmtpParam) {
        idx, sc := codec.FindStartCode(sps, 0)
        if idx == -1 {
            extra.sps = make([]byte, len(sps))
            copy(extra.sps, sps)
        } else {
            extra.sps = make([]byte, len(sps)-idx-int(sc))
            copy(extra.sps, sps[idx+int(sc):])
        }
    }
}

func WithH265PPS(pps []byte) H265FmtpPramOption {
    return func(extra *H265FmtpParam) {
        idx, sc := codec.FindStartCode(pps, 0)
        if idx == -1 {
            extra.pps = make([]byte, len(pps))
            copy(extra.pps, pps)
        } else {
            extra.pps = make([]byte, len(pps)-idx-int(sc))
            copy(extra.pps, pps[idx+int(sc):])
        }
    }
}

func WithH265VPS(vps []byte) H265FmtpPramOption {
    return func(extra *H265FmtpParam) {
        idx, sc := codec.FindStartCode(vps, 0)
        if idx == -1 {
            extra.vps = make([]byte, len(vps))
            copy(extra.vps, vps)
        } else {
            extra.vps = make([]byte, len(vps)-idx-int(sc))
            copy(extra.vps, vps[idx+int(sc):])
        }
    }
}

//a=fmtp:96 sprop-vps=QAEMAfAIAAAAMAAAMAAAMAAAMAALUCQA==;sprop-sps=QgEBAIAAAAMAAAMAAAMAAAMAAKACgIAtH+W1kkbQzkkktySqSfKSyA==;sprop-pps=RAHBpVgeSA==
func NewH265FmtpParam(opt ...H265FmtpPramOption) *H265FmtpParam {
    param := &H265FmtpParam{}
    for _, o := range opt {
        o(param)
    }
    return param
}

func (param *H265FmtpParam) GetVpsSpsPps() ([]byte, []byte, []byte) {
    return param.vps, param.sps, param.pps
}

func (param *H265FmtpParam) Load(fmtp string) {
    items := strings.SplitN(fmtp, " ", 2)
    if len(items) < 2 {
        return
    }

    codecParams := strings.Split(items[1], ";")
    for _, p := range codecParams {
        kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
        if len(kv) < 2 {
            continue
        }
        switch kv[0] {
        case "sprop-vps":
            param.vps, _ = base64.StdEncoding.DecodeString(kv[1])
        case "sprop-sps":
            param.sps, _ = base64.StdEncoding.DecodeString(kv[1])
        case "sprop-pps":
            param.pps, _ = base64.StdEncoding.DecodeString(kv[1])
        }
    }
}

func (param *H265FmtpParam) Save() string {
    if len(param.pps) == 0 || len(param.vps) == 0 || len(param.sps) == 0 {
        return ""
    }
    return fmt.Sprintf("sprop-vps=%s; sprop-sps=%s; sprop-pps=%s", base64.StdEncoding.EncodeToString(param.vps),
        base64.StdEncoding.EncodeToString(param.sps), base64.StdEncoding.EncodeToString(param.pps))
}

// m=audio 49230 RTP/AVP 96
// a=rtpmap:96 mpeg4-generic/48000/6
// a=fmtp:96 streamtype=5; profile-level-id=16; mode=AAC-hbr;
// config=11B0; sizeLength=13; indexLength=3;indexDeltaLength=3
type AACFmtpParam struct {
    asc              []byte
    profileLevelId   int
    mode             string
    sizeLength       int
    indexLength      int
    indexDeltaLength int
}

type AACFmtpParamOption func(extra *AACFmtpParam)

func WithAudioSpecificConfig(asc []byte) AACFmtpParamOption {
    return func(extra *AACFmtpParam) {
        if len(asc) < 2 {
            panic("length of asc must >= 2 bytes")
        }
        extra.asc = make([]byte, len(asc))
        copy(extra.asc, asc)
    }
}

func NewAACFmtpParam(opt ...AACFmtpParamOption) *AACFmtpParam {
    param := &AACFmtpParam{
        mode:             "AAC-hbr",
        sizeLength:       13,
        indexLength:      3,
        indexDeltaLength: 3,
    }
    for _, o := range opt {
        o(param)
    }
    return param
}

func (param *AACFmtpParam) SizeLength() int {
    return param.sizeLength
}

func (param *AACFmtpParam) IndexLength() int {
    return param.indexLength
}

func (param *AACFmtpParam) IndexDeltaLength() int {
    return param.indexDeltaLength
}

func (param *AACFmtpParam) AudioSpecificConfig() []byte {
    return param.asc
}

func (param *AACFmtpParam) Load(fmtp string) {
    items := strings.SplitN(fmtp, " ", 2)
    if len(items) < 2 {
        return
    }

    codecParams := strings.Split(items[1], ";")
    for _, p := range codecParams {
        kv := strings.Split(strings.TrimSpace(p), "=")
        if len(kv) < 2 {
            continue
        }
        switch kv[0] {
        case "profile-level-id":
            param.profileLevelId, _ = strconv.Atoi(kv[1])
        case "mode":
            param.mode = kv[1]
        case "config":
            param.asc, _ = hex.DecodeString(kv[1])
        case "sizeLength":
            param.sizeLength, _ = strconv.Atoi(kv[1])
        case "indexLength":
            param.indexLength, _ = strconv.Atoi(kv[1])
        case "indexDeltaLength":
            param.indexDeltaLength, _ = strconv.Atoi(kv[1])
        }
    }

}

func (param *AACFmtpParam) Save() string {

    paramstr := fmt.Sprintf("streamtype=5;profile-level-id=%d;mode=%s;sizeLength=%d;indexLength=%d;indexDeltaLength=%d",
        param.profileLevelId, param.mode, param.sizeLength, param.indexLength, param.indexDeltaLength)
    if len(param.asc) > 0 {
        paramstr += ";config=" + hex.EncodeToString(param.asc)
    }
    return paramstr
}
