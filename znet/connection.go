package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/lomoye/getty/ziface"
)

/*
 链接模块
*/
type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer

	//当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接的ID
	ConnID uint32

	//当前的链接状态
	isClosed bool

	//告知当前链接已经退出的/停止 channel（由Reader告知Writer退出）
	ExitChan chan bool

	//无缓冲管道，用于读、写Goroutine之间的消息通信
	msgChan chan []byte

	//消息的管理MsgID对应的处理业务API关系
	MsgHandler ziface.IMsgHandler

	//链接属性集合
	property map[string]interface{}

	//保护链接属性的锁
	propertyLock sync.RWMutex
}

//初始化链接模块的方法
func NewConnetion(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandler) *Connection {
	c := &Connection{
		TcpServer:  server,
		Conn:       conn,
		ConnID:     connID,
		MsgHandler: msgHandler,
		isClosed:   false,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
		property:   make(map[string]interface{}),
	}
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

//写消息Goroutine, 专门发送给客户端消息的模块
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]", c.RemoteAddr().String())

	//不断的阻塞的等待channl的消息, 进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error ", err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，此时Writer也要退出
			return
		}

	}
}

//链接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID=", c.ConnID, "Reader is exit， remote addr is ", c.RemoteAddr().String())
	defer c.Stop()
	for {
		//创建一个拆包解包对象
		dp := NewDataPack()

		//读取客户端的Msg Head 8个字节
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("rad msg head error", err)
			break
		}
		//拆包，得到msgId和msgDataLen 放在一个msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			break
		}
		//根据dataLen 再次读取Data，放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				break
			}
		}
		msg.SetData(data)
		//得到当前conn数据的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		if c.TcpServer.GetConfig().WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从路由中，创造到注册绑定的Conn对应的router调用
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

//提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection closed when send msg")
	}

	//将data进行封包 MsgDataLen MsgID|Data
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id =", msgId)
		return errors.New("Pack error msg")
	}
	//将数据发送给客户端
	c.msgChan <- msg

	return nil
}

//启动链接 让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start() ... ConnID = ", c.ConnID)
	//启动从当前链接的读数据的业务
	go c.StartReader()
	//启动从当前链接写数据的业务
	go c.StartWriter()

	//按照开发者传递进来的创建链接之后需要执行的业务
	c.TcpServer.CallOnConnStart(c)
}

//停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop().. ConnID =", c.ConnID)
	//如果当前链接已经关闭
	if c.isClosed {
		return
	}
	c.isClosed = true

	//调用开发者注册的销毁链接之前需要执行的业务Hook函数
	c.TcpServer.CallOnConnStop(c)
	//关闭socket链接
	c.Conn.Close()
	//告知Writeg关闭
	c.ExitChan <- true
	//将当前链接从ConnMgr中移除
	c.TcpServer.GetConnMgr().Remove(c)
	//回收资源
	close(c.ExitChan)
	close(c.msgChan)
}

//获取当前链接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//获取远程客户端的 TCP状态 IP port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
