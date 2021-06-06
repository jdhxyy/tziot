// Copyright 2021-2021 The jdh99 Authors. All rights reserved.
// 路由表.存储代理节点信息
// Authors: jdh99 <jdh821@163.com>

package tziot

import (
	"github.com/jdhxyy/lagan"
	"github.com/jdhxyy/skiplist"
	"github.com/jdhxyy/utz"
	"sync"
	"time"
)

// 路由条目超时时间.单位:小时
const rtTimeout = 24

// rtItem 路由条目
type rtItem struct {
	ia      uint64
	agentIA uint64

	// 更新时间戳.单位:s
	timestamp int64
}

var rt *skiplist.SkipList

// 资源锁
var rtLock sync.RWMutex

func init() {
	rt = skiplist.New(skiplist.Uint64)
	go rtCheckTimeout()
}

func rtCheckTimeout() {
	var item *rtItem
	var elem *skiplist.Element
	var elemNext *skiplist.Element
	var now int64

	timeout := int64(rtTimeout) * 3600
	for {
		rtLock.Lock()

		if elem == nil || (elem.Prev() == nil && elem.Next() == nil) {
			elem = rt.Front()
		}
		now = time.Now().Unix()

		for i := 0; i < 100; i++ {
			if elem == nil {
				break
			}
			elemNext = elem.Next()

			item = elem.Value.(*rtItem)
			if now-item.timestamp > timeout {
				lagan.Warn(tag, "rt item timeout.ia:0x%x", item.agentIA)
				rt.Remove(item.ia)
			}

			elem = elemNext
		}

		rtLock.Unlock()

		time.Sleep(time.Minute)
	}
}

// rtAdd 增加代理节点地址.条目如果存在则会更新
func rtAdd(ia uint64, agentIA uint64) {
	if ia == utz.IAInvalid || agentIA == utz.IAInvalid {
		return
	}

	rtLock.Lock()
	defer rtLock.Unlock()

	elem := rt.Get(ia)
	var value *rtItem

	if elem == nil {
		value = new(rtItem)
		rt.Set(ia, value)
	} else {
		value = elem.Value.(*rtItem)
	}
	value.ia = ia
	value.agentIA = agentIA
	value.timestamp = time.Now().Unix()
}

// rtDelete 删除代理节点地址
func rtDelete(ia uint64) {
	rtLock.Lock()
	rt.Remove(ia)
	rtLock.Unlock()
}

// rtFind 寻找代理地址.如果条目不存在则返回0
func rtFind(ia uint64) uint64 {
	rtLock.RLock()
	defer rtLock.RUnlock()

	elem := rt.Get(ia)
	if elem == nil {
		return utz.IAInvalid
	} else {
		return elem.Value.(*rtItem).agentIA
	}
}
