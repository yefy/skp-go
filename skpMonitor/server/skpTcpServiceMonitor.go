package server

import (
	"net"
	"skp/skpPublic/skpProtocol"
	"skp/skpPublic/skpSocket"
	"skp/skpPublic/skpUtility"
	"strconv"
)

type SkpTcpServiceMonitor struct {
	*skpSocket.SkpTcpService
	ManageMonitor *SkpManageMonitor
}

func SkpNewTcpServiceMonitor() (this *SkpTcpServiceMonitor) {
	tcpServiceMonitor := new(SkpTcpServiceMonitor)
	tcpServiceMonitor.SkpTcpService = skpSocket.SkpNewTcpService(tcpServiceMonitor)
	tcpServiceMonitor.ManageMonitor = SkpNewManageMonitor(tcpServiceMonitor)
	tcpServiceMonitor.Manage = tcpServiceMonitor.ManageMonitor.SkpManage

	return tcpServiceMonitor
}

func (this *SkpTcpServiceMonitor) SkpLoadConf() {
	skpUtility.Printf("read config \n")

	serverIDStr := this.Conf.GetValue("server", "port")
	serverID, _ := strconv.Atoi(serverIDStr)

	this.Manage.SkpSetServerType(skpProtocol.ConTypeMonitor)

	this.Manage.SkpSetServerID(uint32(serverID))

	this.Manage.SkpAddServerIPPort(11111, "127.0.0.1", 11111)
	this.Manage.SkpAddServerIPPort(11112, "127.0.0.1", 11112)
	this.Manage.SkpAddServerIPPort(22221, "127.0.0.1", 22221)
	this.Manage.SkpAddServerIPPort(33331, "127.0.0.1", 33331)
	//this.ManageMonitor.SkpAddServerIPPort(33331, "192.168.164.135", 33331)
	//this.ManageMonitor.SkpAddServerIPPort(33331, "172.16.100.40", 33331)
	//this.ManageMonitor.SkpAddServerIPPort(33331, "192.168.164.147", 33331)
}

func (this *SkpTcpServiceMonitor) SkpAcceptConn(conn *net.TCPConn) {
	tcpSocketMonitor := SkpNewTcpSocketMonitor(conn, this)
	go tcpSocketMonitor.SkpRun()
}
