package skpProtocol

import (
	"bytes"
	"encoding/binary"
	_ "encoding/gob"
	"errors"
)

const (
	ConnTypeClient = iota
	ConnTypeProxy
	ConnTypeLocal
	ConnTypeRoute
	ConTypeMonitor
)

const (
	OrderTypeRequest = iota
	OrderTypeResponse
	OrderTypeServer
)

const (
	ModuleTypeMonitor = iota
	ModuleTypeProxy
	ModuleTypeLocalMemoryFile
	ModuleTypeLocalText
)

var (
	ModuleMonitor         uint64 = ModuleTypeMonitor << 16
	ModuleLocalMemoryFile uint64 = ModuleTypeLocalMemoryFile << 16
)

var (
	OrderMonitorLoadConfig   uint64 = ModuleMonitor + 0x01
	OrderLocalMemoryFilePull uint64 = ModuleLocalMemoryFile + 0x01
)

type SkpIns struct {
	Number           uint16
	Module           uint16
	IsCryptoRequest  uint8 // 指令是加密的
	IsCryptoResponse uint8 // 回复需要加密
	InsType          uint8
	R                uint8
}

type SkpProtocalConnHead struct {
	HeadSize     uint16
	IsKeep       uint8
	IsConnHead   uint8
	DataSize     uint32
	FromServerID uint64
	ToServerID   uint64
	ConnType     uint8
	R            uint8
	R2           uint8
	R3           uint8
	R4           uint32
}

type SkpProtocalHead struct {
	HeadSize    uint16
	OrderType   uint8
	IsConnHead  uint8
	DataSize    uint32
	Send        uint64
	Recv        uint64
	ProxyID     uint32
	Status      uint16
	Error       uint16
	AddrSocket  uint64
	ClientMark  uint64
	Param       uint32
	R           uint32
	OrderRequst uint64
	OrderServer uint64
}

type SkpProxyID struct {
	Data1 uint8
	Data2 uint8
	Data3 uint8
	Data4 uint8
}

type SkpAddrSocket struct {
	Data1 uint8
	Data2 uint8
	Data3 uint8
	Data4 uint8
	Data5 uint8
	Data6 uint8
	Data7 uint8
	Data8 uint8
}

func SkpPakegeConnHead(head *SkpProtocalConnHead, isKeep uint8, connType uint8, dataSize uint32, fromServerID uint64, toServerID uint64) {
	head.HeadSize = uint16(SkpInterfaceSize(head))
	head.IsConnHead = 1
	head.IsKeep = isKeep
	head.ConnType = connType
	head.DataSize = dataSize
	head.FromServerID = fromServerID
	head.ToServerID = toServerID
}

func SkpPakegeHead(head *SkpProtocalHead, dataSize uint32, send uint64, recv uint64, orderType uint8, proxyID uint32, orderRequst uint64, orderServer uint64) {
	head.HeadSize = uint16(SkpInterfaceSize(head))
	head.IsConnHead = 0
	head.OrderType = orderType
	head.DataSize = dataSize
	head.Send = send
	head.Recv = recv
	head.ProxyID = proxyID
	head.OrderRequst = orderRequst
	head.OrderServer = orderServer
}

func SkpCheckConnHead(data []byte, head *SkpProtocalConnHead) error {
	var headSize int = binary.Size(head)
	if len(data) < headSize {
		return errors.New("head size is not enough")
	}

	SkpDecode(data, head)

	if len(data) < int(head.HeadSize) {
		return errors.New("head size is not enough")
	}

	//fmt.Printf("conn head = %+v \n", head)

	allSize := int(int(head.HeadSize) + int(head.DataSize))
	if len(data) < allSize {
		return errors.New("data size is not enough")
	}

	return nil
}

func SkpCheckHead(data []byte, head *SkpProtocalHead) error {
	var headSize int = binary.Size(head)
	if len(data) < headSize {
		return errors.New("head size is not enough")
	}

	SkpDecode(data, head)

	if len(data) < int(head.HeadSize) {
		return errors.New("head size is not enough")
	}

	//fmt.Printf("head = %+v \n", head)

	allSize := int(int(head.HeadSize) + int(head.DataSize))
	if len(data) < allSize {
		return errors.New("data size is not enough")
	}

	return nil
}

func SkpEncode(data interface{}) ([]byte, error) {

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, data)
	return buf.Bytes(), nil

	//	buf := bytes.NewBuffer(nil)
	//	enc := gob.NewEncoder(buf)
	//	err := enc.Encode(data)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return buf.Bytes(), nil
}

func SkpDecode(data []byte, to interface{}) error {

	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, to)
	return nil

	//	buf := bytes.NewBuffer(data)
	//	dec := gob.NewDecoder(buf)
	//	return dec.Decode(to)
}

func SkpInterfaceSize(data interface{}) int {
	return binary.Size(data)
}
