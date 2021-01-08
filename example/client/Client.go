package main

import (
	"fmt"
	"getty/znet"
	"io"
	"net"
	"time"
)

/*
 模拟客户端
*/
func main() {
	fmt.Println("client start...")
	time.Sleep(1 * time.Second)

	maxConn := 1
	for i := 0; i < maxConn; i++ {
		// 1.直接连接远程服务器，得到一个conn连接
		conn, err := net.Dial("tcp", "127.0.0.1:9999")
		if err != nil {
			fmt.Println("client start err, exit!")
			return
		}
		//2 连接调用Write 写数据
		go func() {
			for {
				//发送封包的message消息
				dp := znet.NewDataPack()
				binaryMsg, err := dp.Pack(znet.NewMsgPackage(1, []byte("hello it's testing server tps")))
				if err != nil {
					fmt.Println("Pack error", err)
					return
				}

				_, err = conn.Write(binaryMsg)
				if err != nil {
					fmt.Println("write error", err)
					return
				}

				//服务器应该给我们回复一个message数据， MsgID: 0, pingpingping

				//先读取流中的head部分 得到ID 和 dataLen

				//再根据DataLen进行第二次读取，将data读出来
				binaryHead := make([]byte, dp.GetHeadLen())
				if _, err := io.ReadFull(conn, binaryHead); err != nil {
					fmt.Println("read head error", err)
					break
				}

				//将二进制的head拆包到msg结构体中
				msgHead, err := dp.Unpack(binaryHead)
				if err != nil {
					fmt.Println("client unpack msgHead error", err)
					break
				}

				if msgHead.GetMsgLen() > 0 {
					//2.再根据DataLen进行第二次读取，将data读出来
					msg := msgHead.(*znet.Message)
					msg.Data = make([]byte, msg.GetMsgLen())
					if _, err := io.ReadFull(conn, msg.Data); err != nil {
						fmt.Println("read msg data error", err)
						return
					}
					fmt.Println("---> Recv Server Msg : ID = ", msg.Id, ", len= ", msg.DataLen, ", data = ", string(msg.Data))
				}
			}
		}()
	}

	select {}
}
