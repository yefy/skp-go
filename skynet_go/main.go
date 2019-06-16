package main

import (
	"skp-go/skynet_go/defaultConf"
	log "skp-go/skynet_go/logger"
)

func main() {
	defaultConf.SetDebug()
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Lall, log.Lscreen|log.Lfile)
	log.All("main start")
}
