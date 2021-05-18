package main

import (
	"github.com/hpb-project/go-hpb/common"
)

var (
	hpbScanURL = "hpbscan.org/HpbScan/node/list"
	hpbNodeURL = "hpbnode.com"
)

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int64         `json:"id"`
}

func NewRPCRequest(method string, params []interface{}) *RPCRequest {
	return &RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      int64(1),
	}
}

type BlockResult struct {
	Miner          common.Address
	HardwareRandom []byte
}

func GetBlockMinerAndHardwareRandom(number uint64) (*BlockResult, error) {

	return nil, nil
}

func GetHpbNodeSnapSigners(number uint64) ([]common.Address, error) {

	return nil, nil

}

type Node struct {
	NodeName string

	NodeAddress common.Address

	LockAmount string
	Country    string

	LocationDetail string
}

func GetNodeListFromHpbScan() ([]Node, error) {
	return nil, nil

}

func main() {

}
