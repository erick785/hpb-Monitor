package rpc

import (
	"github.com/hpb-project/go-hpb/common"
	"github.com/hpb-project/go-hpb/common/hexutil"
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
	ParentHash     common.Hash    `json:"parentHash"`
	Difficulty     string         `json:"difficulty"`
	Miner          common.Address `json:"miner"`
	HardwareRandom hexutil.Bytes  `json:"hardwareRandom"`
}

type Signers struct {
	Addresses map[common.Address]struct{} `json:"signers"`
}

type Node struct {
	NodeName       string
	NodeAddress    common.Address
	LockAmount     float64
	Country        string
	LocationDetail string
}
