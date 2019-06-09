package mq

const (
	typSend int32 = iota
	typSendReq
	typCall
	typRespond
)

const (
	encodeGob int32 = iota
	encodeProto
	encodeJson
)

const (
	connStateInit  int32 = 0
	connStateStart int32 = 1 << iota //tag 才会发送
	connStateErr
	connStateStopSubscribe
	connStateStopping
	connStateStop
)
