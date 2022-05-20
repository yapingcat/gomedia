package rtmp

func makeConnect(app, tcurl string) []byte {
	command := makeStringItem("connect")
	transactionId := makeNumberItem(1)
	obj := amfObject{
		items: []*amfObjectItem{
			{name: "app", value: makeStringItem(app)},
			{name: "flashVer", value: makeStringItem("FMSc/1.0")},
			{name: "tcUrl", value: makeStringItem(tcurl)},
			{name: "fpad", value: makeBoolItem(false)},
			{name: "capabilities", value: makeNumberItem(15)},
			{name: "audioCodecs", value: makeNumberItem(4071)},
			{name: "videoCodecs", value: makeNumberItem(252)},
		},
	}
	msg := command.encode()
	msg = append(msg, transactionId.encode()...)
	msg = append(msg, obj.encode()...)
	return msg
}

func makeConnectRes() []byte {
	command := makeStringItem("_result")
	transactionId := makeNumberItem(1)
	properties := amfObject{
		items: []*amfObjectItem{
			{name: "fmsVer", value: makeStringItem("FMS/3,0,1,123")},
			{name: "capabilities", value: makeNumberItem(15)},
		},
	}
	information := amfObject{
		items: []*amfObjectItem{
			{name: "level", value: makeStringItem("status")},
			{name: "code", value: makeStringItem("NetConnection.Connect.Success")},
			{name: "description", value: makeStringItem("Connection Succeeded")},
			{name: "objectEncoding", value: makeNumberItem(0)},
		},
	}
	msg := command.encode()
	msg = append(msg, transactionId.encode()...)
	msg = append(msg, properties.encode()...)
	msg = append(msg, information.encode()...)
	return msg
}

func makeCreateStream(streamName string, tid int) []byte {
	command := makeStringItem("createStream")
	transactionId := makeNumberItem(float64(tid))
	msg := command.encode()
	msg = append(msg, transactionId.encode()...)
	msg = append(msg, NullItem...)
	return msg
}

func makeCreateStreamRes(transactionId uint32, streamId uint32) []byte {
	command := makeStringItem("_result")
	tid := makeNumberItem(float64(transactionId))
	sid := makeNumberItem(float64(streamId))
	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sid.encode()...)
	return msg
}

func makeGetStreamLength(transactionId int, streamName string) []byte {
	command := makeStringItem("getStreamLength")
	tid := makeNumberItem(float64(transactionId))
	stream := makeStringItem(streamName)
	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, stream.encode()...)
	return msg
}

func makeGetStreamLengthRes(transactionId int, duration float64) []byte {
	command := makeStringItem("_result")
	tid := makeNumberItem(float64(transactionId))
	d := makeNumberItem(duration)
	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, d.encode()...)
	return msg
}

func makeErrorRes(transactionId int, level, code, description string) []byte {
	command := makeStringItem("_error")
	tid := makeNumberItem(float64(transactionId))
	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	des := amfObject{
		items: []*amfObjectItem{
			{name: "level", value: makeStringItem(level)},
			{name: "code", value: makeStringItem(code)},
			{name: "description", value: makeStringItem(description)},
		},
	}
	msg = append(msg, des.encode()...)
	return msg
}
