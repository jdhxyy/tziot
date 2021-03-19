// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 申请父路由
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/dcom"
	"github.com/jdhxyy/knock"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
	"net"
	"time"
)

// tParentInfo 父路由信息
type tParentInfo struct {
	ia        uint64
	pipe      uint64
	cost      uint8
	isConn    bool
	timestamp int64
}

var parent tParentInfo

func init() {
	knock.Register(utz.HeaderCmp, utz.CmpMsgTypeAssignSlaveRouter, dealAssignSlaveRouter)
	go applyThread()
}

// dealAssignSlaveRouter 处理分配从机帧
// 返回值是应答数据和应答标志.应答标志为false表示不需要应答
func dealAssignSlaveRouter(req []uint8, params ...interface{}) ([]uint8, bool) {
	if len(req) == 0 {
		lagan.Warn(tag, "deal apply failed.payload len is wrong:%d", len(req))
		return nil, false
	}

	j := 0
	if req[j] != 0 {
		lagan.Warn(tag, "deal apply failed.error code:%d", req[j])
		return nil, false
	}
	j++

	if len(req) != 16 {
		lagan.Warn(tag, "deal apply failed.payload len is wrong:%d", len(req))
		return nil, false
	}

	parent.ia = utz.BytesToIA(req[j : j+utz.IALen])
	j += utz.IALen

	ip := make([]uint8, 4)
	copy(ip, req[j:j+4])
	j += 4
	port := (int(req[j]) << 8) + int(req[j+1])
	j += 2
	addr := net.UDPAddr{IP: net.IPv4(ip[0], ip[1], ip[2], ip[3]), Port: port}
	parent.pipe = dcom.AddrToPort(&addr)

	lagan.Info(tag, "apply success.parent ia:0x%x addr:%v cost:%d", parent.ia, addr, req[j])
	return nil, false
}

func applyThread() {
	for {
		// 如果网络通道不开启则无需申请
		if pipeIsAllowSend(pipeNet) == false {
			time.Sleep(time.Second)
			continue
		}

		if isDComInit && parent.ia == utz.IAInvalid {
			lagan.Info(tag, "send apply frame")
			sendApplyFrame()
		}

		if isDComInit {
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(time.Second)
		}
	}
}

func sendApplyFrame() {
	var securityHeader utz.SimpleSecurityHeader
	securityHeader.NextHead = utz.HeaderCmp
	securityHeader.Pwd = localPwd
	payload := utz.SimpleSecurityHeaderToBytes(&securityHeader)

	var body []uint8
	body = append(body, utz.CmpMsgTypeRequestSlaveRouter)
	body = append(body, utz.IAToBytes(parent.ia)...)
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

	standardLayerSend(payload, &header, corePipe)
}
