// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 管道操作
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/tzaccess"
	"net"
	"sync"
)

const (
	pipeCore = 1
)

// SendFunc 发送函数
type SendFunc func(data []uint8)

// IsAllowSendFunc 是否允许发送
type IsAllowSendFunc func() bool

type tApi struct {
	send        SendFunc
	isAllowSend IsAllowSendFunc
}

var pipes map[uint64]tApi
var pipeNum uint64 = pipeCore
var once sync.Once

var corePipeSend tzaccess.SendFunc = nil
var corePipeIsAllowSend tzaccess.IsAllowSendFunc = nil

func init() {
	pipes = make(map[uint64]tApi)
}

// BindPipeCore 绑定核心网管道.绑定成功后返回管道号,管道号如果是0表示绑定失败
// 注意:核心网管道只能绑定一个
func BindPipeCore(send tzaccess.SendFunc, isAllowSend tzaccess.IsAllowSendFunc) (pipe uint64) {
	corePipeSend = send
	corePipeIsAllowSend = isAllowSend
	tzaccess.Load(localIA, localPwd, send, isAllowSend)
	return pipeCore
}

// PipeCoreReceive 核心网管道接收
// 用户在管道中接收到数据时需回调本函数
func PipeCoreReceive(data []uint8, addr *net.UDPAddr) {
	tzaccess.Receive(data, addr)
	if tzaccess.IsConn() {
		standardLayerRx(pipeCore, data)
	}
}

// BindPipe 绑定管道.绑定成功后返回管道号
func BindPipe(send SendFunc, isAllowSend IsAllowSendFunc) (pipe uint64) {
	pipe = getPipeNum()

	var api tApi
	api.send = send
	api.isAllowSend = isAllowSend

	pipes[pipe] = api
	return pipe
}

func getPipeNum() uint64 {
	pipeNum++
	return pipeNum
}

// PipeReceive 管道接收.pipe是发送方的管道号
// 如果是用户自己绑定管道,则在管道中接收到数据需回调本函数
func PipeReceive(pipe uint64, data []uint8) {
	standardLayerRx(pipe, data)
}

func pipeIsAllowSend(pipe uint64) bool {
	if tzaccess.IsConn() == false {
		return false
	}

	if pipe == pipeCore {
		return corePipeIsAllowSend()
	}

	v, ok := pipes[pipe]
	if ok == false {
		return false
	}
	return v.isAllowSend()
}

func pipeSend(pipe uint64, data []uint8) {
	if pipe == pipeCore {
		addr := tzaccess.GetParentAddr()
		if addr == nil {
			return
		}
		corePipeSend(data, addr)
	}

	v, ok := pipes[pipe]
	if ok == false {
		return
	}
	v.send(data)
}

// BindPipeNet 绑定网络管道.绑定成功后返回管道号
func BindPipeNet(ia uint64, pwd string, ip string, port int) (pipe uint64) {
	ConfigLocalParam(ia, pwd)

	addr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}
	listener, err := net.ListenUDP("udp", &addr)
	if err != nil {
		lagan.Error(tag, "bind pipe net failed:%v", err)
		return 0
	}

	go netRxThread(listener)

	return BindPipeCore(
		func(data []uint8, addr *net.UDPAddr) {
			_, err := listener.WriteToUDP(data, addr)
			if err != nil {
				lagan.Error(tag, "udp send error:%v addr:%v", err, addr)
				return
			}
			lagan.Info(tag, "udp send:addr:%v len:%d", addr, len(data))
			lagan.PrintHex(tag, lagan.LevelDebug, data)
		},
		func() bool {
			return true
		})
}

func netRxThread(listener *net.UDPConn) {
	data := make([]uint8, frameMaxLen)
	for {
		num, addr, err := listener.ReadFromUDP(data)
		if err != nil {
			lagan.Error(tag, "listen pipe net failed:%v", err)
			continue
		}
		if num <= 0 {
			continue
		}
		lagan.Info(tag, "udp rx:%v len:%d", addr, num)
		lagan.PrintHex(tag, lagan.LevelDebug, data[:num])
		PipeCoreReceive(data[:num], addr)
	}
}
