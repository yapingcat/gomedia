package rtsp

import "github.com/yapingcat/gomedia/go-rtsp/sdp"

type ClientHandle interface {
    HandleOption(cli *RtspClient, res RtspResponse, public []string) error
    HandleDescribe(cli *RtspClient, res RtspResponse, sdp *sdp.Sdp, tracks map[string]*RtspTrack) error
    HandleSetup(cli *RtspClient, res RtspResponse, currentTrack *RtspTrack, tracks map[string]*RtspTrack, sessionId string, timeout int) error
    HandleAnnounce(cli *RtspClient, res RtspResponse) error
    HandlePlay(cli *RtspClient, res RtspResponse, timeRange *RangeTime, info *RtpInfo) error
    HandlePause(cli *RtspClient, res RtspResponse) error
    HandleTeardown(cli *RtspClient, res RtspResponse) error
    HandleGetParameter(cli *RtspClient, res RtspResponse) error
    HandleSetParameter(cli *RtspClient, res RtspResponse) error
    HandleRedirect(cli *RtspClient, req RtspRequest, location string, timeRange *RangeTime) error
    HandleRecord(cli *RtspClient, res RtspResponse, timeRange *RangeTime, info *RtpInfo) error
    HandleRequest(cli *RtspClient, req RtspRequest) error
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
