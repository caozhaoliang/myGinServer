package utils

import (
	"fmt"
	"testing"
)

func TestNewGraph(t *testing.T) {
	// 创建图并添加节点
	g := NewGraph()
	nodes := []string{"A", "B", "C", "D", "E"}
	for _, node := range nodes {
		g.AddNode(node)
	}

	// 添加边（构建一个DAG）
	edges := []EdgeEntity{
		{AheadId: "A", BehindId: "B"},
		{AheadId: "A", BehindId: "C"},
		{AheadId: "B", BehindId: "D"},
		{AheadId: "C", BehindId: "D"},
		{AheadId: "D", BehindId: "E"},
	}

	for _, edge := range edges {
		err := g.AddEdge(edge.AheadId, edge.BehindId)
		if err != nil {
			fmt.Printf("添加边失败: %v\n", err)
		} else {
			fmt.Printf("成功添加边: %s -> %s\n", edge.AheadId, edge.BehindId)
		}
	}

	// 检查是否为DAG
	if g.CheckDAG() {
		fmt.Println("图是有效的DAG（有向无环图）")
	} else {
		fmt.Println("图中存在环！")
		cycle := g.GetCycle()
		fmt.Printf("环中的节点: %v\n", cycle)
	}

	// 测试添加会形成环的边
	fmt.Println("\n尝试添加会形成环的边 E -> A...")
	err := g.AddEdge("E", "A")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}
}
