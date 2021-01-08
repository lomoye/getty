package ziface

//定义一个服务器接口
type IServer interface {
	//启动服务器
	Start()
	//停止服务器
	Stop()
	//运行服务器
	Serve()
	//获取服务器配置信息
	GetConfig() *ServerConfig
	//路由功能：给当前的服务注册一个路由方法
	AddRouter(msgID uint32, router IRouter)
	//获取当前Server的链接管理器
	GetConnMgr() IConnManager
	//注册OnConnStart 钩子函数的方法
	SetOnConnStart(func(conn IConnection))
	//注册OnConnStop钩子函数的方法
	SetOnConnStop(func(conn IConnection))
	//调用OnConnStart钩子函数的方法
	CallOnConnStart(conn IConnection)
	//调用OnConnStop钩子函数的方法
	CallOnConnStop(conn IConnection)
}

type ServerConfig struct {
	Host    string //当前服务器监听的IP
	TcpPort int    //当前服务器监听的端口号
	Name    string //当前服务器的名称

	Version          string //当前Zinx的版本号
	MaxConn          int    //当前服务器主机允许的最大链接数
	WorkerPoolSize   uint32 //当前业务工作Worker池的Goroutine数量
	MaxWorkerTaskLen uint32 //单个队列缓冲长度
}
