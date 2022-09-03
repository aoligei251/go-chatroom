package main

import (
	"net"
	"runtime"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	c      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	useraddr := conn.RemoteAddr().String()
	user := &User{
		Name:   useraddr,
		Addr:   useraddr,
		c:      make(chan string),
		conn:   conn,
		server: server,
	}
	//new的时候就启动发送信息的gorutine
	go user.ListenMessage()
	return user
}

//监听user channal 吧消息写给客户端
func (t *User) ListenMessage() {
	for {
		//
		_, ok := t.server.OnlineMap[t.Name]
		if !ok {
			runtime.Goexit()
		}
		meg := <-t.c
		t.conn.Write([]byte(meg))
	}
}

func (t *User) OnLine() {
	t.server.maplock.Lock()
	t.server.OnlineMap[t.Name] = t //将用户加入map中
	t.server.maplock.Unlock()
	t.server.BroadCast(t, "上线")
}

func (t *User) OffLine() {
	t.server.maplock.Lock()
	//t.server.OnlineMap[t.Name] = t //将用户从map中删除
	delete(t.server.OnlineMap, t.Name)
	t.server.maplock.Unlock()
	t.server.BroadCast(t, "下线")
}

func (t *User) SendMsg(msg string) {
	t.conn.Write([]byte(msg))
}

//广播
func (t *User) Domessage(msg string) {
	//查询当前在线用户
	if msg == "who" {
		t.server.maplock.Lock()
		for _, usr := range t.server.OnlineMap {
			OnLineMsg := "[" + usr.Addr + "]" + usr.Name + "online......\n"
			t.SendMsg(OnLineMsg)
		}
		t.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//修改用户名
		NewName := strings.Split(msg, "|")[1]
		//判断NewName是否已经存在

		_, ok := t.server.OnlineMap[NewName]
		if ok {
			t.SendMsg("当前用户名已存在")
		} else {
			t.server.maplock.Lock()
			//t.server
			delete(t.server.OnlineMap, t.Name)
			t.server.OnlineMap[NewName] = t
			t.server.maplock.Unlock()
			t.Name = NewName
			t.SendMsg("新用户名" + NewName + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// to|user1|message
		//得到用户名
		remotename := strings.Split(msg, "|")[1]
		if remotename == "" {
			t.SendMsg("消息格式不正确...to|user1|message\n")
			return
		}
		//获取user对象
		remoteuser, ok := t.server.OnlineMap[remotename]
		if !ok {
			t.SendMsg("该用户不存在\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			t.SendMsg("消息为空\n")
		}
		remoteuser.SendMsg(t.Name + ":" + content)

	} else {
		t.server.BroadCast(t, msg)
	}

}
