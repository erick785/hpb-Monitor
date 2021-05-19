package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/hpb-project/go-hpb/common"
	"github.com/hpb-project/go-hpb/common/hexutil"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
	"github.com/urfave/cli"
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

type BlockResult struct {
	ParentHash     common.Hash    `json:"parentHash"`
	Difficulty     string         `json:"difficulty"`
	Miner          common.Address `json:"miner"`
	HardwareRandom hexutil.Bytes  `json:"hardwareRandom"`
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

type Signers struct {
	Addresses map[common.Address]struct{} `json:"signers"`
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

type Node struct {
	NodeName       string
	NodeAddress    common.Address
	LockAmount     float64
	Country        string
	LocationDetail string
}

func GetNodeListFromHpbScan(endpoint string) (map[common.Address]Node, error) {
	req, err := http.NewRequest("POST", endpoint+"/HpbScan/node/list",
		strings.NewReader(`{"currentPage": 1, "pageSize": 2000, "nodeType": "hpbnode", "country": ""}`))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json;charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	page := gojsonq.New().FromString(string(body)).Nth(3)
	result := gojsonq.New().FromInterface(page).Find("list")

	nodes := make(map[common.Address]Node)

	for _, v := range result.([]interface{}) {
		node := v.(map[string]interface{})
		nodes[common.HexToAddress(node["nodeAddress"].(string))] = Node{
			NodeName:       node["nodeName"].(string),
			NodeAddress:    common.HexToAddress(node["nodeAddress"].(string)),
			LockAmount:     node["lockAmount"].(float64),
			Country:        node["country"].(string),
			LocationDetail: node["locationDetail"].(string),
		}

	}
	return nodes, err
}

type Monitor struct {
	nodeEndpoint string
	scanEndpoint string
	startBlock   int64
	endBlock     int64
	sigKillChan  chan os.Signal
}

func (m *Monitor) start(c *cli.Context) error {
	if m.startBlock == 0 || m.startBlock > m.endBlock || m.endBlock == 0 {
		return fmt.Errorf("start monitor args error ,startBlock %d,endBlock %d ", m.startBlock, m.endBlock)
	}

	m.ShutDown()

	nodes, err := GetNodeListFromHpbScan(m.scanEndpoint)
	if err != nil {
		return err
	}

	var parentBlockResult *BlockResult

	for m.startBlock <= m.endBlock {
		blockResult, err := GetBlockMinerAndHardwareRandom(m.startBlock, m.nodeEndpoint)
		if err != nil {
			return err
		}

		if blockResult.Difficulty != "0x2" {
			signers, err := GetHpbNodeSnapSigners(m.startBlock, m.nodeEndpoint)
			if err != nil {
				return err
			}
			if parentBlockResult == nil {
				parentBlockResult, err = GetBlockMinerAndHardwareRandom(m.startBlock-1, m.nodeEndpoint)
				if err != nil {
					return err
				}
			}

			addr := m.CalculateMiner(signers, parentBlockResult.Miner, blockResult.HardwareRandom)

			if node, ok := nodes[*addr]; ok {
				fmt.Printf("Block number %d should be miner %s, nodeName %s, lockAmount %v, country %s, locationDetail %s \n",
					m.startBlock, addr.Hex(), node.NodeName, node.LockAmount, node.Country, node.LocationDetail)
			} else {
				fmt.Printf("WARN: Block number %d should be miner %s ,but not found in hpb node list \n", m.startBlock, addr.Hex())

			}

		}

		m.startBlock++
		parentBlockResult = blockResult
	}
	return nil
}

func (m *Monitor) loop(c *cli.Context) error {
	m.ShutDown()

	nodes, err := GetNodeListFromHpbScan(m.scanEndpoint)
	if err != nil {
		return err
	}

	lastNum, err := GetLastBlockNumber(m.nodeEndpoint)
	if err != nil {
		return err
	}
	fmt.Println("HPB last number ", lastNum)

	startNum := lastNum - 10
	var parentBlockResult *BlockResult

	for {
		time.Sleep(3 * time.Second)
		lastNum, err := GetLastBlockNumber(m.nodeEndpoint)
		if err != nil {
			return err
		}

		if lastNum-10 < startNum {
			continue
		}

		blockResult, err := GetBlockMinerAndHardwareRandom(startNum, m.nodeEndpoint)
		if err != nil {
			return err
		}

		if blockResult.Difficulty != "0x2" {
			signers, err := GetHpbNodeSnapSigners(startNum, m.nodeEndpoint)
			if err != nil {
				return err
			}

			if parentBlockResult == nil {
				parentBlockResult, err = GetBlockMinerAndHardwareRandom(startNum-1, m.nodeEndpoint)
				if err != nil {
					return err
				}
			}

			addr := m.CalculateMiner(signers, parentBlockResult.Miner, blockResult.HardwareRandom)

			if node, ok := nodes[*addr]; ok {
				fmt.Printf("Block number %d should be miner %s, nodeName %s, lockAmount %v, country %s, locationDetail %s \n",
					startNum, addr.Hex(), node.NodeName, node.LockAmount, node.Country, node.LocationDetail)
			} else {
				fmt.Printf("WARN: Block number %d should be miner %s, but not found in hpb node list \n", startNum, addr.Hex())

			}

		}

		parentBlockResult = blockResult
		startNum++
	}
}

func (m *Monitor) CalculateMiner(Signers []common.Address, parentBlockMiner common.Address, hardwareRandom []byte) *common.Address {
	var chooseSigners common.Addresses
	for _, addr := range Signers {
		if addr != parentBlockMiner {
			chooseSigners = append(chooseSigners, addr)
		}
	}
	sort.Sort(chooseSigners)
	index := new(big.Int).SetBytes(hardwareRandom).Uint64() % uint64(len(chooseSigners))
	return &chooseSigners[index]
}

func (m *Monitor) ShutDown() {
	signal.Notify(m.sigKillChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-m.sigKillChan //阻塞等待
		fmt.Println("system exit")
		os.Exit(0)
	}()
}

func LoopCommand(sigKillChan chan os.Signal) cli.Command {
	m := new(Monitor)
	m.sigKillChan = sigKillChan
	return cli.Command{
		Name:   "loop",
		Usage:  "Monitor the HPB block chain find who did not produce block",
		Action: m.loop,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "hpb-host",
				Value:       "http://hpbnode.com",
				Usage:       "host endpoint of the hpb node",
				Destination: &m.nodeEndpoint,
			},
			cli.StringFlag{
				Name:        "hpb-scan-host",
				Value:       "http://hpbscan.org",
				Usage:       "host endpoint of the hpb scan",
				Destination: &m.scanEndpoint,
			},
		},
	}
}

func StartCommand(sigKillChan chan os.Signal) cli.Command {
	m := new(Monitor)
	m.sigKillChan = sigKillChan
	return cli.Command{
		Name:   "start",
		Usage:  "Monitor the HPB block chain find who did not produce block",
		Action: m.start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "hpb-host",
				Value:       "http://hpbnode.com",
				Usage:       "host endpoint of the hpb node",
				Destination: &m.nodeEndpoint,
			},
			cli.StringFlag{
				Name:        "hpb-scan-host",
				Value:       "http://hpbscan.org",
				Usage:       "host endpoint of the hpb scan",
				Destination: &m.scanEndpoint,
			},
			cli.Int64Flag{
				Name:        "start-block",
				Value:       0,
				Usage:       "block number of the first block to begin scanning",
				Destination: &m.startBlock,
			},
			cli.Int64Flag{
				Name:        "end-block",
				Value:       0,
				Usage:       "block number of the block to end scanning",
				Destination: &m.endBlock,
			},
		},
	}

}

func main() {
	sigKillChan := make(chan os.Signal, 1)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("crashed with: %v", r)
		}
	}()

	app := cli.App{
		Name:    "hpb-monitor",
		Usage:   "Monitor the HPB block chain find who did not produce block",
		Version: "0.0.1",
		Commands: []cli.Command{
			StartCommand(sigKillChan),
			LoopCommand(sigKillChan),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("unexpected error: ", err)
	}

}
