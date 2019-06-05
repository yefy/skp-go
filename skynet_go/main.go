package main

import (
	log "skp-go/skynet_go/logger"
)

type xxx interface {
	print()
}

type AAA struct {
	xxx
	aaa int
}

func (a *AAA) printxxx() {
	a.xxx.print()
}

func (a *AAA) print() {
	log.Fatal("aaa")
}

type BBB struct {
	AAA
	bbb int
}

func (b *BBB) print() {
	log.Fatal("bbb")
}

func main() {
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Lall, log.Lscreen|log.Lfile)
	log.All("main start")
	// b := &BBB{}
	// b.xxx = b
	// b.printxxx()
}
