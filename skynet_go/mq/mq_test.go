package mq

// import (
// 	log "skp-go/skynet_go/logger"
// 	"testing"
// 	"time"
// )

// func Test_Mq1(t *testing.T) {
// 	log.SetLevel(log.Lerr)
// 	mqClient := NewClient(":5678")
// 	msg := &Msg{Topic: "Mq", Tag: "*"}
// 	num := 1
// 	mqClient.Send(msg, "Mq.Register", num)
// 	time.Sleep(3 * time.Second)
// }
