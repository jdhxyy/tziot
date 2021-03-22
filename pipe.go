// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 管道操作
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"errors"
	"github.com/jdhxyy/dcom"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
	"net"
	"sync"
)

const (
	pipeNet = 0xffff
)

// SendFunc 发送函数.dstPipe:目标管道号
type SendFunc func(dstPipe uint64, data []uint8)

// IsAllowSendFunc 是否允许发送
type IsAllowSendFunc func() bool

type tApi struct {
	send        SendFunc
	isAllowSend IsAllowSendFunc
}

var pipes map[uint64]tApi
var pipeNum uint64
var once sync.Once
var isBindPipeNet = false

func init() {
	pipes = make(map[uint64]tApi)
}

// BindPipeNet 绑定网络管道.绑定成功后返回管道号
func BindPipeNet(ia uint64, pwd string, ip string, port int) (pipe uint64, err error) {
	if isBindPipeNet {
		return pipeNet, errors.New("already bind pipe net")
	}

	localPwd = pwd
	addr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}
	listener, err := net.ListenUDP("udp", &addr)
	if err != nil {
		lagan.Error(tag, "bind pipe net failed:%v", err)
		return 0, nil
	}

	isBindPipeNet = true
	bind(
		pipeNet,
		ia,
		func(pipe uint64, data []uint8) {
			addr := dcom.PipeToAddr(pipe)
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

	go netRxThread(listener, pipeNet)
	return pipeNet, nil
}

func bind(pipe uint64, ia uint64, send SendFunc, isAllowSend IsAllowSendFunc) {
	localIA = ia

	var api tApi
	api.send = send
	api.isAllowSend = isAllowSend

	pipes[pipe] = api
	once.Do(initDCom)
}

func netRxThread(listener *net.UDPConn, pipe uint64) {
	data := make([]uint8, frameMaxLen)
	for {
		num, addr, err := listener.ReadFromUDP(data)
		if err != nil {
			lagan.Error(tag, "listen pipe net failed:%v pipe:%d", err, pipe)
			continue
		}
		if num <= 0 {
			continue
		}
		lagan.Info(tag, "udp rx:%v len:%d", addr, num)
		lagan.PrintHex(tag, lagan.LevelDebug, data[:num])
		PipeReceive(dcom.AddrToPipe(addr), data[:num])
	}
}

// PipeReceive 管道接收.pipe是发送方的管道号
// 如果是用户自己绑定管道,则在管道中接收到数据需回调本函数
func PipeReceive(pipe uint64, data []uint8) {
	standardLayerRx(pipe, data)
}

// BindPipe 绑定管道.绑定成功后返回管道号
func BindPipe(ia uint64, send SendFunc, isAllowSend IsAllowSendFunc) (pipe uint64) {
	pipe = getPipeNum()
	bind(pipe, ia, send, isAllowSend)
	return pipe
}

func getPipeNum() uint64 {
	pipeNum++
	return pipeNum
}

func pipeIsAllowSend(pipe uint64) bool {
	var v tApi
	var ok bool

	if pipe >= pipeNet {
		v, ok = pipes[pipeNet]
	} else {
		v, ok = pipes[pipe]
	}

	if ok == false {
		return false
	}
	return v.isAllowSend()
}

func pipeSend(pipe uint64, data []uint8) {
	if pipe == 0 {
		return
	}

	var v tApi
	var ok bool

	if pipe >= pipeNet {
		v, ok = pipes[pipeNet]
	} else {
		v, ok = pipes[pipe]
	}
	if ok == false {
		return
	}

	if pipe == pipeNet {
		if parent.ia == utz.IAInvalid || parent.isConn == false {
			return
		}
		v.send(parent.pipe, data)
	} else {
		// 注意:发给核心网的帧也通过此处发送
		v.send(pipe, data)
	}
}
