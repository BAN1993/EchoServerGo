package main

import (
	"EchoServerGo/flog"
	"sync"
	"sync/atomic"
	"time"
)

// awk -F[=,\(] '{print $2,$4,$6,$8}' Client_***.log

type Monitor struct {
	links    int32
	linkerrs int32
	msgCount uint64
	useTime  uint64
}

// 单例
var once sync.Once
var gMonitor *Monitor
func GetMonitor() *Monitor {
	once.Do(func() {
		gMonitor = new(Monitor)
	})
	return gMonitor
}

func (m* Monitor) Run() {
	for {
		connects := m.GetConnections()
		errors := m.GetErrors()
		msg := m.restoreMsg()
		usetime := m.restoreUseTime()
		var interval float32 = 0.0
		if usetime != 0 {
			interval = float32(usetime) / float32(msg)
		}
		flog.GetInstance().Infof("Links=%d,Errors=%d,MsgCount=%d,UseTime=%d,Interval=%0.2f(ms)",
			connects, errors, msg, usetime, interval)
		time.Sleep(time.Second*1)
	}
}

func (m* Monitor) AddConnection() {
	atomic.AddInt32(&m.links, 1)
}

func (m* Monitor) DelConnection() {
	atomic.AddInt32(&m.links, -1)
}

func (m* Monitor) GetConnections() (int32) {
	return atomic.LoadInt32(&m.links)
}

func (m* Monitor) AddErrors() {
	atomic.AddInt32(&m.linkerrs, 1)
}

func (m* Monitor) GetErrors() (int32) {
	return atomic.LoadInt32(&m.linkerrs)
}

func (m* Monitor) AddMsg(n uint64) {
	atomic.AddUint64(&m.msgCount, n)
}

func (m* Monitor) restoreMsg() (uint64) {
	r := atomic.SwapUint64(&m.msgCount, 0)
	return r
}

func (m* Monitor) AddUseTime(len uint64) {
	atomic.AddUint64(&m.useTime, len)
}

func (m* Monitor) restoreUseTime() (uint64) {
	r := atomic.SwapUint64(&m.useTime, 0)
	return r
}