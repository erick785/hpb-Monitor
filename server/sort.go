package server

import (
	"container/heap"
	"sort"
	"sync"

	"github.com/erick785/hpb-monitor/rpc"
	"github.com/hpb-project/go-hpb/common"
)

type CheckPointHeap []uint64

func (h CheckPointHeap) Len() int           { return len(h) }
func (h CheckPointHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h CheckPointHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *CheckPointHeap) Push(x interface{}) {
	for _, v := range *h {
		if v == x.(uint64) {
			return
		}
	}
	*h = append(*h, x.(uint64))
}

func (h *CheckPointHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type SortMap struct {
	items map[uint64]map[common.Address]rpc.Node
	index *CheckPointHeap
	cache int
	sync.RWMutex
}

func newSortedMap(cache int) *SortMap {
	return &SortMap{
		items: make(map[uint64]map[common.Address]rpc.Node),
		index: new(CheckPointHeap),
		cache: cache,
	}
}

func (m *SortMap) GetByAddr(point uint64, addr common.Address) rpc.Node {
	m.RLock()
	defer m.RUnlock()
	return m.items[point][addr]
}

func (m *SortMap) Get(point uint64) map[common.Address]rpc.Node {
	m.RLock()
	defer m.RUnlock()
	return m.items[point]
}

func (m *SortMap) Put(point uint64, nodes map[common.Address]rpc.Node) {
	m.Lock()
	defer m.Unlock()
	if m.items[point] == nil {
		heap.Push(m.index, point)
	}
	m.items[point] = nodes

	if len(m.items) > m.cache {
		sort.Sort(*m.index)
		threshold := len(m.items) - m.cache
		for size := len(m.items); size > threshold; size-- {
			delete(m.items, (*m.index)[size-1])
		}
		*m.index = (*m.index)[:threshold]
		heap.Init(m.index)
	}
}

type LostBlockInfo struct {
	miner common.Address
	count int
	point uint64
}

type LostBlockInfos []*LostBlockInfo

func (l *LostBlockInfos) add(info LostBlockInfo) {
	for _, v := range *l {
		if v.miner == info.miner {
			v.count++
			return
		}
	}

	*l = append(*l, &info)
}

type roundMap struct {
	items map[uint64]*LostBlockInfos
	cache int
	index *CheckPointHeap
	sync.RWMutex
}

func newRoundMap(cache int) *roundMap {
	return &roundMap{
		items: make(map[uint64]*LostBlockInfos),
		index: new(CheckPointHeap),
		cache: cache,
	}
}

func (m *roundMap) Get(point uint64) *LostBlockInfos {
	m.RLock()
	defer m.RUnlock()
	return m.items[point]
}

func (m *roundMap) forEach(cb func(key uint64, value *LostBlockInfos)) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.items {
		cb(k, v)
	}
}

func (m *roundMap) Len() int {
	m.Lock()
	defer m.Unlock()
	return len(m.items)
}

func (m *roundMap) Put(info LostBlockInfo) {
	m.Lock()
	defer m.Unlock()

	if m.items[info.point] == nil {
		heap.Push(m.index, info.point)
	}
	if infos, ok := m.items[info.point]; ok {
		infos.add(info)
	} else {
		newInfos := new(LostBlockInfos)
		newInfos.add(info)
		m.items[info.point] = newInfos
	}

	if len(m.items) > m.cache {
		sort.Sort(*m.index)
		threshold := len(m.items) - m.cache
		for size := len(m.items); size > threshold; size-- {
			delete(m.items, (*m.index)[size-1])
		}
		*m.index = (*m.index)[:threshold]
		heap.Init(m.index)
	}

}
