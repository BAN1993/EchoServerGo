package main

import (
	"fmt"
	"os"
	"time"

	"EchoServerGo/flog"
	"EchoServerGo/protocol"
)

func main() {
	cfg := Configs{}
	err := cfg.Load("ServerConfig.ini")
	if err != nil {
		fmt.Println("Load config err:", err)
		os.Exit(1)
	}

	err = flog.InitLog(cfg.logfilename, cfg.loglevel)
	if err != nil {
		fmt.Println("InitLog error:", err)
		os.Exit(1)
	}

	s := NewServerSocket(
		acceptCallback,
		closeCallback,
		recvCallback,
		onTimer)

	go s.Run()
	s.Init(cfg.listenaddr)
}

func acceptCallback(client *IOClient) {
	//fmt.Println("accept a client")
}

func closeCallback(client *IOClient) {
	//fmt.Println("a client close")
}

func recvCallback(client *IOClient, buf []byte) {
	var head protocol.Basebyte
	if head.GetHeadAndAttach(buf) {
		switch head.Xyid {
		case protocol.XYID_HEARTBEAT:
			msg_Heartbeat(client, buf)
			break
		default:
			// TODO
			break
		}
	} else {
		flog.GetInstance().Warnf("GetHeadAndAttach error")
	}
}

func onTimer() {

}

func msg_Heartbeat(client *IOClient, buf []byte) {
	var req protocol.HeartBeat
	if req.GetHeadAndAttach(buf) {
		req.Decode()
		flog.GetInstance().Debugf("recv:timestamp=%d", req.Timestamp)

		// 发送回包
		var resp protocol.HeartBeat
		resp.Timestamp = uint32(time.Now().Unix())
		flog.GetInstance().Debugf("send:timestamp=%d", resp.Timestamp)
		sendbuf := resp.Encode()
		client.ChSend <- sendbuf
	} else {
		flog.GetInstance().Warnf("GetHeadAndAttach error")
	}
}