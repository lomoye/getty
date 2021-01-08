package znet

import (
	"encoding/json"
	"fmt"
	"getty/ziface"
	"net"
)

//IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器的名称
	Name string
	//服务器绑定的ip版本
	IPVersion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前的Server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandler ziface.IMsgHandler
	//链接管理器
	ConnMgr ziface.IConnManager
	//调Server创建连接之后自动调用Hook函数
	OnConnStart func(conn ziface.IConnection)
	//调Server销毁连接之后
	OnConnStop func(conn ziface.IConnection)
	//服务器配置信息
	Config *ziface.ServerConfig
}

//启动服务器
func (s *Server) Start() {
	fmt.Printf("[Zinx] Server Name : %s, listenner at IP :%s, Port %d is starting\n", s.Name, s.IP, s.Port)
	go func() {
		// 0 开启消息队列及Worker工作池
		s.MsgHandler.StartWorkerPool(s.GetConfig())

		// 1 获取一个tcp的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error :", err)
		}
		// 2 监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, " err ", err)
			return
		}
		fmt.Println("start Zinx server succ, ", s.Name, " succ, Listenning...")
		var cid uint32
		cid = 0
		// 3 阻塞的等待客户端连接，处理客户端连接业务（读写）
		for {
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//设置最大链接个数的判断，如果超过最大链接，那么则关闭此新的链接
			if s.ConnMgr.Len() >= s.Config.MaxConn {
				//TODO 给客户端响应一个超出最大链接的错误包
				fmt.Println("Too many Connection")
				conn.Close()
				continue
			}

			//已经与客户端建立连接， 做一些业务， 做一个最基本的512字节的回显业务
			dealConn := NewConnetion(s, conn, cid, s.MsgHandler)
			cid++
			//启动当前的链接业务处理
			go dealConn.Start()
		}
	}()
}

//停止服务器
func (s *Server) Stop() {
	//将一些服务器的资源，状态或者一些已开辟的连接信息 进行停止或回收
	fmt.Println("[STOP] Zinx server stop", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add Router Succ!!")
}

//运行服务器
func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	//阻塞状态
	select {}
}

func (s *Server) GetConfig() *ziface.ServerConfig {
	return s.Config
}

/*
 *	初始化Server模块的方法
 */
func NewServer(config *ziface.ServerConfig) ziface.IServer {
	defaultConfig := `{
		"Name":             "zinx",
		"Host":             "127.0.0.1",
		"TcpPort":          9999,
		"Version":          "1.0.0",
		"MaxConn":          65535,
		"WorkerPoolSize":   65535,
		"MaxWorkerTaskLen": 10000
	}`

	err := json.Unmarshal([]byte(defaultConfig), config)
	if err != nil {
		panic(err)
	}

	s := &Server{
		Name:       config.Name,
		IPVersion:  "tcp4",
		IP:         config.Host,
		Port:       config.TcpPort,
		Config:     config,
		MsgHandler: NewMsgHandler(config),
		ConnMgr:    NewConnManager(),
	}

	return s
}

//注册OnConnStart 钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

//注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

//调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> Call OnConnStart() ...")
		s.OnConnStart(conn)
	}
}

//调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> Call OnConnStop() ...")
		s.OnConnStop(conn)
	}
}
