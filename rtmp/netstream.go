package rtmp

type NetStreamStatusCode string

func makePlay(transactionId int, streamName string, start float64, duration float64, reset bool) []byte {
	command := makeStringItem("play")
	tid := makeNumberItem(float64(transactionId))
	sName := makeStringItem(streamName)
	s := makeNumberItem(start)
	d := makeNumberItem(duration)
	r := makeBoolItem(reset)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sName.encode()...)
	msg = append(msg, s.encode()...)
	msg = append(msg, d.encode()...)
	msg = append(msg, r.encode()...)
	return msg
}

func makeLivePlay(transactionId int, streamName string) []byte {
	return makePlay(transactionId, streamName, -1, -1, true)
}

func makeDeleteStream(streamId int) []byte {
	command := makeStringItem("deleteStream")
	tid := makeNumberItem(0)
	sid := makeNumberItem(float64(streamId))

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sid.encode()...)
	return msg
}

func makeReceiveAudio(flag bool) []byte {
	command := makeStringItem("receiveAudio")
	tid := makeNumberItem(0)
	boolFlag := makeBoolItem(flag)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, boolFlag.encode()...)
	return msg
}

func makeReceiveVideo(flag bool) []byte {
	command := makeStringItem("receiveVideo")
	tid := makeNumberItem(0)
	boolFlag := makeBoolItem(flag)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, boolFlag.encode()...)
	return msg
}

func makePublish(pubName, pubType string) []byte {
	command := makeStringItem("publish")
	tid := makeNumberItem(0)
	publishName := makeStringItem(pubName)
	publishType := makeStringItem(pubType)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, publishName.encode()...)
	msg = append(msg, publishType.encode()...)
	return msg
}

func makeSeek(milliSeconds float64) []byte {
	command := makeStringItem("seek")
	tid := makeNumberItem(0)
	m := makeNumberItem(milliSeconds)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, m.encode()...)
	return msg
}

func makePause(pause bool, milliSeconds float64) []byte {
	command := makeStringItem("pause")
	tid := makeNumberItem(0)
	pauseFlag := makeBoolItem(pause)
	m := makeNumberItem(milliSeconds)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, pauseFlag.encode()...)
	msg = append(msg, m.encode()...)
	return msg
}

func makeReleaseStream(streamName string) []byte {
	command := makeStringItem("releaseStream")
	tid := makeNumberItem(0)
	sName := makeStringItem(streamName)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sName.encode()...)
	return msg
}

func makeFcPublish(streamName string) []byte {
	command := makeStringItem("FCPublish")
	tid := makeNumberItem(0)
	sName := makeStringItem(streamName)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sName.encode()...)
	return msg
}

func makeFcUnPublish(streamName string) []byte {
	command := makeStringItem("FCUnpublish")
	tid := makeNumberItem(0)
	sName := makeStringItem(streamName)

	msg := command.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, sName.encode()...)
	return msg
}

func makeStatusRes(transactionId int, code StatusCode, level StatusLevel, description string) []byte {
	commad := makeStringItem("onStatus")
	tid := makeNumberItem(float64(transactionId))
	des := amfObject{
		items: []*amfObjectItem{
			{name: "level", value: makeStringItem(string(level))},
			{name: "code", value: makeStringItem(string(code))},
			{name: "description", value: makeStringItem(description)},
		},
	}
	msg := commad.encode()
	msg = append(msg, tid.encode()...)
	msg = append(msg, NullItem...)
	msg = append(msg, des.encode()...)
	return msg
}
