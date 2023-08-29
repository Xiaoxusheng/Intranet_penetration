package utility

import (
	"log"
	"net"
)

type UdpLimit interface {
	Limit(n int64, conn *net.UDPConn)
	Log(rd int64)
}

func CreateUdpLister(address string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("创建失败" + err.Error())
	}

	udp, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("创建Udp失败" + err.Error())
	}
	return udp
}

func CreateUdpConn(address string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("创建失败" + err.Error())
	}
	udpConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println("创建Udp连接失败" + err.Error())
	}
	return udpConn
}

func CreateUDPAddr(address string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("创建失败" + err.Error())
	}
	return addr
}
