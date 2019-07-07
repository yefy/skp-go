package defaultConf

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc/rpcU"
)

func SetDebug() {
	log.SetPanic(true)
	errorCode.SetStack(true)
	rpcU.SetCheck(true)
}

func SetRelease() {
	log.SetPanic(false)
	errorCode.SetStack(true)
	rpcU.SetCheck(true)
}
