package mq

const (
	TypeSend int32 = iota
	TypeSendReq
	TypeCall
	TypeRespond
)

const (
	ClientStateInit  int32 = 0
	ClientStateStart int32 = 1 << iota //tag 才会发送
	ClientStateErr
	ClientStateStopSubscribe
	ClientStateStopping
	ClientStateStop
)
