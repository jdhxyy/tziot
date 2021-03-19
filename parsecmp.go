// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// CMP协议解析处理
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/knock"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/utz"
)

func init() {
	standardLayerRegisterRxObserver(cmpDealStandardLayerRx)
}

func cmpDealStandardLayerRx(data []uint8, standardHeader *utz.StandardHeader, pipe uint64) {
	if standardHeader.DstIA != localIA || standardHeader.NextHead != utz.HeaderCmp {
		return
	}
	payload := utz.FlpFrameToBytes(data)
	if payload == nil {
		lagan.Warn(tag, "parse cmp failed.flp frame to bytes failed")
		return
	}

	if len(payload) == 0 {
		lagan.Warn(tag, "parse cmp failed.payload len is wrong:%d", len(payload))
		return
	}
	knock.Call(utz.HeaderCmp, uint16(payload[0]), payload[1:])
}
