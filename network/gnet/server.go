package gnet

import (
	"context"
	"github.com/dobyte/due/v2/network"
	"github.com/panjf2000/gnet/v2"
	"net"
)

type server struct {
	opts              *serverOptions            // 配置
	listener          net.Listener              // 监听器
	connMgr           *serverConnMgr            // 连接管理器
	startHandler      network.StartHandler      // 服务器启动hook函数
	stopHandler       network.CloseHandler      // 服务器关闭hook函数
	connectHandler    network.ConnectHandler    // 连接打开hook函数
	disconnectHandler network.DisconnectHandler // 连接关闭hook函数
	receiveHandler    network.ReceiveHandler    // 接收消息hook函数
	engine            *engine                   // 引擎
}

var _ network.Server = &server{}

func NewServer(opts ...ServerOption) network.Server {
	o := defaultServerOptions()
	for _, opt := range opts {
		opt(o)
	}

	s := &server{}
	s.opts = o
	s.connMgr = newConnMgr(s)
	s.engine = &engine{server: s, connMgr: newConnMgr(s)}

	return s
}

// Addr 监听地址
func (s *server) Addr() string {
	return s.opts.addr
}

// Start 启动服务器
func (s *server) Start() error {
	if err := s.init(); err != nil {
		return err
	}

	return nil
}

// Stop 关闭服务器
func (s *server) Stop() error {
	if err := s.engine.stop(context.Background()); err != nil {
		return err
	}

	s.connMgr.close()

	return nil
}

// Protocol 协议
func (s *server) Protocol() string {
	return "tcp"
}

// OnStart 监听服务器启动
func (s *server) OnStart(handler network.StartHandler) {
	s.startHandler = handler
}

// OnStop 监听服务器关闭
func (s *server) OnStop(handler network.CloseHandler) {
	s.stopHandler = handler
}

// OnConnect 监听连接打开
func (s *server) OnConnect(handler network.ConnectHandler) {
	s.connectHandler = handler
}

// OnDisconnect 监听连接关闭
func (s *server) OnDisconnect(handler network.DisconnectHandler) {
	s.disconnectHandler = handler
}

// OnReceive 监听接收到消息
func (s *server) OnReceive(handler network.ReceiveHandler) {
	s.receiveHandler = handler
}

// 初始化TCP服务器
func (s *server) init() error {
	return gnet.Run(s.engine, "tcp4://"+s.opts.addr, gnet.WithTicker(s.opts.heartbeatMechanism == TickHeartbeat))
}

//// 等待连接
//func (s *server) serve() {
//	var tempDelay time.Duration
//
//	for {
//		conn, err := s.listener.Accept()
//		if err != nil {
//			if e, ok := err.(net.Error); ok && e.Timeout() {
//				if tempDelay == 0 {
//					tempDelay = 5 * time.Millisecond
//				} else {
//					tempDelay *= 2
//				}
//				if max := 1 * time.Second; tempDelay > max {
//					tempDelay = max
//				}
//
//				log.Warnf("tcp accept error: %v; retrying in %v", err, tempDelay)
//				time.Sleep(tempDelay)
//				continue
//			}
//
//			log.Errorf("tcp accept error: %v", err)
//			return
//		}
//
//		tempDelay = 0
//
//		if err = s.connMgr.allocate(conn); err != nil {
//			log.Errorf("connection allocate error: %v", err)
//			_ = conn.Close()
//		}
//	}
//}