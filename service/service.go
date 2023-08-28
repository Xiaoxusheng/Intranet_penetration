package main

import (
	"Intranet_penetration/utility"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var useConn *net.TCPConn
var controlCon *net.TCPConn
var err error

// 服务器 放置在公网上
func controlService() {
	con := utility.CreateLister(utility.ControlPort)
	log.Printf("[控制启动监听]%v", con.Addr().String())
	data := make([]byte, 11)
	for {
		controlCon, err = con.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		//验证客户端
		_, err := controlCon.Read(data)
		if err != nil {
			log.Println("解析读取错误" + err.Error())
			continue
		}
		fmt.Println("string(data):", string(data), len(data), string(data) == "hello,wrold", len("hello,wrold"))
		if string(data) != "hello,wrold" {
			err := controlCon.Close()
			if err != nil {
				log.Println("主动断开连接失败！")
			}
			continue
		}
		//保活

		err = controlCon.SetKeepAlive(true)
		if err != nil {
			log.Println("保活失败" + err.Error())
			continue
		}
		//go utility.KeepAlive(controlCon)
	}
}

func userRequestService() {
	Con := utility.CreateLister(utility.UserRequestPort)
	log.Printf("[用户请求监听] %v\n", Con.Addr().String())
	for {
		useConn, err = Con.AcceptTCP()
		if err != nil {
			log.Println(err)
		}
		fmt.Println("controlCon", controlCon)
		//controlCon只有有客户端连接才会生成，所有需要等待一下
		if controlCon == nil {
			continue
		}
		//防止controlCon还没有启动造成空指针问题
		_, err = controlCon.Write([]byte(utility.SendMessage))
		if err != nil {
			log.Println("发送出错！" + err.Error())
			controlCon.Close()
		}
		fmt.Println("发送成功")
	}

}

func tunnelService() {
	con := utility.CreateLister(utility.TunnelPort)
	log.Printf("[隧道启动监听]%v", con.Addr().String())

	for {
		tunnelConn, err := con.AcceptTCP()
		if err != nil {
			log.Println(err)
		}
		r := &utility.Reader{
			Reader: tunnelConn,
			Writer: tunnelConn,
		}
		//流量限制,关闭隧道
		go r.Limit(utility.FlowRate, controlCon)
		go io.Copy(useConn, r)
		go io.Copy(r, useConn)

	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	//控制端，控制连接
	go controlService()
	//用户请求端
	go userRequestService()
	//隧道端
	go tunnelService()

	wg.Wait()

	//addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	//if err != nil {
	//	log.Println(err)
	//}
	//tcp, err := net.ListenTCP("tcp", addr)
	//if err != nil {
	//	log.Println(err)
	//}
	//fmt.Println(tcp.Addr())
	//
	//log.Println("服务端启动!")
	//for {
	//	//data := make([]byte, 1024)
	//	acceptTCP, err := tcp.AcceptTCP()
	//	if err != nil {
	//		log.Println(err)
	//	}
	//
	//	_, err = acceptTCP.Write([]byte("go is good"))
	//	if err != nil {
	//		log.Println(err)
	//	}
	//	for {
	//		readString, err := bufio.NewReader(acceptTCP).ReadString('\n')
	//		if err != nil {
	//			log.Println(err)
	//		}
	//		fmt.Println("读到的数据", readString)
	//		fmt.Println("阻塞...")
	//
	//	}
	//
	//	//_, err = acceptTCP.Read(readString)
	//	//if err != nil {
	//	//	log.Println(err)
	//	//}
	//}
}
