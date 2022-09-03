package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	maplock   sync.RWMutex
	message   chan string
}

/*
 */
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		message:   make(chan string),
	}
	return server
}

//监听器，将收到的信息发送给所有用户
func (t *Server) ListenMessager() {
	for {
		msg := <-t.message

		t.maplock.Lock()
		for _, cli := range t.OnlineMap {
			cli.c <- msg
		}
		t.maplock.Unlock()
	}
}

//向消息channal 中加入新消息

func (t *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg + "\n"

	t.message <- sendMsg
}

func (t *Server) Handler(conn net.Conn) {
	//defer runtime.Goexit()
	//
	fmt.Println("连接建立 SUCCEED")
	user := NewUser(conn, t)
	//将用户加入到表中
	// t.maplock.Lock()
	// t.OnlineMap[user.Name] = user //将用户加入map中
	// t.maplock.Unlock()
	user.OnLine()
	isLive := make(chan bool)
	//监听是否活跃

	//t.BroadCast(user, "上线")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				//t.BroadCast(user, "下线")
				user.OffLine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("error!!!!!!!!")
				// if err != nil {
				// 	fmt.Println("!nil")
				// }
				// if err != io.EOF {
				// 	fmt.Println("!eof")
				// }
			}
			msg := string(buf[:n-1])

			//广播消息

			user.Domessage(msg)

			//用户任意操作代表当前活跃
			isLive <- true
		}

	}()
	for {
		select {
		case <-isLive:
			//重置定时器 自动执行下面case
		case <-time.After(time.Second * 60 * 20):
			//已经超时
			//强制关闭当前客户端
			user.SendMsg("滚蛋")
			user.server.maplock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.maplock.Unlock()
			close(user.c)
			conn.Close()

			return
			//
		}
	}
}

func (t *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.Ip, t.Port))
	if err != nil {
		return
	}
	defer listener.Close()
	//监听message
	go t.ListenMessager()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("")
			continue
		}
		go t.Handler(conn)
	}

}
