package server

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/erick785/hpb-monitor/rpc"
	"github.com/hpb-project/go-hpb/common"
	"github.com/urfave/cli"
)

const (
	loopInterval              = time.Second * 3
	nodesCache                = 5
	delayBlockNum             = 10
	HpbNodeCheckpointInterval = 200
)

type Monitor struct {
	cfg              *Config
	latestCheckPoint uint64
	nodes            *SortMap
	roundMap         *roundMap
	sigKillChan      chan os.Signal
}

func NewMonitor(c *Config, sigKillChan chan os.Signal) *Monitor {
	return &Monitor{
		cfg:         c,
		nodes:       newSortedMap(nodesCache),
		roundMap:    newRoundMap(nodesCache),
		sigKillChan: sigKillChan,
	}
}

func (m *Monitor) LatestCheckPointNumber(number int64) uint64 {
	return uint64(math.Floor(float64(number/HpbNodeCheckpointInterval))) * HpbNodeCheckpointInterval
}

func (m *Monitor) Start(c *cli.Context) error {
	m.ShutDown()

	nodes, err := rpc.GetNodeListFromHpbScan(m.cfg.ScanEndpoint)
	if err != nil {
		return err
	}

	var parentBlockResult *rpc.BlockResult

	for m.cfg.StartBlock <= m.cfg.EndBlock {
		blockResult, err := rpc.GetBlockMinerAndHardwareRandom(m.cfg.StartBlock, m.cfg.NodeEndpoint)
		if err != nil {
			return err
		}

		if blockResult.Difficulty != "0x2" {
			signers, err := rpc.GetHpbNodeSnapSigners(m.cfg.StartBlock, m.cfg.NodeEndpoint)
			if err != nil {
				return err
			}
			if parentBlockResult == nil {
				parentBlockResult, err = rpc.GetBlockMinerAndHardwareRandom(m.cfg.StartBlock-1, m.cfg.NodeEndpoint)
				if err != nil {
					return err
				}
			}

			addr := m.CalculateMiner(signers, parentBlockResult.Miner, blockResult.HardwareRandom)

			if node, ok := nodes[*addr]; ok {
				fmt.Printf("Block number %d should be miner %s, nodeName %s, lockAmount %v, country %s, locationDetail %s \n",
					m.cfg.StartBlock, addr.Hex(), node.NodeName, node.LockAmount, node.Country, node.LocationDetail)
			} else {
				fmt.Printf("WARN: Block number %d should be miner %s ,but not found in hpb node list \n", m.cfg.StartBlock, addr.Hex())

			}

		}

		m.cfg.StartBlock++
		parentBlockResult = blockResult
	}
	return nil
}

func (m *Monitor) Loop(c *cli.Context) error {
	m.ShutDown()

	go m.HttpServer()

	nodes, err := rpc.GetNodeListFromHpbScan(m.cfg.ScanEndpoint)
	if err != nil {
		return err
	}

	lastNum, err := rpc.GetLastBlockNumber(m.cfg.NodeEndpoint)
	if err != nil {
		return err
	}

	startNum := lastNum - delayBlockNum
	m.latestCheckPoint = m.LatestCheckPointNumber(startNum)
	m.nodes.Put(m.latestCheckPoint, nodes)

	fmt.Println("HPB last number ", lastNum, "start number ", startNum, "checkoutPoit ", m.latestCheckPoint)

	var parentBlockResult *rpc.BlockResult

	for {
		if lastNum-delayBlockNum < startNum {
			time.Sleep(loopInterval)
			lastNum, _ = rpc.GetLastBlockNumber(m.cfg.NodeEndpoint)
			continue
		}

		blockResult, err := rpc.GetBlockMinerAndHardwareRandom(startNum, m.cfg.NodeEndpoint)
		if err != nil {
			fmt.Println("warn: GetBlockMinerAndHardwareRandom", err)
			time.Sleep(loopInterval)
			continue
		}

		fmt.Println("---get block->", startNum, blockResult.Difficulty)

		if blockResult.Difficulty != "0x2" {

			if m.LatestCheckPointNumber(startNum) != m.latestCheckPoint {

				nodes, err := rpc.GetNodeListFromHpbScan(m.cfg.ScanEndpoint)
				if err != nil {
					fmt.Println("warn: GetNodeListFromHpbScan", err)
					time.Sleep(loopInterval)
					continue
				}
				m.latestCheckPoint = m.LatestCheckPointNumber(startNum)
				m.nodes.Put(m.latestCheckPoint, nodes)
				fmt.Println("---update nodes->", startNum, m.latestCheckPoint, nodes)
			}

			signers, err := rpc.GetHpbNodeSnapSigners(startNum, m.cfg.NodeEndpoint)
			if err != nil {
				fmt.Println("warn: GetHpbNodeSnapSigners", err)
				time.Sleep(loopInterval)
				continue
			}

			if parentBlockResult == nil {
				parentBlockResult, err = rpc.GetBlockMinerAndHardwareRandom(startNum-1, m.cfg.NodeEndpoint)
				if err != nil {
					fmt.Println("warn: GetBlockMinerAndHardwareRandom parentBlockResult", err)
					time.Sleep(loopInterval)
					continue
				}
			}

			addr := m.CalculateMiner(signers, parentBlockResult.Miner, blockResult.HardwareRandom)

			nodes := m.nodes.Get(m.latestCheckPoint)
			if node, ok := nodes[*addr]; ok {
				fmt.Printf("Block number %d should be miner %s, nodeName %s, lockAmount %v, country %s, locationDetail %s \n",
					startNum, addr.Hex(), node.NodeName, node.LockAmount, node.Country, node.LocationDetail)

				m.roundMap.Put(LostBlockInfo{miner: *addr, count: 1, point: m.latestCheckPoint})

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
