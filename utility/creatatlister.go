package utility

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

const (
	ControlPort     = ":8080"
	UserRequestPort = ":8081"
	TunnelPort      = ":8082"
	Localhost       = ":80"
	SendMessage     = "A New Request Join!\n"
)

var (
	FlowRate  int64 = 220 //M为单位，超过就会主动断开客户端连接
	StartChan       = make(chan bool, 10)
)

type Reader struct {
	io.Reader
	io.Writer
	Size int64
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return 0, err
	}
	atomic.AddInt64(&r.Size, int64(n))
	r.log(r.Size)
	return n, err
}

func (r *Reader) Write(p []byte) (n int, err error) {
	n, err = r.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (r *Reader) log(rd int64) {
	switch {
	case rd <= (1024*1024)/2:
		log.Printf("转发流量为：%vM", rd/(1024*1024)/2)
	case (1024*1024)/2 < rd && rd <= 1024*((1024*1024)/2):
		log.Printf("转发流量为：%vG %vM", int(rd/(1024*1024)/2/1024), rd/(1024*1024)/2%1024)
	}
}

func (r *Reader) Limit(n int64, net *net.TCPConn) {
	fmt.Println("Limit启动", r.Size/(1024*1024), n)
	for {
		fmt.Println("已经使用流量数：", r.Size/(1024*1024)/2)
		time.Sleep(time.Second * 3)
		if r.Size/(1024*1024)/2 > n {
			_, err := net.Write([]byte("您已达到流量上限！\n"))
			if err != nil {
				log.Println("写入出错!" + err.Error())
			}
			//主动关闭连接
			net.Close()
			break
		}
	}

}

// 创建监听
func CreateLister(addrs string) *net.TCPListener {
	addr, err := net.ResolveTCPAddr("tcp", addrs)
	if err != nil {
		log.Println("服务", err)
	}
	tcp, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println(err)
	}
	return tcp

}

// 创建连接
func CreateConn(addrs string) *net.TCPConn {
	addr, err := net.ResolveTCPAddr("tcp", addrs)
	if err != nil {
		log.Println("客户", err)
	}
	tcp, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
	}
	return tcp
}

//func KeepAlive(tcp *net.TCPConn) {
//	for {
//		_, err := tcp.Write([]byte("hello,world!\n"))
//		if err != nil {
//			log.Println(err)
//			continue
//		}
//		time.Sleep(time.Second * 5)
//	}
//}
