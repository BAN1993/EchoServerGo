package main

import (
	"net/http"
	"time"

	"EchoServerGo/flog"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
	timerWait = time.Second
)

type OnAcceptCallback 	func(c *IOClient)
type OnCloseCallback 	func(c *IOClient)
type OnRecvCallback 	func(c *IOClient, buf []byte)
type OnTimerCallback	func()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type IOClient struct {
	manager *ServerSocket
	conn    *websocket.Conn

	ChSend		chan []byte
}

type RecvPackage struct {
	Client 		*IOClient
	buf			[]byte
}

func newIOClient(s *ServerSocket, c *websocket.Conn) *IOClient {
	return &IOClient{
		manager: s,
		conn:    c,
		ChSend:  make(chan []byte),
	}
}

func (c *IOClient) goRead() {
	defer func() {
		c.manager.ChClose <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			_= c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		if c.manager.ChRecv != nil {
			p := &RecvPackage{Client:c, buf:message}
			c.manager.ChRecv <- p
		}
	}
}

func (c *IOClient) goWrite() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case buf, ok := <- c.ChSend:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(buf)
			if err := w.Close(); err != nil {
				return
			}
			//c.recvCount++

		case <- ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

type ServerSocket struct {
	ClientList	map[*IOClient]bool // 客户端列表

	ChAccept	chan *IOClient
	ChClose		chan *IOClient
	ChRecv		chan *RecvPackage

	funAccept	OnAcceptCallback
	funClose	OnCloseCallback
	funRecv		OnRecvCallback
	funTimer	OnTimerCallback

	acceptCount	uint32
	closeCount	uint32
	recvCount	uint32
	sendCount	uint32
}

func NewServerSocket(acceptcb OnAcceptCallback, closecb OnCloseCallback, recvcb OnRecvCallback, timercb OnTimerCallback) *ServerSocket {
	return &ServerSocket {
		ClientList:	make(map[*IOClient]bool),
		ChAccept:	make(chan *IOClient),
		ChClose:	make(chan *IOClient),
		ChRecv:		make(chan *RecvPackage),
		funAccept:	acceptcb,
		funClose:	closecb,
		funRecv:	recvcb,
		funTimer:	timercb,
		acceptCount:0,
		closeCount:	0,
		recvCount:	0,
		sendCount:	0,
	}
}

// 初始化
func (s *ServerSocket) Init(addr string) bool {

	http.HandleFunc("/",
		func(res http.ResponseWriter, r *http.Request) {
			ServerAcceptHandle(s, res, r)
		})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		flog.GetInstance().Panic(err)
		return false
	}

	return true
}

func ServerAcceptHandle(s *ServerSocket, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		flog.GetInstance().Panic(err)
		return
	}

	client := newIOClient(s, conn)

	s.ChAccept <- client

	go client.goRead()
	go client.goWrite()
}

func (s *ServerSocket) Run() {
	ticker := time.NewTicker(timerWait)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case client := <- s.ChAccept:
			s.ClientList[client] = true
			s.funAccept(client)
			s.acceptCount++

		case client := <- s.ChClose:
			_, ok := s.ClientList[client]
			if ok {
				delete(s.ClientList, client)
				close(client.ChSend)
				s.funClose(client)
				s.closeCount++
				break
			}

		case pack := <- s.ChRecv:
			s.funRecv(pack.Client, pack.buf)
			s.recvCount++

		case <- ticker.C:
			s.funTimer()
			flog.GetInstance().Infof("Timer:accept=%d,close=%d,recv=%d,send=%d", s.acceptCount, s.closeCount, s.recvCount, s.sendCount)
			s.acceptCount = 0
			s.closeCount = 0
			s.recvCount = 0
			s.sendCount = 0
		}
	}
}