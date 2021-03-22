// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 天泽物联网sdk
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/dcom"
	"github.com/jdhxyy/utz"
)

// Call RPC同步调用
// timeout是超时时间,单位:ms.为0表示不需要应答
// 返回值是应答字节流和错误码.错误码非0表示调用失败
func Call(pipe uint64, dstIA uint64, rid int, timeout int, req []uint8) ([]uint8, int) {
	if parent.ia == utz.IAInvalid || parent.isConn == false {
		// todo
		return nil, dcom.SystemErrorRxTimeout
	}
	if pipe >= pipeNet {
		pipe = parent.pipe
	}
	return dcom.Call(protocolNum, pipe, dstIA, rid, timeout, req)
}

// Register 注册服务回调函数
func Register(rid int, callback dcom.CallbackFunc) {
	dcom.Register(protocolNum, rid, callback)
}
