// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 连接父路由
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/knock"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
	"time"
)

// 最大连接次数.超过连接次数这回清除父路由IA地址,重连父路由
const connNumMax = 3

var connNum = 0

func init() {
	knock.Register(utz.HeaderCmp, utz.CmpMsgTypeAckConnectParentRouter, dealAckConnectParentRouter)
	go connThread()
	go connTimeout()
}

// dealAckConnectParentRouter 处理应答连接帧
// 返回值是应答数据和应答标志.应答标志为false表示不需要应答
func dealAckConnectParentRouter(req []uint8, params ...interface{}) ([]uint8, bool) {
	if len(req) == 0 {
		lagan.Warn(tag, "deal conn failed.payload len is wrong:%d", len(req))
		return nil, false
	}

	j := 0
	if req[j] != 0 {
		lagan.Warn(tag, "deal conn failed.error code:%d", req[j])
		return nil, false
	}
	j++

	if len(req) != 2 {
		lagan.Warn(tag, "deal conn failed.payload len is wrong:%d", len(req))
		return nil, false
	}

	connNum = 0
	parent.isConn = true
	parent.cost = req[j]
	parent.timestamp = time.Now().Unix()
	lagan.Info(tag, "conn success.parent ia:0x%x cost:%d", parent.ia, parent.cost)
	return nil, false
}

func connThread() {
	for {
		// 如果网络通道不开启则无需连接
		if pipeIsAllowSend(pipeNet) == false {
			time.Sleep(time.Second)
			continue
		}

		if parent.ia != utz.IAInvalid {
			connNum += 1
			if connNum > connNumMax {
				connNum = 0
				parent.ia = utz.IAInvalid
				lagan.Warn(tag, "conn num is too many!")
				continue
			}
			lagan.Info(tag, "send conn frame")
			sendConnFrame()
		}

		if parent.ia == utz.IAInvalid {
			time.Sleep(time.Second)
		} else {
			time.Sleep(connInterval * time.Second)
		}
	}
}

func sendConnFrame() {
	var securityHeader utz.SimpleSecurityHeader
	securityHeader.NextHead = utz.HeaderCmp
	securityHeader.Pwd = localPwd
	payload := utz.SimpleSecurityHeaderToBytes(&securityHeader)

	var body []uint8
	body = append(body, utz.CmpMsgTypeConnectParentRouter)
	// 前缀长度
	body = append(body, 64)
	// 子膜从机固定单播地址
	body = append(body, make([]uint8, utz.IALen)...)
	// 开销值
	body = append(body, 0)
	body = utz.BytesToFlpFrame(body, true, 0)

	payload = append(payload, body...)

	var header utz.StandardHeader
	header.Version = utz.ProtocolVersion
	header.FrameIndex = utz.GenerateFrameIndex()
	header.PayloadLen = uint16(len(payload))
	header.NextHead = utz.HeaderSimpleSecurity
	header.HopsLimit = 0xff
	header.SrcIA = localIA
	header.DstIA = coreIA

	standardLayerSend(payload, &header, parent.pipe)
}

func connTimeout() {
	for {
		if parent.ia == utz.IAInvalid || parent.isConn == false {
			time.Sleep(time.Second)
			continue
		}
		if time.Now().Unix()-parent.timestamp > connTimeoutMax {
			parent.ia = utz.IAInvalid
			parent.isConn = false
		}
		time.Sleep(time.Second)
	}
}

// IsConn 是否连接核心网
func IsConn() bool {
	return parent.ia != utz.IAInvalid && parent.isConn
}
