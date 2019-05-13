package skpSocket

import (
	"errors"
	"fmt"
	"net"

	//"runtime"
	"skp-go/skpPublic/skpProtocol"
	"skp-go/skpPublic/skpUtility"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	Data_Chace_Max           = 1024 * 16
	Timeout_Max              = 60 * 5
	Read_Max                 = 1024 * 4
	Read_Chan_Max            = 1000
	Write_Chan_Max           = 1000
	Close_Socket_Sec         = 2
	Close_Socket_System_Wait = 10
	Socket_Read_Buffer       = 1024 * 128
	Socket_Write_Buffer      = 1024 * 128
)

const (
	SocketClose = iota
	SocketError
)

const (
	IS_SENDER_CHAN = 1
)

type SkpTcpSocketBase interface {
	SkpStart()
	SkpData(data []byte) error
	SkpTimeout()
	SkpEnd()
}

type SkpTcpSocket struct {
	*SkpObject
	SkpTcpSocketBase

	Conn       *net.TCPConn
	Manage     *SkpManage
	TcpService *SkpTcpService

	ReadChan chan []byte

	WriteChan chan []byte

	Waitgroup sync.WaitGroup

	CloseChanMutex sync.Mutex
	CloseChan      chan bool
	IsCloseChan    bool

	//CloseMainMutex  sync.Mutex
	//CloseMainChan   chan int
	//IsCloseMain     bool
	//IsRealCloseMain bool
	//CloseTimeSec int
	//IsCloseing   bool

	WaitMonitorChan chan bool

	IsConn  bool
	IsOwner bool

	Datas []byte

	ProtocalConnHead skpProtocol.SkpProtocalConnHead
	ProtocalHead     skpProtocol.SkpProtocalHead

	AddrSocket uint64
}

func SkpNewTcpSocket(tcpSocketBase SkpTcpSocketBase, conn *net.TCPConn, tcpService *SkpTcpService) (this *SkpTcpSocket) {
	tcpSocket := new(SkpTcpSocket)
	tcpSocket.SkpObject = SkpNewObject()
	tcpSocket.SkpTcpSocketBase = tcpSocketBase
	tcpSocket.Conn = conn
	tcpSocket.Conn.SetLinger(Close_Socket_System_Wait)
	tcpSocket.Conn.SetReadBuffer(Socket_Read_Buffer)
	tcpSocket.Conn.SetWriteBuffer(Socket_Write_Buffer)
	tcpSocket.Conn.SetNoDelay(true)
	tcpSocket.Manage = tcpService.SkpGetManage()
	tcpSocket.TcpService = tcpService
	tcpSocket.ReadChan = make(chan []byte, Read_Chan_Max)
	tcpSocket.WriteChan = make(chan []byte, Write_Chan_Max)

	tcpSocket.CloseChan = make(chan bool, 5)
	tcpSocket.IsCloseChan = false

	tcpSocket.WaitMonitorChan = make(chan bool, 5)
	tcpSocket.IsConn = false
	tcpSocket.IsOwner = false
	tcpSocket.Datas = make([]byte, 0, Data_Chace_Max)

	tcpSocket.AddrSocket = uint64(uintptr(unsafe.Pointer(tcpSocket)))

	return tcpSocket
}

func (this *SkpTcpSocket) SkpAppendData(data []byte) {
	left := cap(this.Datas) - len(this.Datas)
	if left > len(data) {
		this.Datas = append(this.Datas, data...)
	} else {
		d := make([]byte, 0, len(this.Datas)+len(data)+Data_Chace_Max)
		d = append(d, this.Datas[0:len(this.Datas)]...)

		this.Datas = d[0:]
		this.Datas = append(this.Datas, data...)
	}
}

func (this *SkpTcpSocket) SkpGetData(size int) []byte {
	return this.Datas[:size]
}

func (this *SkpTcpSocket) SkpSkipData(skip int) {
	this.Datas = this.Datas[skip:]
}

func (this *SkpTcpSocket) SkpLoadConnHead(data []byte) error {
	err := skpProtocol.SkpCheckConnHead(data, &this.ProtocalConnHead)
	if err != nil {
		return err
	}

	this.IsConn = true

	ret := this.Manage.SkpAddTcpSocketToMap(this.ProtocalConnHead.FromServerID, this)
	if ret {
		return nil
	} else {
		return errors.New("add tcp socket error")
	}
}

func (this *SkpTcpSocket) SkpLoadHead(data []byte) error {
	err := skpProtocol.SkpCheckHead(data, &this.ProtocalHead)
	return err
}

func (this *SkpTcpSocket) SkpCloseChan() {
	this.CloseChanMutex.Lock()
	defer this.CloseChanMutex.Unlock()
	if this.IsCloseChan == true {
		return
	}

	this.IsCloseChan = true

	this.CloseChan <- true
}

func (this *SkpTcpSocket) SkpClose() {
	serverID := this.SkpServerID()
	this.Manage.SkpDelTcpSocketForMap(serverID, this)
}

func (this *SkpTcpSocket) SkpSetAddrSocket(data []byte) {

	if this.ProtocalConnHead.ConnType == skpProtocol.ConnTypeClient {
		proxyData, _ := skpProtocol.SkpEncode(this.Manage.ServerID)
		var proxy skpProtocol.SkpProxyID
		skpProtocol.SkpDecode(proxyData, &proxy)
		data[24] = proxy.Data1
		data[25] = proxy.Data2
		data[26] = proxy.Data3
		data[27] = proxy.Data4

		addrData, _ := skpProtocol.SkpEncode(this.AddrSocket)
		var addrSocket skpProtocol.SkpAddrSocket
		skpProtocol.SkpDecode(addrData, &addrSocket)

		data[32] = addrSocket.Data1
		data[33] = addrSocket.Data2
		data[34] = addrSocket.Data3
		data[35] = addrSocket.Data4
		data[36] = addrSocket.Data5
		data[37] = addrSocket.Data6
		data[38] = addrSocket.Data7
		data[39] = addrSocket.Data8
	}
}

func (this *SkpTcpSocket) SkpRun() {
	this.TcpService.Waitgroup.Add(1)
	defer this.TcpService.Waitgroup.Done()

	go this.SkpRead()

	var timeoutMSec int64 = 1000 * Timeout_Max
	updateTime := time.Now()

	timer := time.NewTimer(time.Duration(timeoutMSec) * time.Millisecond)

LOOP:
	for {
		select {
		case data, ok := <-this.ReadChan:
			if ok == false {
				this.SkpCloseChan()
				continue
			}

			if this.ProtocalConnHead.IsKeep == 1 {
				updateTime = time.Now()
			}

			if this.IsConn == false {
				err := this.SkpLoadConnHead(data)
				if err != nil {
					this.SkpCloseChan()
					continue
				}
			} else {
				err := this.SkpLoadHead(data)
				if err != nil {
					this.SkpCloseChan()
					continue
				}

				this.SkpSetAddrSocket(data)

				err = this.SkpData(data)

				if err != nil {
					this.SkpCloseChan()
					continue
				}
			}
		case data, ok := <-this.WriteChan:
			if ok == false {
				this.SkpCloseChan()
				continue
			}

			if this.ProtocalConnHead.IsKeep == 1 {
				updateTime = time.Now()
			}

			writeLen, err := this.Conn.Write(data)
			if err != nil || writeLen != len(data) {
				this.SkpCloseChan()
				continue
			}
		case _ = <-this.CloseChan:
			this.SkpClose()
			fmt.Printf("tcp socket %d close \n", this.AddrSocket)
			break LOOP
		case <-timer.C:
			currTime := time.Now()
			var diff time.Duration = currTime.Sub(updateTime)
			var diffMSec int64 = diff.Nanoseconds() / 1000 / 1000
			var registTimeoutMSec int64 = timeoutMSec

			if diffMSec >= timeoutMSec {
				fmt.Println("tcp socket timeout **********************")

				updateTime = time.Now()
				if this.ProtocalConnHead.IsKeep == 1 {
					this.SkpTimeout()
				}
				this.SkpCloseChan()
			} else {
				registTimeoutMSec = timeoutMSec - diffMSec
			}
			timer = time.NewTimer(time.Duration(registTimeoutMSec) * time.Millisecond)
		}
	}

	close(this.WriteChan)
	this.Conn.Close()

	this.Waitgroup.Wait()

	close(this.ReadChan)
	close(this.WaitMonitorChan)
	close(this.CloseChan)

	this.SkpEnd()

	//skpUtility.Printf("**************exit = %d \n", this.AddrSocket)
	fmt.Printf("**************exit = %d \n", this.AddrSocket)
}

func (this *SkpTcpSocket) SkpRead() {
	this.Waitgroup.Add(1)
	defer this.Waitgroup.Done()

	isConn := false
	if this.IsOwner {
		isConn = true
	}

	data := make([]byte, Read_Max)
	for {
		readLen, err := this.Conn.Read(data)

		if err != nil || readLen <= 0 {
			this.SkpCloseChan()
			return
		}

		this.SkpAppendData(data[:readLen])
		for {
			if isConn == false {
				var connHead skpProtocol.SkpProtocalConnHead
				err := skpProtocol.SkpCheckConnHead(this.Datas, &connHead)
				if err != nil {
					break
				}

				allSize := int(int(connHead.HeadSize) + int(connHead.DataSize))
				this.ReadChan <- this.SkpGetData(allSize)
				this.SkpSkipData(allSize)

				isConn = true
			}

			var head skpProtocol.SkpProtocalHead
			err := skpProtocol.SkpCheckHead(this.Datas, &head)
			if err != nil {
				break
			}

			allSize := int(int(head.HeadSize) + int(head.DataSize))
			this.ReadChan <- this.SkpGetData(allSize)
			this.SkpSkipData(allSize)
		}

	}
}

func (this *SkpTcpSocket) SkpWriteChan(data []byte) bool {
	this.CloseChanMutex.Lock()
	defer this.CloseChanMutex.Unlock()
	if this.IsCloseChan == true {
		return false
	}

	if IS_SENDER_CHAN == 1 {
		this.WriteChan <- data
	} else {
		writeLen, err := this.Conn.Write(data)
		if err != nil || writeLen != len(data) {
			this.SkpCloseChan()
			return false
		}
	}

	return true
}

func (this *SkpTcpSocket) SkpAddMonitorIPList(data []byte) {
	str := string(data[this.ProtocalHead.HeadSize:])

	skpUtility.Println(str)

	strs := strings.Split(str, ",")

	for _, v := range strs {
		if len(v) == 0 {
			break
		}

		vs := strings.Split(v, ":")
		serverID, _ := strconv.Atoi(vs[1])
		skpUtility.Printf("id = %d, v = %s \n", serverID, v)

		this.Manage.SkpAddServerAddr(uint32(serverID), v)
	}

	this.Manage.SkpSerializeAddr()

	this.WaitMonitorChan <- true
}

func (this *SkpTcpSocket) SkpServerID() uint64 {

	var serverID uint64
	if this.IsOwner {
		serverID = this.ProtocalConnHead.ToServerID
	} else {
		serverID = this.ProtocalConnHead.FromServerID
	}
	return serverID
}
