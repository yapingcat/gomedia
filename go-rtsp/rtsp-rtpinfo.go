package rtsp

import (
    "strconv"
    "strings"
)

type RtpInfo struct {
    Url     string
    Seq     uint16
    Rtptime int64
}

func NewRtpInfo(url string, seq uint16) *RtpInfo {
    return &RtpInfo{Url: url, Seq: seq}
}

func (info *RtpInfo) EncodeString() string {
    str := "url=" + info.Url + ";" + strconv.Itoa(int(info.Seq))
    if info.Rtptime < 0 {
        str += ";rtptime=" + strconv.Itoa(int(info.Rtptime))
    }
    return str
}

func (info *RtpInfo) Decode(str string) {
    items := strings.Split(str, ";")
    for _, item := range items {
        kv := strings.Split(item, "=")
        switch kv[0] {
        case "url":
            info.Url = kv[1]
        case "seq":
            seq, _ := strconv.Atoi(kv[1])
            info.Seq = uint16(seq)
        case "rtptime":
            t, _ := strconv.Atoi(kv[1])
            info.Rtptime = int64(t)
        }
    }
}
