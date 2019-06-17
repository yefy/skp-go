package rpcD

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync/atomic"
)

type ServerI interface {
	RPC_SetServer(*Server)
	RPC_GetServer() *Server
	RPC_Dispath(method string, args []interface{}) error
	RPC_Stop()
}

type ServerB struct {
	server *Server
}

func (sb *ServerB) RPC_SetServer(server *Server) {
	sb.server = server
}

func (sb *ServerB) RPC_GetServer() *Server {
	return sb.server
}

func (sb *ServerB) RPC_Stop() {

}

type Server struct {
	service ServerI
	*rpc.Server
}

func NewServer(obj ServerI) *Server {
	server := &Server{}
	server.service = obj
	server.Server = rpc.NewServer(server)

	return server
}

func (server *Server) Object() interface{} {
	return server.service
}

func (server *Server) RPC_Start() {
	server.service.RPC_SetServer(server)
}
func (server *Server) RPC_Stop() {
	server.service.RPC_Stop()
}

func (server *Server) RPC_DoMsg(msg *rpc.Msg) {
	service := server.service
	args := msg.Args.([]interface{})

	if msg.Typ == rpc.TypSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			service.RPC_Dispath(msg.Method, args)
		}()
		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypSendReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			err := service.RPC_Dispath(msg.Method, args)
			msg.CB1(err)
		}()

		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.Err = log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			err := service.RPC_Dispath(msg.Method, args)
			msg.Err = err
		}()
		msg.Pending <- msg
	}
}

func (server *Server) Send(method string, args ...interface{}) {
	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSend
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
}

func (server *Server) SendReq(callBack rpc.CallBack1, method string, args ...interface{}) {
	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSendReq
	msg.CB1 = callBack
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
}

func (server *Server) Call(method string, args ...interface{}) error {
	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	defer server.MsgPool.Put(msg)
	msg.Typ = rpc.TypCall
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	<-msg.Pending

	if msg.Err != nil {
		return msg.Err.(error)
	}

	return nil
}
