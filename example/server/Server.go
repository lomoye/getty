package main

import (
	"encoding/json"
	"fmt"

	"github.com/lomoye/getty/utils"
	"github.com/lomoye/getty/ziface"
	"github.com/lomoye/getty/znet"
)

/*
 基于Zinx框架来开发的 服务器端应用程序
*/
type TestTpsRouter struct {
	znet.BaseRouter
}

func (br *TestTpsRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handle...")
	//读取客户端的数据，再回写ping
	fmt.Println("recv from client: msgId = ", request.GetMsgID(), ", data = ", string(request.GetData()))
	resp, err := json.Marshal(utils.GlobalTpsCounter.GetWindows())
	if err != nil {
		panic("Handle Json Marshal error")
	}
	err = request.GetConnection().SendMsg(request.GetMsgID(), []byte(resp))
	if err != nil {
		fmt.Println(err)
	}
}

//在处理conn业务之后的钩子方法Hook
func (br *TestTpsRouter) PostHandle(request ziface.IRequest) {
	utils.GlobalTpsCounter.Inc()
}

func main() {
	//1.创建一个server句柄，使用Zinx的api
	config := &ziface.ServerConfig{
		Name:             "test",
		Version:          "1.0.0",
		MaxConn:          65535,
		WorkerPoolSize:   65535,
		MaxWorkerTaskLen: 10000,
	}
	s := znet.NewServer(config)
	//2.给当前zinx添加一个自定义的router
	s.AddRouter(1, &TestTpsRouter{})
	//3.启动server
	s.Serve()
}

//Test PreHandle
