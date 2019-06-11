package mq

const (
	TypeSend int32 = iota
	TypeSendReq
	TypeCall
	TypeRespond
)

const (
	ConnStateInit  int32 = 0
	ConnStateStart int32 = 1 << iota //tag 才会发送
	ConnStateErr
	ConnStateStopSubscribe
	ConnStateStopping
	ConnStateStop
)
