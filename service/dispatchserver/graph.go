package dispatchserver

import (
	"myGinServer/models/dispatch"
	"sync"
)

// Graph 内存中的有向图，支持并发读写
type Graph struct {
	mu sync.RWMutex

	// 正向邻接表：head -> []tail
	adj map[string][]string
	// 反向邻接表：tail -> []head，加速反向可达查询
	radj map[string][]string
}

// NewGraph 从数据库全量加载边，构建内存图
func NewGraph(lines []dispatch.Line) *Graph {
	g := &Graph{
		adj:  make(map[string][]string),
		radj: make(map[string][]string),
	}

	for _, e := range lines {
		g.adj[e.AheadId] = append(g.adj[e.AheadId], e.BehindId)
		g.radj[e.BehindId] = append(g.radj[e.BehindId], e.AheadId)
	}

	return g
}

// wouldCreateCycleLocked 在持有锁的前提下，判断插入 head->tail 是否会成环
// 判断逻辑：当前图中 tail 是否能到达 head
func (g *Graph) wouldCreateCycleLocked(head, tail string) bool {
	// 自环直接判定成环
	if head == tail {
		return true
	}

	visited := make(map[string]struct{})
	queue := []string{tail}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == head {
			return true
		}

		if _, ok := visited[node]; ok {
			continue
		}
		visited[node] = struct{}{}

		// 从 node 出发的所有后继
		for _, next := range g.adj[node] {
			if _, ok := visited[next]; !ok {
				queue = append(queue, next)
			}
		}
	}

	return false
}
