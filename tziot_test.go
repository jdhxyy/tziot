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

	pipe := BindPipeNet(0x2141000000000401, "abc123", "192.168.1.119", 12021)
	if pipe == 0 {
		fmt.Println("bind pipe net failed")
		return
	}
	fmt.Println(pipe)

	for {
		if IsConn() == false {
			time.Sleep(time.Second)
			continue
		}
		resp, err := Call(pipe, 0x2141000000000004, 1, 1000, nil)
		fmt.Println("call", string(resp), err)
		time.Sleep(10 * time.Second)
	}

	select {}
}
