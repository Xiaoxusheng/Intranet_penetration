package main

import (
	`Intranet_penetration/utility`
	`bufio`
	`io`
	`log`
)

//连接控制端
func controlsClient() {
	//连接控制端
	tCPConn := utility.CreateConn(utility.ControlPort)
	//验证身份
	_, err := tCPConn.Write([]byte("hello,wrold"))
	if err != nil {
		log.Println(err)
	}
	log.Printf("[客户端连接成功::] %v", tCPConn.RemoteAddr().String())
	for {
		readString, err := bufio.NewReader(tCPConn).ReadString(byte('\n'))
		log.Println(readString)
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("验证失败，服务器断开连接!")
				break
			}
			log.Println("读取失败！", err)
			break
		}
		if readString == utility.SendMessage {
			go getMessage()
		}
	}
}

//连接隧道
func getMessage() {
	//隧道
	conn := utility.CreateConn(utility.TunnelPort)
	//本地服务器
	localhost := utility.CreateConn(utility.Localhost)

	r := &utility.Reader{
		Reader: localhost,
		Writer: localhost,
	}
	go io.Copy(r, conn)
	go io.Copy(conn, r)

}

func main() {
	//连接控制端
	controlsClient()

}
