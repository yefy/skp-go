package main

import (
	log "skp-go/skynet_go/logger"
)

func main() {
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Lall, log.Lscreen|log.Lfile)

	log.All("main start")

}
