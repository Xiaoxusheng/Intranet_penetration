package main

import (
	"Intranet_penetration/utility"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
)

var (
	size int64
)

type UdpConn struct {
	udpClient *net.UDPConn //请求隧道
	errChan   chan error   //存放错误管道
	read      chan []byte
	localhost chan []byte
	io.Writer
	io.Reader
	utility.UdpLimit
	clientIP string
}

func NewUdpConn() *UdpConn {
	udpClient := utility.CreateUdpConn(utility.TunnelPort)
	fmt.Println("udpClient", udpClient)
	return &UdpConn{
		udpClient: udpClient,
		errChan:   make(chan error),
		Writer:    udpClient,
		Reader:    udpClient,
	}
}

func (u *UdpConn) Read(p []byte) (n int, err error) {
	n, err = u.Reader.Read(p)
	if err != nil {
		return 0, err
	}
	atomic.AddInt64(&size, int64(n))
	u.UdpLimit.Log(size)
	fmt.Println("size", size/(1024*1024)/2)
	return n, err
}

func (u *UdpConn) Write(p []byte) (n int, err error) {
	n, err = u.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func read(ctx context.Context, u *UdpConn) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			data := make([]byte, 1024*1000)
			n, add, err := u.udpClient.ReadFromUDP(data)
			if err != nil {
				u.errChan <- err
				continue
			}
			if add.IP.String() == utility.TunnelPort {
				u.read <- data[:n]
			} else {
				u.clientIP = add.IP.String()
				u.localhost <- data[:n]
			}

		}
	}
}

func write(ctx context.Context, u *UdpConn) {
	for {
		select {
		case <-ctx.Done():
			break
		case data := <-u.read:
			_, err := u.udpClient.WriteToUDP(data, utility.CreateUDPAddr(utility.Localhost))
			if err != nil {
				u.errChan <- err
				continue
			}
		case r := <-u.localhost:
			_, err := u.udpClient.WriteToUDP(r, utility.CreateUDPAddr(utility.TunnelPort))
			if err != nil {
				u.errChan <- err
				continue
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	//	初始化
	updClient := NewUdpConn()
	log.Printf("  [隧道连接成功::]%v  [本地连接成功::]", updClient.udpClient.RemoteAddr().String())

	go read(ctx, updClient)
	go write(ctx, updClient)

	for {
		select {
		case <-ctx.Done():
			cancel()
		}
	}

}
