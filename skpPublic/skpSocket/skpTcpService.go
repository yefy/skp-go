package skpSocket

import (
	"fmt"
	"net"
	"os"
	"runtime"
	//	"sync"
	//"sync/atomic"
	//	"time"
	//"bytes"
	//"encoding/binary"
	_ "bytes"
	_ "encoding/binary"
	//"github.com/davecheney/profile"
	"github.com/widuu/goini"
	//"skpPublic/skpProtocol"
	"skp/skpPublic/skpUtility"
	"strconv"
	"sync"
)

const (
	Accept_Max = 10
)

var CPU_Number int = 0

type SkpTcpServiceBase interface {
	SkpAcceptConn(conn *net.TCPConn)
}

type SkpTcpService struct {
	SkpTcpServiceBase
	Manage    *SkpManage
	Listener  *net.TCPListener
	Conf      *goini.Config
	Waitgroup sync.WaitGroup
}

func SkpNewTcpService(tcpServiceBase SkpTcpServiceBase) (this *SkpTcpService) {
	tcpService := new(SkpTcpService)
	tcpService.SkpTcpServiceBase = tcpServiceBase
	tcpService.Conf = goini.SetConfig("./conf.ini")
	return tcpService
}

func (this *SkpTcpService) SkpListen(addr *net.TCPAddr) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", addr)
	this.Listener = listener
	return listener, err
}

func (this *SkpTcpService) SkpService() {
	//cp /tmp/profile930080374/cpu.pprof ./
	//go tool pprof --pdf go-skp-server cpu.pprof > cpu-report.pdf

	//cp /tmp/profile664139879/mem.pprof ./
	//go tool pprof --pdf go-skp-server mem.pprof > mem-report.pdf

	//defer profile.Start(profile.CPUProfile).Stop()
	//defer profile.Start(profile.MemProfile).Stop()

	sig := skpUtility.SkpNewSignal()
	sig.SkpRegister()

	if CPU_Number == 0 {
		CPU_Number = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(CPU_Number)
	fmt.Printf("cpu number = %d\n", CPU_Number)

	port := this.Conf.GetValue("server", "port")

	addr := ":" + port
	fmt.Printf("listen addr %v \n", addr)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		skpUtility.Printf("create addr %v error %v \n", addr, err)
		os.Exit(1)
	}

	listener, err := this.SkpListen(tcpAddr)

	if err != nil {
		skpUtility.Printf("listen addr %v error %v \n", addr, err)
		os.Exit(1)
	}

	for i := 0; i < Accept_Max; i++ {
		go this.SkpAccept()
	}

	sig.SkpWait()

	this.Manage.SkpListerClose()
	listener.Close()

	this.Manage.SkpCloseAllTcpSocket()

	this.Waitgroup.Wait()

	fmt.Println("*******service exit")
}

func (this *SkpTcpService) SkpAccept() {
	this.Waitgroup.Add(1)
	defer this.Waitgroup.Done()

	for {
		conn, err := this.Listener.AcceptTCP()
		if err != nil {
			//skpUtility.Printf("accept client error\n")
			break
		}

		this.SkpAcceptConn(conn)
	}
	skpUtility.Printf("tcp service close\n")
}

func (this *SkpTcpService) SkpLoadConf(serverType uint8) {

	serverIDStr := this.Conf.GetValue("server", "port")
	serverID, _ := strconv.Atoi(serverIDStr)
	monitorServerIDStr := this.Conf.GetValue("server", "monitorPort")
	monitorServerID, _ := strconv.Atoi(monitorServerIDStr)
	monitorServerIP := this.Conf.GetValue("server", "monitorIP")

	this.Manage.SkpSetServerType(serverType)
	this.Manage.SkpSetServerID(uint32(serverID))

	this.Manage.MonitorServerID = uint32(monitorServerID)
	this.Manage.MonitorServerIP = monitorServerIP

	this.Manage.SkpAddServerIPPort(uint32(monitorServerID), monitorServerIP, uint16(monitorServerID))

	this.Manage.SkpGetMonitorIPList()
}

func (this *SkpTcpService) SkpGetManage() *SkpManage {
	return this.Manage
}
