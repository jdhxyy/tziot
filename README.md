# 海萤物联网教程：Go SDK

欢迎前往社区交流：[海萤物联网社区](http://www.ztziot.com)

[在线文档地址](https://jdhxyy.github.io/tziot)

## 简介
此SDK适用于Go 1.12.5版本及以上版本。使用此SDK可以让节点连接海萤物联网，与其他节点通信。节点可以是终端设备，也可以是个人电脑上的程序，可以使用此SDK与其他节点通信，并可以以服务的形式开放自己的能力。

Go SDK是海萤物联网的SDK之一，主要有如下功能：

- 连接海萤物联网
- 与其他节点通信
- 以服务的形式开放节点自身的能力

## 特点
基于此SDK可以极大降低物联网的开发门槛，可以实现：
- 一行代码连接网络
- 一行代码开放服务
- 一行代码通信

## 开源
- [github上的项目地址](https://github.com/jdhxyy/tziot)
- [gitee上的项目地址](https://gitee.com/jdhxyy/tziot)

## 安装
推荐使用go mod：github.com/jdhxyy/dcom

安装好后在项目中即可导入使用：
```go
import "https://github.com/jdhxyy/tziot"
```

导入其他相关的包：
```go
import "https://github.com/jdhxyy/dcom"
```

## 背景知识
- [海萤物联网教程：IA地址格式及地址申请方法](https://blog.csdn.net/jdh99/article/details/115340195)

- [海萤物联网教程：物联网RPC框架Go DCOM](https://blog.csdn.net/jdh99/article/details/115331198)

## API
```go
// BindPipeNet 绑定网络管道.绑定成功后返回管道号
func BindPipeNet(ia uint64, pwd string, ip string, port int) (pipe uint64, err error)

// BindPipe 绑定管道.绑定成功后返回管道号
func BindPipe(ia uint64, send SendFunc, isAllowSend IsAllowSendFunc) (pipe uint64)

// PipeReceive 管道接收.pipe是发送方的管道号
// 如果是用户自己绑定管道,则在管道中接收到数据需回调本函数
func PipeReceive(pipe uint64, data []uint8)

// IsConn 是否连接核心网
func IsConn() bool

// Call RPC同步调用
// timeout是超时时间,单位:ms.为0表示不需要应答
// 返回值是应答字节流和错误码.错误码非0表示调用失败
func Call(pipe uint64, dstIA uint64, rid int, timeout int, req []uint8) ([]uint8, int)

// Register 注册服务回调函数
func Register(rid int, callback dcom.CallbackFunc)

// ConfigCoreParam 配置核心网参数
func ConfigCoreParam(ia uint64, ip string, port int)

// ConfigDComParam 配置dcom参数
// retryNum: 重发次数
// retryInterval: 重发间隔.单位:ms
func ConfigDComParam(retryNum, retryInterval int)
```

- 数据结构

```go
// SendFunc 发送函数.dstPipe:目标管道号
type SendFunc func(dstPipe uint64, data []uint8)

// IsAllowSendFunc 是否允许发送
type IsAllowSendFunc func() bool

// dcom中定义
// CallbackFunc 注册DCOM服务回调函数
// 返回值是应答和错误码.错误码为0表示回调成功,否则是错误码
type CallbackFunc func(pipe uint64, srcIA uint64, req []uint8) ([]uint8, int)
```

### 默认参数
当前默认的参数：

参数|值
---|---
DCOM重发次数|5
DCOM重发间隔|500ms

调用ConfigDComParam函数可以修改DCOM参数。

ConfigCoreParam函数可以修改海萤物联网平台默认地址，使用默认值即可，不需要调用函数修改。

### 绑定管道
tziot包中封装了dcom包，在绑定管道时会初始化DCOM。

tziot中调用BindPipe函数可以绑定自定义管道，如果使用自定义管道，则需应用中调用PipeReceive函数将接收到的数据发送给tziot包。

绑定网络管道是绑定管道的一个特例，如果节点可以直接连接互联网（比如使用以太网或者wifi），则调用BindPipeNet函数即可，不需要使用BindPipe函数和PipeReceive函数。

- 示例：绑定网络管道，节点地址是0x2140000000000101，本地端口号是12021
```go
pipe, err := BindPipeNet(0x2140000000000101, pwd, "0.0.0.0", 12021)
```
返回的是管道号pipe。后续使用Call函数与其他节点通信，需要使用此管道号。

绑定管道后sdk会自动连接海萤物联网，可以调用IsConn函数查看连接是否成功。

### 注册服务
节点可以通过注册服务开放自身的能力。

```go
// Register 注册服务回调函数
func Register(rid int, callback dcom.CallbackFunc)
```

注册函数中，每个服务号（rid），都可以绑定一个服务。

- 示例：假设节点2140::101是智能插座，提供控制和读取开关状态两个服务：

```go
tziot.Register(1, controlService)
tziot.Register(2, getStateService)

// controlService 控制开关服务
// 返回值是应答和错误码.错误码为0表示回调成功,否则是错误码
func controlService(pipe uint64, srcIA uint64, req []uint8) ([]uint8, int) {
	if req[0] == 0 {
		off()
	} else {
		on()
	}
	return nil, dcom.SystemOK
}

// getStateService 读取开关状态服务
// 返回值是应答和错误码.错误码为0表示回调成功,否则是错误码
func getStateService(pipe uint64, srcIA uint64, req []uint8) ([]uint8, int) {
	return []uint8{state()}, dcom.SystemOK
}
```

### 调用目的节点服务
```go
// Call RPC同步调用
// timeout是超时时间,单位:ms.为0表示不需要应答
// 返回值是应答字节流和错误码.错误码非0表示调用失败
func Call(pipe uint64, dstIA uint64, rid int, timeout int, req []uint8) ([]uint8, int)
```

同步调用会在获取到结果之前阻塞。节点可以通过同步调用，调用目标节点的函数或者服务。timeout字段是超时时间，单位是毫秒。如果目标节点超时未回复，则会调用失败。如果超时时间填0，则表示不需要目标节点回复。

- 示例：2141::102节点控制智能插座2141::101开关状态为开

```go
resp, errCode := tziot.Call(1, 0x2140000000000101, 3000, []uint8{1})
```

- 示例：2141::102节点读取智能插座2141::101开关状态

```go
resp, errCode := tziot.Call(2, 0x2140000000000101, 3000, nil)
if errCode == dcom.SystemOK {
	fmt.println("开关状态:", resp[0])
}
```

## 请求和应答数据格式
建议使用结构体来通信。详情可参考： [海萤物联网教程：物联网RPC框架Go DCOM](https://blog.csdn.net/jdh99/article/details/115331198) 中的数据格式章节。

## 完整示例
示例以与海萤物联网中的ntp服务通信为例。

### ntp服务器开源地址
- [github上的ntp服务项目地址](https://github.com/jdhxyy/ntp)
- [gitee上的ntp服务项目地址](https://gitee.com/jdhxyy/ntp)

### ntp服务介绍
[海萤物联网ntp服务上线](https://blog.csdn.net/jdh99/article/details/115368543)

ntp服务器地址：
```text
0x2141000000000404
```

当前提供两个服务：

服务号|服务
---|---
1|读取时间1
2|读取时间2.返回的是结构体

#### 读取时间服务1
- CON请求：空或者带符号的1个字节。

当CON请求为空时，则默认为读取的是北京时间（时区8）。

也可以带1个字节表示时区号。这个字节是有符号的int8。

小技巧，可以使用0x100减去正值即负值。比如8对应的无符号数是0x100-8=248。

- ACK应答：当前时间的字符串

当前时间字符串的格式：2006-01-02 15:04:05 -0700 MST

#### 读取时间服务2.返回的是结构体
- CON请求：格式与读取时间服务1一致

- ACK应答：
```c
struct {
    // 时区
    uint8 TimeZone
    uint16 Year
    uint8 Month
    uint8 Day
    uint8 Hour
    uint8 Minute
    uint8 Second
    // 星期
    uint8 Weekday
}
```

### 开放服务示例
```go
// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 网络校时服务
// Authors: jdh99 <jdh821@163.com>

package main

import (
	"github.com/jdhxyy/dcom"
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/tziot"
	"ntp/config"
	"time"
)

const tag = "ntp"

// 应用错误码
const (
	// 内部错误
	errorCodeInternalError = 0x40
	// 接收格式错误
	errorCodeRxFormat = 0x41
)

// rid号
const (
	// 读取时间.返回的是字符串
	ridGetTime1 = 1
	// 读取时间.返回的是结构体
	ridGetTime2 = 2
)

// ACK格式
type AckRidGetTime2 struct {
	// 时区
	TimeZone uint8
	Year     uint16
	Month    uint8
	Day      uint8
	Hour     uint8
	Minute   uint8
	Second   uint8
	// 星期
	Weekday uint8
}

func main() {
	err := lagan.Load(0)
	if err != nil {
		panic(err)
	}
	lagan.EnableColor(true)
	lagan.SetFilterLevel(lagan.LevelInfo)

	_, err = tziot.BindPipeNet(config.LocalIA, config.LocalPwd, config.LocalIP, config.LocalPort)
	if err != nil {
		panic(err)
		return
	}
	tziot.Register(ridGetTime1, ntpService1)
	tziot.Register(ridGetTime2, ntpService2)

	select {}
}

// ntpService1 校时服务
// 返回值是应答和错误码.错误码为0表示回调成功,否则是错误码
func ntpService1(pipe uint64, srcIA uint64, req []uint8) ([]uint8, int) {
	addr := dcom.PipeToAddr(pipe)

	var timeZone int
	if len(req) == 0 {
		timeZone = 8
	} else if len(req) == 1 {
		timeZone = int(int8(req[0]))
	} else {
		lagan.Warn(tag, "addr:%v ia:0x%x ntp failed.len is wrong:%d", addr, srcIA, len(req))
		return nil, errorCodeRxFormat
	}

	t := getTime(timeZone)
	lagan.Info(tag, "addr:%v ia:0x%x ntp time:%v", addr, srcIA, t)
	return []uint8(t.Format("2006-01-02 15:04:05 -0700 MST")), 0
}

func getTime(timeZone int) time.Time {
	t := time.Now().UTC()
	secondsEastOfUTC := int((time.Duration(timeZone) * time.Hour).Seconds())
	loc := time.FixedZone("CST", secondsEastOfUTC)
	t = t.In(loc)
	return t
}

// ntpService2 校时服务
// 返回值是应答和错误码.错误码为0表示回调成功,否则是错误码
func ntpService2(pipe uint64, srcIA uint64, req []uint8) ([]uint8, int) {
	addr := dcom.PipeToAddr(pipe)

	var timeZone int
	if len(req) == 0 {
		timeZone = 8
	} else if len(req) == 1 {
		timeZone = int(int8(req[0]))
	} else {
		lagan.Warn(tag, "addr:%v ia:0x%x ntp failed.len is wrong:%d", addr, srcIA, len(req))
		return nil, errorCodeRxFormat
	}

	t := getTime(timeZone)
	lagan.Info(tag, "addr:%v ia:0x%x ntp time:%v", addr, srcIA, t)

	var ack AckRidGetTime2
	ack.TimeZone = uint8(timeZone)
	ack.Year = uint16(t.Year())
	ack.Month = uint8(t.Month())
	ack.Day = uint8(t.Day())
	ack.Hour = uint8(t.Hour())
	ack.Minute = uint8(t.Minute())
	ack.Second = uint8(t.Second())
	ack.Weekday = uint8(t.Weekday())

	data, err := dcom.StructToBytes(ack)
	if err != nil {
		lagan.Error(tag, "addr:%v ia:0x%x ntp failed.struct to bytes error:%v", addr, srcIA, err)
		return nil, errorCodeInternalError
	}
	return data, 0
}
```

### 读取时间服务1
节点2141::401读取ntp服务器的服务1，并打印时间字符串。

```go
package main

import (
    "fmt"
    "github.com/jdhxyy/tziot"
)

func main() {
    pipe, _ := tziot.BindPipeNet(0x2141000000000401, "abc123", "192.168.1.119", 12021)
    for tziot.IsConn() == false{}
    resp, err := tziot.Call(pipe, 0x2141000000000004, 1, 3000, []uint8{8})
    fmt.Println("err:", err, "time:", string(resp))
}
```

输出结果：
```text
err: 0 time: 2021-04-01 09:05:33 +0800 CST
```

### 读取时间服务2
ntp服务器的2号是结构体形式的时间。

```go
package main

import (
"fmt"
    "github.com/jdhxyy/dcom"
    "github.com/jdhxyy/tziot"
)

// ACK格式
type AckRidGetTime struct {
    // 时区
    TimeZone uint8
    Year     uint16
    Month    uint8
    Day      uint8
    Hour     uint8
    Minute   uint8
    Second   uint8
    // 星期
    Weekday uint8
}

func main() {
    pipe, _ := tziot.BindPipeNet(0x2141000000000401, "abc123", "192.168.1.119", 12021)
    for tziot.IsConn() == false{}
    resp, _ := tziot.Call(pipe, 0x2141000000000004, 2, 3000, []uint8{8})
    
    var ack AckRidGetTime
    _ = dcom.BytesToStruct(resp, &ack)
    fmt.Println(ack)
}
```

输出结果：
```text
{8 2021 4 1 9 7 20 4}
```
