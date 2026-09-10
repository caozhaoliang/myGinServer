package utils

import (
	"fmt"
)

// EdgeEntity 边实体
type EdgeEntity struct {
	AheadId  string `json:"ahead_id"`
	BehindId string `json:"behind_id"`
}

// Graph 图结构，用于存储节点和边的关系
type Graph struct {
	nodes map[string]bool     // 所有节点
	edges map[string][]string // 邻接表：节点 -> 后继节点列表
}

// NewGraph 创建新的图
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]bool),
		edges: make(map[string][]string),
	}
}

// AddNode 添加节点
func (g *Graph) AddNode(nodeId string) {
	if !g.nodes[nodeId] {
		g.nodes[nodeId] = true
		g.edges[nodeId] = []string{}
	}
}

// AddEdge 添加边（从 ahead 指向 behind）
func (g *Graph) AddEdge(aheadId, behindId string) error {
	// 检查节点是否存在
	if !g.nodes[aheadId] {
		return fmt.Errorf("节点 %s 不存在", aheadId)
	}
	if !g.nodes[behindId] {
		return fmt.Errorf("节点 %s 不存在", behindId)
	}

	// 检查是否会形成环
	if g.wouldCreateCycle(aheadId, behindId) {
		return fmt.Errorf("添加边 %s -> %s 会形成环", aheadId, behindId)
	}

	// 添加边
	g.edges[aheadId] = append(g.edges[aheadId], behindId)
	return nil
}

// wouldCreateCycle 使用广度优先遍历检查添加边后是否会形成环
func (g *Graph) wouldCreateCycle(aheadId, behindId string) bool {
	// 如果 ahead 和 behind 相同，直接形成自环
	if aheadId == behindId {
		return true
	}

	// 检查从 behind 出发能否到达 ahead（BFS）
	// 如果能到达，说明 ahead -> behind 会形成环
	return g.hasPath(behindId, aheadId)
}

// hasPath 使用广度优先遍历检查从 start 能否到达 target
func (g *Graph) hasPath(start, target string) bool {
	if start == target {
		return true
	}

	// 如果起始节点不存在，返回 false
	if !g.nodes[start] {
		return false
	}

	// BFS 队列
	queue := []string{start}
	visited := make(map[string]bool)
	visited[start] = true

	for len(queue) > 0 {
		// 出队
		current := queue[0]
		queue = queue[1:]

		// 遍历当前节点的所有后继节点
		for _, next := range g.edges[current] {
			// 如果找到了目标节点
			if next == target {
				return true
			}

			// 如果未访问过，加入队列
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	return false
}

// CheckDAG 检查整个图是否为有向无环图（DAG）
func (g *Graph) CheckDAG() bool {
	// 计算所有节点的入度
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}

	for _, successors := range g.edges {
		for _, succ := range successors {
			inDegree[succ]++
		}
	}

	// BFS：将所有入度为0的节点入队
	queue := []string{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// 记录已访问的节点数
	visitedCount := 0

	// 执行拓扑排序
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visitedCount++

		// 将当前节点的所有后继节点的入度减1
		for _, successor := range g.edges[current] {
			inDegree[successor]--
			if inDegree[successor] == 0 {
				queue = append(queue, successor)
			}
		}
	}

	// 如果访问的节点数少于总节点数，说明存在环
	return visitedCount == len(g.nodes)
}

// GetCycle 检测图中是否存在环，并返回一个环的路径（BFS方式）
func (g *Graph) GetCycle() []string {
	// 使用 Kahn 算法检测环
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}

	for _, successors := range g.edges {
		for _, succ := range successors {
			inDegree[succ]++
		}
	}

	// 初始化队列
	queue := []string{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// 拓扑排序
	var sorted []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, successor := range g.edges[current] {
			inDegree[successor]--
			if inDegree[successor] == 0 {
				queue = append(queue, successor)
			}
		}
	}

	// 如果存在环，找出环中的节点
	if len(sorted) != len(g.nodes) {
		// 找出在环中的节点（入度不为0的节点）
		cycleNodes := []string{}
		for node, degree := range inDegree {
			if degree > 0 {
				cycleNodes = append(cycleNodes, node)
			}
		}
		return cycleNodes
	}

	return nil
}

// GetEdges 获取所有边
func (g *Graph) GetEdges() []EdgeEntity {
	var edges []EdgeEntity
	for ahead, successors := range g.edges {
		for _, behind := range successors {
			edges = append(edges, EdgeEntity{
				AheadId:  ahead,
				BehindId: behind,
			})
		}
	}
	return edges
}
