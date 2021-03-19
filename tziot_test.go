package tziot

import (
	"fmt"
	"github.com/jdhxyy/lagan"
	"testing"
	"time"
)

func TestBindPipeNet(t *testing.T) {
	_ = lagan.Load(0)
	lagan.SetFilterLevel(lagan.LevelDebug)
	lagan.EnableColor(true)

	pipe, err := BindPipeNet(0x2141000000000401, "abc123", "192.168.1.119", 12021)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(pipe)

	for {
		if IsConn() == false {
			time.Sleep(time.Second)
			continue
		}
		resp, err := Call(pipe, 0x2141000000000402, 2, 1000, []uint8{5, 6, 7})
		fmt.Println("call", string(resp), err)
		time.Sleep(time.Second)
	}

	select {}
}
