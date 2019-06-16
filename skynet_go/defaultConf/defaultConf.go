package defaultConf

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
)

func SetDebug() {
	log.SetPanic(true)
	errorCode.SetStack(false)
}

func SetRelease() {
	log.SetPanic(false)
	errorCode.SetStack(true)
}