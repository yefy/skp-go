package mq

const (
	typSend = iota
	typSendReq
	typCall
	typRespond
)

const (
	encodeGob = iota
	encodeProto
	encodeJson
)
