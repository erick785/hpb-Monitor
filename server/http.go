package server

import (
	"fmt"
	"math/big"
	"net/http"
	"sort"

	"github.com/erick785/hpb-monitor/rpc"
)

type MinerState struct {
	ConsensusStats States
	Round          uint64
}

type MinerStates []MinerState

func (h MinerStates) Len() int           { return len(h) }
func (h MinerStates) Less(i, j int) bool { return h[i].Round > h[j].Round }
func (h MinerStates) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

type State struct {
	Node           rpc.Node
	LostBlockCount uint64
}

type States []State

func (h States) Len() int           { return len(h) }
func (h States) Less(i, j int) bool { return h[i].LostBlockCount > h[j].LostBlockCount }
func (h States) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (m *Monitor) PrintMinerState(w http.ResponseWriter, r *http.Request) {

	result := MinerStates{}

	m.roundMap.forEach(func(key uint64, value *LostBlockInfos) {
		ms := States{}
		for _, info := range *value {
			ms = append(ms, State{
				Node:           m.nodes.GetByAddr(key, info.miner),
				LostBlockCount: uint64(info.count),
			})
		}
		sort.Sort(ms)
		result = append(result, MinerState{Round: key, ConsensusStats: ms})
	})

	sort.Sort(result)

	tableTemplate := `<h4>轮次： %v </h4>
	<table border="1">
	<tr>
		<td>缺少数量</td>
		<td>节点名称</td>
		<td>节点Addr</td>
		<td>国家</td>
		<td>地址</td>
	</tr>

	%v

	</table>`

	lineTemplate := `
	<tr>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
		<td>%v</td>
	</tr>`

	var resultStr string
	for _, v := range result {

		var lines string
		for _, state := range v.ConsensusStats {
			lines += fmt.Sprintf(lineTemplate, state.LostBlockCount,
				state.Node.NodeName, state.Node.NodeAddress.Hex(),
				state.Node.Country, state.Node.LocationDetail)
		}

		resultStr += fmt.Sprintf(tableTemplate, v.Round, lines)
	}

	lastNum, _ := rpc.GetLastBlockNumber(m.cfg.NodeEndpoint)
	fmt.Fprintf(w, `<html>

	<body>
	
	<p>高性能节点出块监控 最新高度 `+big.NewInt(lastNum).String()+` </p>

	`+resultStr+`
	</body>
	</html>
	`) //这个写入到w的是输出到客户端的
}

func (m *Monitor) HttpServer() {
	http.HandleFunc("/", m.PrintMinerState)                           //设置访问的路由
	err := http.ListenAndServe(m.cfg.HttpURL+":"+m.cfg.HttpPort, nil) //设置监听的端口
	if err != nil {
		fmt.Println("ListenAndServe Error: ", err)
	}
}
