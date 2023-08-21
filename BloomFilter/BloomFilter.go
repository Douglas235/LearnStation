package BloomFilter

import (
	"github.com/bits-and-blooms/bitset"
	"math"
)

type BloomFilter struct {
	m uint           //哈希空间大小
	k uint           //哈希函数个数
	b *bitset.BitSet //bitmap
}

func max(x, y uint) uint {
	if x > y {
		return x
	}
	return y
}
func New(m uint, k uint) *BloomFilter {
	return &BloomFilter{max(1, m), max(1, k), bitset.New(m)}
}

// 初始化过滤器
func NewWithEstimates(n uint, fp float64) *BloomFilter {
	m, k := EstimateParameters(n, fp)
	return New(m, k)
}
func EstimateParameters(n uint, p float64) (m uint, k uint) {
	m = uint(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	k = uint(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return

}

// 添加一个元素到布隆过滤器
func (f *BloomFilter) Add(data []byte) *BloomFilter {
	h := bashHashes(data)
	for i := uint(0); i < f.k; i++ {
		f.b.Set(f.location(h, i))
	}
	return f
}

// bashHashes 返回哈希值，调用murmur3算法
func bashHashes(data []byte) [4]uint64 {
	var d digest128 // murmur hashing
	hash1, hash2, hash3, hash4 := d.sum256(data)
	return [4]uint64{
		hash1, hash2, hash3, hash4,
	}
}

// 返回哈希值对应的bitmap的位置索引，将bitmap对应位置进行设置，就可以将元素添加到布隆过滤器
func (f *BloomFilter) location(h [4]uint64, i uint) uint {
	return uint(location(h, i) % uint64(f.m))
}

func location(h [4]uint64, i uint) uint64 {
	ii := uint64(i)
	return h[ii%2] + ii*h[2+(((ii+(ii%2))%4)/2)]

}

// 检测一个元素是否存在于布隆过滤器
func (f *BloomFilter) Test(data []byte) bool {
	h := bashHashes(data)
	for i := uint(0); i < f.k; i++ {
		if !f.b.Test(f.location(h, i)) {
			return false
		}
	}
	return true

}
