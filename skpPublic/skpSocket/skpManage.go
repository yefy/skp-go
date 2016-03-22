package skpSocket

import (
	//"errors"
	//"fmt"
	"fmt"
	"math/rand"
	"net"
	"skp/skpPublic/skpProtocol"
	"skp/skpPublic/skpUtility"
	"strconv"
	"sync"
	"time"
)

var Socket_List_Max int = 50
var Manage_Agent_List_Max int = 100

//===============================================================
type SkpTcpSocketAgent struct {
	TcpSockets      []*SkpTcpSocket
	TcpSocketsMutex sync.Mutex
	Rand            *rand.Rand
	IsListerClose   bool
}

func SkpNewTcpSocketAgent() (this *SkpTcpSocketAgent) {
	tcpSocketAgent := new(SkpTcpSocketAgent)
	tcpSocketAgent.TcpSockets = make([]*SkpTcpSocket, 0, Socket_List_Max)
	tcpSocketAgent.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	tcpSocketAgent.IsListerClose = false
	return tcpSocketAgent
}

func (this *SkpTcpSocketAgent) SkpTcpSocketsLock() {
	this.TcpSocketsMutex.Lock()
}

func (this *SkpTcpSocketAgent) SkpTcpSocketsUnlock() {
	this.TcpSocketsMutex.Unlock()
}

func (this *SkpTcpSocketAgent) SkpLen() int {
	this.TcpSocketsMutex.Lock()
	defer this.TcpSocketsMutex.Unlock()

	return len(this.TcpSockets)
}

func (this *SkpTcpSocketAgent) SkpListerClose() {
	this.TcpSocketsMutex.Lock()
	defer this.TcpSocketsMutex.Unlock()
	this.IsListerClose = true
}

func (this *SkpTcpSocketAgent) SkpAddTcpSocket(conn *SkpTcpSocket) bool {
	this.TcpSocketsMutex.Lock()
	defer this.TcpSocketsMutex.Unlock()

	if this.IsListerClose {
		return false
	}

	this.TcpSockets = append(this.TcpSockets, conn)

	return true
}

func (this *SkpTcpSocketAgent) SkpDelTcpSocket(conn *SkpTcpSocket) {
	this.TcpSocketsMutex.Lock()
	defer this.TcpSocketsMutex.Unlock()

	if len(this.TcpSockets) == 0 {
		skpUtility.Printf("SkpDeleteMap list nil \n")
		return
	}
	var index int
	for index = 0; index < len(this.TcpSockets); index++ {
		c := this.TcpSockets[index]
		if c == conn {
			break
		}
	}

	if index == len(this.TcpSockets) {
		skpUtility.Printf("SkpDeleteMap list not find \n")
		return
	}

	this.TcpSockets = skpUtility.SkpRemoveSliceElement(interface{}(this.TcpSockets), index).([]*SkpTcpSocket)
	fmt.Printf("SkpTcpSocketAgent remove %d \n", conn.AddrSocket)
}

func (this *SkpTcpSocketAgent) SkpGetTcpSocket(addrSocket uint64) *SkpTcpSocket {
	this.TcpSocketsMutex.Lock()
	defer this.TcpSocketsMutex.Unlock()

	if len(this.TcpSockets) == 0 {
		return nil
	}

	var tcpSocket *SkpTcpSocket

	if addrSocket == 0 {
		rangeNumber := this.Rand.Intn(len(this.TcpSockets))
		tcpSocket = this.TcpSockets[rangeNumber]
		return tcpSocket
	}

	for _, tcpSocket = range this.TcpSockets {
		if tcpSocket.AddrSocket == addrSocket {
			return tcpSocket
		}
	}

	tcpSocket = this.TcpSockets[0]
	return tcpSocket
}

func (this *SkpTcpSocketAgent) SkpCloseAllTcpSocket() {
	this.TcpSocketsMutex.Lock()
	tcpSockets := make([]*SkpTcpSocket, len(this.TcpSockets))
	copy(tcpSockets, this.TcpSockets)
	this.TcpSockets = this.TcpSockets[0:0]
	this.TcpSocketsMutex.Unlock()

	for _, tcpSocket := range tcpSockets {
		fmt.Printf("SkpTcpSocketAgent remove %d \n", tcpSocket.AddrSocket)
		tcpSocket.SkpCloseChan()
	}
}

//===============================================================
type SkpManageAgent struct {
	Parent        *SkpManageAgent
	ServerIDMap   map[uint64]*SkpTcpSocketAgent
	ServerIDMutex sync.Mutex
}

func SkpNewManageAgent(parent *SkpManageAgent) (this *SkpManageAgent) {
	manageAgent := new(SkpManageAgent)
	manageAgent.Parent = parent
	manageAgent.ServerIDMap = make(map[uint64]*SkpTcpSocketAgent)

	return manageAgent
}

func (this *SkpManageAgent) SkpGetTcpSocketAgentForMap(serverID uint64) *SkpTcpSocketAgent {
	this.ServerIDMutex.Lock()
	defer this.ServerIDMutex.Unlock()

	tcpSocketAgent, ok := this.ServerIDMap[serverID]

	if ok {
		return tcpSocketAgent
	}

	if this.Parent != nil {
		tcpSocketAgent = this.Parent.SkpGetTcpSocketAgentForMap(serverID)
	} else {
		tcpSocketAgent = SkpNewTcpSocketAgent()
	}

	this.ServerIDMap[serverID] = tcpSocketAgent
	return tcpSocketAgent
}

func (this *SkpManageAgent) SkpAddTcpSocketToMap(serverID uint64, conn *SkpTcpSocket) bool {
	tcpSocketAgent := this.SkpGetTcpSocketAgentForMap(serverID)
	return tcpSocketAgent.SkpAddTcpSocket(conn)
}

func (this *SkpManageAgent) SkpDelTcpSocketForMap(serverID uint64, conn *SkpTcpSocket) {
	tcpSocketAgent := this.SkpGetTcpSocketAgentForMap(serverID)
	tcpSocketAgent.SkpDelTcpSocket(conn)
}

func (this *SkpManageAgent) SkpGetTcpSocketForMap(serverID uint64, addrSocket uint64) *SkpTcpSocket {
	tcpSocketAgent := this.SkpGetTcpSocketAgentForMap(serverID)
	return tcpSocketAgent.SkpGetTcpSocket(addrSocket)
}

func (this *SkpManageAgent) SkpGetTcpSocketMaxForMap(serverID uint64, addrSocket uint64) *SkpTcpSocket {
	tcpSocketAgent := this.SkpGetTcpSocketAgentForMap(serverID)

	if tcpSocketAgent.SkpLen() < Socket_List_Max {
		return nil
	}

	return tcpSocketAgent.SkpGetTcpSocket(addrSocket)
}

func (this *SkpManageAgent) SkpCloseAllTcpSocket() {
	this.ServerIDMutex.Lock()
	defer this.ServerIDMutex.Unlock()

	for _, v := range this.ServerIDMap {
		v.SkpCloseAllTcpSocket()
	}
}

func (this *SkpManageAgent) SkpListerClose() {
	this.ServerIDMutex.Lock()
	defer this.ServerIDMutex.Unlock()

	for _, v := range this.ServerIDMap {
		v.SkpListerClose()
	}
}

//===============================================================

type SkpManageBase interface {
	SkpNewTcpSocket(conn *net.TCPConn, connType uint8) (*SkpTcpSocket, error)
}

type SkpManage struct {
	SkpManageBase
	ServerAddrMap   map[uint32]string
	ServerAddrs     []byte
	ManageAgent     *SkpManageAgent
	ManageAgents    []*SkpManageAgent
	ServerType      uint8
	ServerID        uint32
	MonitorServerID uint32
	MonitorServerIP string
}

func SkpNewManage(manageBase SkpManageBase) (this *SkpManage) {
	manage := new(SkpManage)
	manage.ServerAddrMap = make(map[uint32]string)
	manage.ServerAddrs = make([]byte, 0, 1024)
	manage.SkpManageBase = manageBase

	manage.ManageAgent = SkpNewManageAgent(nil)
	for i := 0; i < Manage_Agent_List_Max; i++ {
		manageAgent := SkpNewManageAgent(manage.ManageAgent)
		manage.ManageAgents = append(manage.ManageAgents, manageAgent)
	}

	manage.ServerType = 0
	manage.ServerID = 0
	manage.MonitorServerID = 0
	manage.MonitorServerIP = ""
	return manage
}

func (this *SkpManage) SkpAddServerAddr(serverID uint32, addr string) {
	skpUtility.Printf("map add serverID = %d, addr = %s \n", serverID, addr)

	this.ServerAddrMap[serverID] = addr
}

func (this *SkpManage) SkpAddServerIPPort(serverID uint32, ip string, port uint16) {
	addr := ip
	addr = addr + ":"
	addr = addr + strconv.Itoa(int(port))
	this.SkpAddServerAddr(serverID, addr)
}

func (this *SkpManage) SkpGetServerAddr(serverID uint32) string {

	addr := this.ServerAddrMap[serverID]
	skpUtility.Printf("map get serverID = %d, addr = %s \n", serverID, addr)
	return addr
}

func (this *SkpManage) SkpGetServerTcpAddr(serverID uint32) *net.TCPAddr {

	addr := this.ServerAddrMap[serverID]
	skpUtility.Printf("map get serverID = %d, addr = %s \n", serverID, addr)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", addr)
	return tcpAddr
}

func (this *SkpManage) SkpSerializeAddr() {

	if len(this.ServerAddrs) != 0 {
		return
	}

	// 遍历map
	for _, v := range this.ServerAddrMap {
		this.ServerAddrs = append(this.ServerAddrs, []byte(v)...)
		this.ServerAddrs = append(this.ServerAddrs, []byte(",")...)
	}
}

func (this *SkpManage) SkpGetServerAllAddr(addrs *[]byte) {
	this.SkpSerializeAddr()
	*addrs = append(*addrs, this.ServerAddrs...)
}

/*  not supper SkpDeleteServerAddr
func (this *SkpManage) SkpDeleteServerAddr(serverID uint32) {
	skpUtility.Printf("map delete serverID = %d \n", serverID)

	delete(this.serverAddrMap, serverID)

}
*/

func (this *SkpManage) SkpSetServerID(serverID uint32) {
	this.ServerID = serverID
}

func (this *SkpManage) SkpSetServerType(serverType uint8) {
	this.ServerType = serverType
}

func (this *SkpManage) SkpGetManageAgent(serverID uint64) *SkpManageAgent {
	return this.ManageAgents[(serverID % uint64(len(this.ManageAgents)))]
}

func (this *SkpManage) SkpListerClose() {
	this.ManageAgent.SkpListerClose()
}
func (this *SkpManage) SkpAddTcpSocketToMap(serverID uint64, conn *SkpTcpSocket) bool {
	return this.SkpGetManageAgent(serverID).SkpAddTcpSocketToMap(serverID, conn)
}

func (this *SkpManage) SkpDelTcpSocketForMap(serverID uint64, conn *SkpTcpSocket) {
	this.SkpGetManageAgent(serverID).SkpDelTcpSocketForMap(serverID, conn)
}

func (this *SkpManage) SkpGetTcpSocketForMap(serverID uint64, addrSocket uint64) *SkpTcpSocket {
	return this.SkpGetManageAgent(serverID).SkpGetTcpSocketForMap(serverID, addrSocket)
}

func (this *SkpManage) SkpGetTcpSocketMaxForMap(serverID uint64, addrSocket uint64) *SkpTcpSocket {
	return this.SkpGetManageAgent(serverID).SkpGetTcpSocketMaxForMap(serverID, addrSocket)
}

func (this *SkpManage) SkpCloseAllTcpSocket() {
	this.ManageAgent.SkpCloseAllTcpSocket()
}

func (this *SkpManage) SkpServer(serverID uint64, data []byte, addrSocket uint64) {
	tcpSocket := this.SkpGetTcpSocketForMap(serverID, addrSocket)
	//fmt.Printf("data to server = %d, addrSocket = %d \n", serverID, addrSocket)
	if tcpSocket != nil {
		tcpSocket.SkpWriteChan(data)
	} else {
		fmt.Printf("server = %d, addrSocket = %d is null \n", serverID, addrSocket)
	}
}

func (this *SkpManage) SkpServerConn(connTypeLocal uint8, connTypeRemote uint8, serverID uint64, data []byte) *SkpTcpSocket {
	skpUtility.Printf("send to serverID = %d, this = %p \n", serverID, this)
	tcpSocket := this.SkpGetTcpSocketMaxForMap(serverID, 0)

	if tcpSocket != nil {
		if tcpSocket.SkpWriteChan(data) {
			return tcpSocket
		}
	}

	addr := this.SkpGetServerTcpAddr(uint32(serverID))
	//skpUtility.Printf("send to serverID = %d, addr = %v \n", serverID, addr)
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		skpUtility.Printf("create client conn server = %d, addr = %v, error = %v \n", serverID, addr, err)
		return nil
	} else {
		skpUtility.Printf("create client conn server = %d, addr = %v, success \n", serverID, addr)
	}

	tcpSocketNew, err := this.SkpNewTcpSocket(conn.(*net.TCPConn), connTypeRemote)
	if err != nil {
		skpUtility.Printf("create socket false, connTypeRemote = %d, err = %v \n", connTypeRemote, err)
		return nil
	}

	tcpSocketNew.IsOwner = true
	tcpSocketNew.IsConn = true
	tcpSocketNew.Manage = this

	skpProtocol.SkpPakegeConnHead(&tcpSocketNew.ProtocalConnHead, 1, connTypeLocal, 0, uint64(this.ServerID), uint64(serverID))

	//skpUtility.Printf("conn head = %+v \n", tcpSocketNew.ProtocalConnHead)

	headData, _ := skpProtocol.SkpEncode(tcpSocketNew.ProtocalConnHead)

	tcpSocketNew.SkpWriteChan(headData)

	go tcpSocketNew.SkpRun()

	ret := this.SkpAddTcpSocketToMap(serverID, tcpSocketNew)
	if ret == false {
		tcpSocketNew.SkpCloseChan()
		return nil
	}

	tcpSocketNew.ProtocalConnHead.ConnType = connTypeRemote

	tcpSocketNew.SkpWriteChan(data)

	return tcpSocketNew
}

func (this *SkpManage) SkpGetMonitorIPList() {
	var head skpProtocol.SkpProtocalHead
	skpProtocol.SkpPakegeHead(&head, 0, uint64(this.ServerID), uint64(this.ServerID), skpProtocol.OrderTypeRequest, 0, skpProtocol.OrderMonitorLoadConfig, 0)
	data, _ := skpProtocol.SkpEncode(&head)

	//skpUtility.Printf("head = %+v \n", head)

	tcpSocket := this.SkpServerConn(skpProtocol.ConnTypeProxy, skpProtocol.ConTypeMonitor, uint64(this.MonitorServerID), data)

	if tcpSocket == nil {
		skpUtility.Printf("create monitor server false \n")
		return
	}

__loop:
	for {
		select {
		case <-tcpSocket.WaitMonitorChan:
			break __loop
		}
	}
}
