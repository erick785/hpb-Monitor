package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/hpb-project/go-hpb/common"
	"github.com/hpb-project/go-hpb/common/hexutil"
)

func GetLastBlockNumber(endpoint string) (int64, error) {

	var (
		rpcRequest *RPCRequest
		body       []byte
	)

	rpcRequest = NewRPCRequest("hpb_blockNumber", []interface{}{})
	rpcPayload, err := json.Marshal(rpcRequest)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(rpcPayload)))
	if err != nil {
		return 0, err
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	type RPCBlockResult struct {
		ID      int           `json:"id"`
		JSONRPC string        `json:"jsonrpc"`
		Result  hexutil.Bytes `json:"result"`
	}

	rpcBlockResp := &RPCBlockResult{}
	err = json.Unmarshal(body, rpcBlockResp)
	return new(big.Int).SetBytes(rpcBlockResp.Result).Int64(), err
}

func GetBlockMinerAndHardwareRandom(number int64, endpoint string) (*BlockResult, error) {
	var (
		rpcRequest *RPCRequest
		body       []byte
	)

	hexBlockNumber := fmt.Sprintf("0x%x", number)
	rpcRequest = NewRPCRequest("hpb_getBlockByNumber", []interface{}{hexBlockNumber, true})
	rpcPayload, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(rpcPayload)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type RPCBlockResult struct {
		ID      int         `json:"id"`
		JSONRPC string      `json:"jsonrpc"`
		Result  BlockResult `json:"result"`
	}

	rpcBlockResp := &RPCBlockResult{}
	err = json.Unmarshal(body, rpcBlockResp)

	return &rpcBlockResp.Result, err
}

func GetHpbNodeSnapSigners(number int64, endpoint string) ([]common.Address, error) {

	var (
		rpcRequest *RPCRequest
		body       []byte
	)
	hexBlockNumber := fmt.Sprintf("0x%x", number)
	rpcRequest = NewRPCRequest("prometheus_getHpbNodeSnap", []interface{}{hexBlockNumber})
	rpcPayload, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(rpcPayload)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type RPCBlockResult struct {
		ID      int     `json:"id"`
		JSONRPC string  `json:"jsonrpc"`
		Result  Signers `json:"result"`
	}

	rpcBlockResp := &RPCBlockResult{}
	err = json.Unmarshal(body, rpcBlockResp)

	var addrs []common.Address
	for addr := range rpcBlockResp.Result.Addresses {
		addrs = append(addrs, addr)
	}
	return addrs, err

}
