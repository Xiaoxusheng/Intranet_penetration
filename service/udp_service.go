package main

import (
	"Intranet_penetration/utility"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

var (
	size int64
	k    bool
	num  int64 = 1024
)

// 隧道外端
type UdpLister struct {
	udpService  *net.UDPConn //服务端
	errChan     chan error   //存放错误管道
	userRead    chan []byte
	userWrite   chan []byte
	clientRead  chan []byte
	clientWrite chan []byte
	clientIP    string
	io.Writer
	io.Reader
}

// 用户
type UdpUserRequest struct {
	udpUserRequest *net.UDPConn //服务端
	errChan        chan error   //存放错误管道
}

func (u *UdpLister) Read(p []byte) (n int, err error) {
	n, err = u.Reader.Read(p)
	if err != nil {
		return 0, err
	}
	atomic.AddInt64(&size, int64(n))
	u.Log(size)
	fmt.Println("size", size/(1024*1024)/2)
	return n, err
}

func (u *UdpLister) Write(p []byte) (n int, err error) {
	n, err = u.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (u *UdpLister) Log(rd int64) {
	switch {
	case rd > 1024*((1024*1024)/2):
		log.Printf("转发流量为：%vG %vM", int(rd/(1024*1024)/2/1024), rd/(1024*1024)/2%1024)
	case rd > (1024*1024)/2:
		log.Printf("转发流量为：%vM", rd/(1024*1024)/2)
	}
}

// 限制流量
func (u *UdpLister) Limit(n int64, net *net.UDPConn) {
	fmt.Println("Limit启动", size/(1024*1024), n)
	for {

		if size/(1024*1024)/2 > n {
			_, err := net.Write([]byte("您已达到流量上限！\n"))
			if err != nil {
				log.Println("写入出错!" + err.Error())
			}
			//主动关闭连接
			net.Close()
			k = true
		}
		if k {
			return
		}
		time.Sleep(time.Second * 3)
		fmt.Println("已经使用流量数：", size/(1024*1024)/2)
	}

}

func (u *UdpLister) readByte(ctx context.Context) {
	for {

		select {
		case <-ctx.Done():
			break
		default:
			data := make([]byte, 1024)
			n, m, err := u.udpService.ReadFromUDP(data)
			if err != nil {
				u.errChan <- err
			}
			if m.IP.String() == utility.UserRequestPort {
				u.userRead <- data[:n]
			} else {
				u.clientIP = m.IP.String()
				u.clientRead <- data[:n]
			}
			log.Println("收到来自", m.IP.String())
		}
	}
}

func (u *UdpLister) writeByte(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		//用户发来的，转发到client
		case r := <-u.userRead:
			//写入客户端
			_, err = u.udpService.WriteToUDP(r, utility.CreateUDPAddr(u.clientIP))
			if err != nil {
				u.errChan <- err
			}
		case m := <-u.clientRead:
			//写入用户端
			_, err := u.udpService.WriteToUDP(m, utility.CreateUDPAddr(utility.UserRequestPort))
			if err != nil {
				u.errChan <- err
			}
		}

	}
}

func NewUdpLister() *UdpLister {
	udp := utility.CreateUdpLister(utility.TunnelPort)
	return &UdpLister{
		udpService: udp,
		errChan:    make(chan error),
		Writer:     udp,
		Reader:     udp,
	}
}

func NewUdpUserRequest() *UdpUserRequest {
	udp := utility.CreateUdpLister(utility.UserRequestPort)
	return &UdpUserRequest{
		udpUserRequest: udp,
		errChan:        make(chan error),
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	//隧道初始化
	tunnelUdp := NewUdpLister()
	//用户端初始化
	userUdp := NewUdpUserRequest()

	//defer func() {
	//	tunnelUdp.udpService.Close()
	//	userUdp.udpUserRequest.Close()
	//}()
	fmt.Println("请输入流量限制")
	fmt.Scanln(&num)

	//go tunnelUdp.Limit(num, tunnelUdp.udpService)
	go tunnelUdp.readByte(ctx)
	go tunnelUdp.writeByte(ctx)

	for {

		select {
		case <-ctx.Done():
			break
		case err := <-tunnelUdp.errChan:
			log.Println("隧道出现错误" + err.Error())
			cancel()
		case err := <-userUdp.errChan:
			log.Println("用户端出现错误" + err.Error())
			cancel()
		}
	}

}
