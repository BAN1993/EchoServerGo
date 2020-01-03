package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"EchoServerGo/flog"
	"EchoServerGo/protocol"
)

// top -H -d 0.5 -p `pgrep "Web|server" |xargs perl -e "print join ',',@ARGV"`

func runOneClient(looptimes uint32, waittime uint32, addr string, wg *sync.WaitGroup) {
	u := url.URL{Scheme:"ws", Host:addr, Path:""}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		flog.GetInstance().Errorf("connect error,err=", err)
		GetMonitor().AddErrors()
		if c != nil {
			c.Close()
		}
		wg.Done()
		return
	}

	GetMonitor().AddConnection()
	defer func() {
		GetMonitor().DelConnection()
		c.Close()
		wg.Done()
	}()

	for nowloop := 0; looptimes == 0 || nowloop < int(looptimes);  nowloop++ {

		begin := time.Now().UnixNano() / 1e6

		// 发送心跳
		var resp protocol.HeartBeat
		resp.Timestamp = uint32(time.Now().Unix())
		flog.GetInstance().Debugf("send:timestamp=%d", resp.Timestamp)
		sendbuf := resp.Encode()
		err := c.WriteMessage(websocket.BinaryMessage, sendbuf)
		if err != nil {
			flog.GetInstance().Error("write err:", err)
			return
		}

		// 收到心跳
		_, recvbuf, err := c.ReadMessage()
		if err != nil {
			flog.GetInstance().Error("read err:", err)
			return
		}
		var req protocol.HeartBeat
		if(req.GetHeadAndAttach(recvbuf)) {
			req.Decode()
			flog.GetInstance().Debugf("recv:timestamp=%d", req.Timestamp)
		} else {
			flog.GetInstance().Errorf("recv a error message")
			return
		}

		use := time.Now().UnixNano() / 1e6 - begin
		GetMonitor().AddMsg(1)
		GetMonitor().AddUseTime(uint64(use))

		if waittime >0 {
			time.Sleep(time.Millisecond * time.Duration(waittime))
		}
	}

	_ = c.WriteMessage(websocket.CloseMessage, []byte{})
}

func main() {
	cfg := Configs{}
	err := cfg.Load("ClientConfig.ini")
	if err != nil {
		fmt.Println("Load config err:", err)
		os.Exit(1)
	}

	err = flog.InitLog(cfg.logfilename, cfg.loglevel)
	if err != nil {
		fmt.Println("InitLog error:", err)
		os.Exit(1)
	}

	go GetMonitor().Run()

	begin := time.Now().UnixNano()  / 1e6
	var wg sync.WaitGroup
	for i:=0; i < int(cfg.links) ; i++ {
		wg.Add(1)
		go runOneClient(cfg.loops, cfg.interval, cfg.host, &wg)
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
	end := time.Now().UnixNano()  / 1e6
	flog.GetInstance().Debug("usetimes=", (end-begin))
}
