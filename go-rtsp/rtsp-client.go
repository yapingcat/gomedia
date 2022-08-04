package rtsp

import "net/url"

type RtspClient struct {
	uri     string
	usrName string
	passwd  string
}

type ClientOption func(cli *RtspClient)

func NewRtspClient(uri string, opt ...ClientOption) (*RtspClient, error) {
	cli := &RtspClient{}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	cli.usrName = u.User.Username()
	if _, ok := u.User.Password(); ok {
		cli.passwd, _ = u.User.Password()
	}
	return cli, nil
}

func (client *RtspClient) RegisterCodec() {

}

func (client *RtspClient) OnTrack() {

}
