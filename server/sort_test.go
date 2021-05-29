package server

import (
	"fmt"
	"testing"

	"github.com/erick785/hpb-monitor/rpc"
	"github.com/hpb-project/go-hpb/common"
)

func TestSortMap(t *testing.T) {
	sm := newSortedMap(5)

	sm.Put(0, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})

	sm.Put(1, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})
	sm.Put(2, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})
	sm.Put(3, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})
	sm.Put(4, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})

	fmt.Println(sm.items)
	fmt.Println("------------------")

	sm.Put(5, map[common.Address]rpc.Node{
		common.StringToAddress("0"): rpc.Node{},
	})

	fmt.Println(sm.items)
}
