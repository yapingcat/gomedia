package sdp

import (
	"fmt"
	"testing"
)

var sdpstr string = `v=0
o=- 0 0 IN IP6 ::1
s=No Name
c=IN IP6 ::1
t=0 0
a=tool:libavformat 56.40.101
m=video 0 RTP/AVP 96
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z2QAH6zZQFAFuhAASRsQDqYAAPGDGWA=,aOvjyyLA; profile-level-id=64001F
a=control:streamid=0
m=audio 0 RTP/AVP 97
b=AS:34
a=rtpmap:97 MPEG4-GENERIC/16000/1
a=fmtp:97 profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3; config=1408
a=control:streamid=1`

func TestParserSdp(t *testing.T) {
	t.Run("parse_sdp", func(t *testing.T) {
		sdp := &Sdp{}
		err := sdp.ParserSdp(sdpstr)
		fmt.Println(err)
		fmt.Printf("%+v\n", sdp)
		fmt.Printf("%+v\n", sdp.Medias[0])
		fmt.Printf("%+v\n", sdp.Medias[1])
	})
}
