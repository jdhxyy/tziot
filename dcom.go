// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// dcom操作
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/dcom"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
)

// dcom是否初始化
var isDComInit = false

// initDCom 初始化dcom.全局只应该初始化一次
func initDCom() {
	var param dcom.LoadParam
	param.BlockRetryMaxNum = dcomRetryNum
	param.BlockRetryInterval = dcomRetryInterval
	param.IsAllowSend = pipeIsAllowSend
	param.Send = dcomSend
	dcom.Load(&param)

	standardLayerRegisterRxObserver(dcomDealStandardLayerRx)
	isDComInit = true
}

func dcomSend(protocol int, pipe uint64, dstIA uint64, data []uint8) {
	flpFrame := utz.BytesToFlpFrame(data, true, 0)

	var header utz.StandardHeader
	header.Version = utz.ProtocolVersion
	header.FrameIndex = utz.GenerateFrameIndex()
	header.PayloadLen = uint16(len(flpFrame))
	header.NextHead = utz.HeaderFlp
	header.HopsLimit = 0xff
	header.SrcIA = localIA
	header.DstIA = dstIA

	standardLayerSend(flpFrame, &header, pipe)
}

// dcomDealStandardLayerRx 处理标准层回调函数
func dcomDealStandardLayerRx(data []uint8, standardHeader *utz.StandardHeader, pipe uint64) {
	if standardHeader.DstIA != localIA || standardHeader.NextHead != utz.HeaderFlp {
		return
	}
	body := utz.FlpFrameToBytes(data)
	if body == nil {
		lagan.Warn(tag, "flp frame to bytes failed!src ia:0x%x", standardHeader.SrcIA)
		return
	}
	dcom.Receive(protocolNum, pipe, standardHeader.SrcIA, body)
}
