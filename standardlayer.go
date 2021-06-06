// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 标准层处理模块
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
)

// standardLayerRxCallback 接收回调函数
type standardLayerRxCallback func(data []uint8, standardHeader *utz.StandardHeader, pipe uint64)

var standardLayerObservers []standardLayerRxCallback

// standardLayerRx 标准层接收
func standardLayerRx(pipe uint64, data []uint8) {
	header := getStandardHeader(data)
	if header == nil {
		return
	}

	// 处理代理头部
	offset := utz.NLv1HeadLen
	if header.NextHead == utz.HeaderAgent {
		agentHeader, num := utz.BytesToAgentHeader(data[offset:])
		if num == 0 {
			lagan.Warn(tag, "parse agent header failed")
			return
		}
		offset += num
		rtAdd(header.SrcIA, agentHeader.IA)
	} else {
		rtDelete(header.SrcIA)
	}
	notifyStandardLayerObservers(data[offset:], header, pipe)
}

func getStandardHeader(data []uint8) *utz.StandardHeader {
	header, offset := utz.BytesToStandardHeader(data)
	if header == nil || offset == 0 {
		lagan.Debug(tag, "get standard header failed:bytes to standard header failed")
		return nil
	}
	if header.Version != utz.ProtocolVersion {
		lagan.Debug(tag, "get standard header failed:protocol version is not match:%d", header.Version)
		return nil
	}
	if int(header.PayloadLen)+offset != len(data) {
		lagan.Debug(tag, "get standard header failed:payload len is not match:%d", header.PayloadLen)
		return nil
	}

	return header
}

func notifyStandardLayerObservers(data []uint8, standardHeader *utz.StandardHeader, pipe uint64) {
	n := len(standardLayerObservers)
	for i := 0; i < n; i++ {
		standardLayerObservers[i](data, standardHeader, pipe)
	}
}

// standardLayerRegisterRxObserver 注册接收观察者
func standardLayerRegisterRxObserver(callback standardLayerRxCallback) {
	standardLayerObservers = append(standardLayerObservers, callback)
}

// standardLayerSend 基于标准头部发送
func standardLayerSend(data []uint8, standardHeader *utz.StandardHeader, pipe uint64) {
	dataLen := len(data)
	if dataLen > frameMaxLen {
		lagan.Error(tag, "standard layer send failed!data len is too long:%d src ia:0x%x dst ia:0x%x", dataLen,
			standardHeader.SrcIA, standardHeader.DstIA)
		return
	}
	if standardHeader.PayloadLen != uint16(dataLen) {
		standardHeader.PayloadLen = uint16(dataLen)
	}

	// 判断是否有代理节点
	var routeHeaderBytes []uint8 = nil
	agentIA := rtFind(standardHeader.DstIA)
	if agentIA != utz.IAInvalid {
		routeHeader := utz.RouteHeader{NextHead: standardHeader.NextHead, RouteNum: 1, IsStrict: true, IAList: []uint64{agentIA}}
		routeHeaderBytes = utz.RouteHeaderToBytes(&routeHeader)

		standardHeader.NextHead = utz.HeaderRoute
		standardHeader.PayloadLen += uint16(len(routeHeaderBytes))
	}

	frame := utz.StandardHeaderToBytes(standardHeader)
	if routeHeaderBytes != nil {
		frame = append(frame, routeHeaderBytes...)
	}
	frame = append(frame, data...)
	pipeSend(pipe, frame)
}
