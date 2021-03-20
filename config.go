// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 配置文件
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/dcom"
	"net"
)

const (
	tag = "tziot"

	// 最大帧字节数
	frameMaxLen = 4096

	protocolNum = 0

	// 连接间隔.单位:s
	connInterval = 30
	// 连接超时时间.单位:s
	connTimeoutMax = 120
)

// 本机单播地址
var localIA uint64
var localPwd string

// 核心网参数
// todo
var coreIA uint64 = 0x2141000000000002
var coreIP = "192.168.1.119"
var corePort = 12914
var corePipe uint64

// dcom参数
// dcom重发次数
var dcomRetryNum = 5

// dcom重发间隔.单位:ms
var dcomRetryInterval = 500

func init() {
	corePipe = dcom.AddrToPipe(&net.UDPAddr{IP: net.ParseIP(coreIP), Port: corePort})
}

// ConfigCoreParam 配置核心网参数
func ConfigCoreParam(ia uint64, ip string, port int) {
	coreIA = ia
	coreIP = ip
	corePort = port
	corePipe = dcom.AddrToPipe(&net.UDPAddr{IP: net.ParseIP(coreIP), Port: corePort})
}

// ConfigDComParam 配置dcom参数
// retryNum: 重发次数
// retryInterval: 重发间隔.单位:ms
func ConfigDComParam(retryNum, retryInterval int) {
	if retryNum > 0 {
		dcomRetryNum = retryNum
	}
	if retryInterval > 0 {
		dcomRetryInterval = retryInterval
	}
}
