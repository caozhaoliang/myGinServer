package utils

import (
	"fmt"
	"testing"
)

func TestBloom(t *testing.T) {
	bf := New(10000, 0.01) // 预期 1 万元素，误判率 1%

	bf.Add([]byte("hello"))
	bf.Add([]byte("world"))

	fmt.Println(bf.MightContain([]byte("hello")))  // true
	fmt.Println(bf.MightContain([]byte("world")))  // true
	fmt.Println(bf.MightContain([]byte("golang"))) // 大概率 false，也可能 true（误判）

	// 统计误判率
	fp := 0
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("not-added-%d", i))
		if bf.MightContain(key) {
			fp++
		}
	}
	fmt.Printf("误判数: %d / 10000\n", fp)
}
