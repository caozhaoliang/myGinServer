package utils

import (
	"hash/fnv"
	"math"
)

// BloomFilter 布隆过滤器
type BloomFilter struct {
	bitmap []uint64 // 位图
	m      uint64   // 位图总位数
	k      uint64   // 哈希函数个数
}

// New 根据预期元素数量 n 和误判率 p 创建布隆过滤器
func New(n uint64, p float64) *BloomFilter {
	m := uint64(math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)))
	k := uint64(math.Ceil(float64(m) / float64(n) * math.Ln2))
	if k < 1 {
		k = 1
	}
	words := (m + 63) / 64
	return &BloomFilter{
		bitmap: make([]uint64, words),
		m:      m,
		k:      k,
	}
}

// hash 双重哈希法：h_i(x) = h1(x) + i * h2(x)
func (bf *BloomFilter) hash(data []byte, i uint64) uint64 {
	h1 := fnv.New64a()
	h1.Write(data)
	v1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write(data)
	v2 := h2.Sum64()

	return (v1 + i*v2) % bf.m
}

func (bf *BloomFilter) setBit(pos uint64) {
	bf.bitmap[pos/64] |= 1 << (pos % 64)
}

func (bf *BloomFilter) getBit(pos uint64) bool {
	return bf.bitmap[pos/64]&(1<<(pos%64)) != 0
}

// Add 添加元素
func (bf *BloomFilter) Add(data []byte) {
	for i := uint64(0); i < bf.k; i++ {
		bf.setBit(bf.hash(data, i))
	}
}

// MightContain 判断元素是否可能存在
func (bf *BloomFilter) MightContain(data []byte) bool {
	for i := uint64(0); i < bf.k; i++ {
		if !bf.getBit(bf.hash(data, i)) {
			return false
		}
	}
	return true
}
