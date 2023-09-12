package main

import (
	"Intranet_penetration/utility"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
		//fmt.Println("string(data):", string(data), len(data), string(data) == "hello,wrold", len("hello,wrold"))
		fmt.Println("有客户端连接服务")
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
		// 设置keep-alive的间隔时间
		err = controlCon.SetKeepAlivePeriod(30 * time.Second)
		if err != nil {
			log.Println("设置失败" + err.Error())

		}
		go utility.KeepAlive(controlCon)
	}
}

func userRequestService() {
	Con := utility.CreateLister(utility.UserRequestPort)
	log.Printf("[用户请求监听]%v", Con.Addr().String())
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
	fmt.Println("请输入可用流量大小,单位为m")
	fmt.Scanln(&utility.FlowRate)
	go controlService()
	//用户请求端
	go userRequestService()
	//隧道端
	go tunnelService()

	wg.Wait()
}
