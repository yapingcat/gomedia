package rtsp

import "github.com/yapingcat/gomedia/go-rtsp/sdp"

type ClientHandle interface {
    HandleOption(res RtspResponse, public []string) error
    HandleDescribe(res RtspResponse, sdp *sdp.Sdp, tracks map[string]*RtspTrack) error
    HandleSetup(res RtspResponse, tracks map[string]*RtspTrack, sessionId string, timeout int) error
    HandleAnnounce(res RtspResponse) error
    HandlePlay(res RtspResponse, timeRange *RangeTime, info *RtpInfo) error
    HandlePause(res RtspResponse) error
    HandleTeardown(res RtspResponse) error
    HandleGetParameter(res RtspResponse) error
    HandleSetParameter(res RtspResponse) error
    HandleRedirect(req RtspRequest, location string, timeRange *RangeTime) error
    HandleRecord(res RtspResponse, timeRange *RangeTime, info *RtpInfo) error
    HandleRequest(req RtspRequest) error
}

type ServerHandle interface {
    HandleOption(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleDescribe(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleSetup(svr *RtspServer, req RtspRequest, res *RtspResponse, transport *RtspTransport, tracks *RtspTrack)
    HandleAnnounce(svr *RtspServer, req RtspRequest, tracks map[string]*RtspTrack)
    HandlePlay(svr *RtspServer, req RtspRequest, res *RtspResponse, timeRange *RangeTime, info []*RtpInfo)
    HandlePause(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleTeardown(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleGetParameter(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleSetParameter(svr *RtspServer, req RtspRequest, res *RtspResponse)
    HandleRecord(svr *RtspServer, req RtspRequest, res *RtspResponse, timeRange *RangeTime, info []*RtpInfo)
    HandleResponse(svr *RtspServer, res RtspResponse)
}
