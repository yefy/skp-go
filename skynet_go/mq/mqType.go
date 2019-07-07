package mq

const (
	TypeRespond int32 = iota
	TypeSend
	TypeSendReq
	TypeCall
)

const (
	ClientStateInit  int32 = 0
	ClientStateStart int32 = 1 << iota //tag 才会发送
	ClientStateErr
	ClientStateStopSubscribe
	ClientStateStopping
	ClientStateStop
)

var OnMqRegister string = "OnMqRegister"
var OnMqStopSubscribe string = "OnMqStopSubscribe"
var OnMqClosing string = "OnMqClosing" //通知client退出
var OnMqClose string = "OnMqClose"
