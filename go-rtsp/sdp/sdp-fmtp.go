package sdp

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type CodecParamHandler interface {
	Decode(string) error
	Encode() string
}

func CreateParameterHandler(codecName string) CodecParamHandler {
	switch codecName {
	case "h264", "H264":
		return new(FmtpH264)
	case "h265", "H265":
		return new(FmtpH265)
	}
	return nil
}

type FmtpH264 struct {
	PayloadType       int
	PacketizationMode int
	Sps               []byte //without startcode
	Pps               []byte
}

func (fmtp *FmtpH264) Decode(fmtpAttr string) error {
	items := strings.SplitN(fmtpAttr, " ", 2)
	fmtp.PayloadType, _ = strconv.Atoi(items[0])
	if len(items) < 2 {
		return nil
	}

	params := strings.Split(items[1], ";")
	for _, param := range params {
		kv := strings.Split(strings.TrimSpace(param), "=")
		if len(kv) < 2 {
			continue
		}
		switch kv[0] {
		case "packetization-mode":
			fmtp.PacketizationMode, _ = strconv.Atoi(kv[1])
		case "sprop-parameter-sets":
			spspps := strings.Split(kv[1], ",")
			fmtp.Sps, _ = base64.StdEncoding.DecodeString(spspps[0])
			fmtp.Pps, _ = base64.StdEncoding.DecodeString(spspps[1])
		case "profile-level-id":
		}
	}
	return nil
}

func (fmtp *FmtpH264) Encode() string {
	return fmt.Sprintf("a=fmtp:%d packetization-mode=%d; sprop-parameter-sets=%s,%s; profile-level-id=%s", fmtp.PayloadType, fmtp.PacketizationMode,
		base64.StdEncoding.EncodeToString(fmtp.Sps), base64.StdEncoding.EncodeToString(fmtp.Pps), strings.ToUpper(hex.EncodeToString(fmtp.Sps[1:4])))
}

type FmtpH265 struct {
	PayloadType int
	Vps         []byte
	Sps         []byte
	Pps         []byte
}

func (fmtp *FmtpH265) Decode(fmtpAttr string) error {
	items := strings.SplitN(fmtpAttr, " ", 2)
	fmtp.PayloadType, _ = strconv.Atoi(items[0])
	if len(items) < 2 {
		return nil
	}

	params := strings.Split(items[1], ";")
	for _, param := range params {
		kv := strings.Split(strings.TrimSpace(param), "=")
		if len(kv) < 2 {
			continue
		}
		switch kv[0] {
		case "sprop-vps":
			fmtp.Vps, _ = base64.StdEncoding.DecodeString(kv[1])
		case "sprop-sps":
			fmtp.Sps, _ = base64.StdEncoding.DecodeString(kv[1])
		case "sprop-pps":
			fmtp.Pps, _ = base64.StdEncoding.DecodeString(kv[1])
		}
	}
	return nil
}

func (fmtp *FmtpH265) Encode() string {
	return fmt.Sprintf("a=fmtp:%d ;sprop-vps=%s;sprop-sps=%s;sprop-pps=%s", fmtp.PayloadType,
		strings.ToUpper(base64.StdEncoding.EncodeToString(fmtp.Vps)),
		strings.ToUpper(base64.StdEncoding.EncodeToString(fmtp.Sps)),
		strings.ToUpper(base64.StdEncoding.EncodeToString(fmtp.Pps)))
}
