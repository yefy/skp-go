package mq

type RegisteRequest struct {
	Instance string
	Harbor   int32
	Topic    string
	Tag      string
}

type RegisterReply struct {
	Harbor int32
}

type StopSubscribeRequest struct {
}

type StopSubscribeReply struct {
}
