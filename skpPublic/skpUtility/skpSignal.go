package skpUtility

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type SkpSignal struct {
	sig chan os.Signal
}

func SkpNewSignal() (this *SkpSignal) {
	sig := new(SkpSignal)
	sig.sig = make(chan os.Signal)
	return sig
}

func (this *SkpSignal) SkpRegister() {
	signal.Notify(this.sig, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
}

func (this *SkpSignal) SkpWait() {
__loop:
	for {
		select {
		case s := <-this.sig:
			fmt.Printf("********sig = %v process exit \n", s)
			break __loop
		}
	}
}
