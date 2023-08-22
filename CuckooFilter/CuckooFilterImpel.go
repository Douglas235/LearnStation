package CuckooFilter

import (
	metro "github.com/dgryski/go-metro"
	"math/bits"
	"math/rand"
)

// The simplest cuckoo hash structure is a one-dimensional array structure, where two hashes map incoming elements to two positions in the array
// If one of the two positions is empty, you can put the element directly into it. But if both positions are full, it will have to The Other Boleyn Girl, kick out one at random, and take over the position itself.
const (
	buckectSize    = 4
	nullFp         = 0
	maxCuckooCount = 500
)

type fingerprint byte

// 二维数组 大小4
type bukect [buckectSize]fingerprint

type Filter struct {
	bukects []bukect
	//已插入的元素
	count uint
	//数组buckets长度中对应二进制包含0的个数
	bucketPow uint
}

var (
	altHash = [256]uint{}
	masks   = [65]uint{}
)

func init() {
	for i := 0; i < 256; i++ {
		//缓存指纹信息的hash,`metro.Hash64([]byte{byte(i)}, 1337)` 是一个哈希函数调用，将一个字节切片作为输入，并使用哈希算法生成一个64位的哈希值
		//指定的种子值（1337）计算哈希值。请确保你的代码中已经导入了正确的包，并且`metro`包中存在`Hash64`函数的定义
		altHash[i] = (uint(metro.Hash64([]byte{byte(i)}, 1337)))
	}
	for i := uint(0); i <= 64; i++ {
		//取hash值的最后n位
		masks[i] = (1 << i) - 1
	}
}
func NewFilter(capacity uint) *Filter {
	//计算buckets数组大小
	//getNextPow2将capacity 调整到 2 的指数倍，如果传入的 capacity 是 9 ，那么调用 getNextPow2 后会返回 16
	capacity = getNextPow2(uint(capacity)) / buckectSize
	if capacity == 0 {
		capacity = 1
	}
	bukects := make([]bukect, capacity)
	return &Filter{
		bukects: bukects,
		count:   0,
		// 获取 buckets 数组大小的二进制中以 0 结尾的个数,因为 capacity 是 2 的指数倍，所以 bucketPow 是 capacity 二进制的位数减 1
		bucketPow: uint(bits.TrailingZeros(capacity)),
	}
}
func (cf *Filter) Insert(data []byte) bool {
	// 获取 data 的 fingerprint 以及 位置 i1
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	// 将 fingerprint 插入到 Filter 的 buckets 数组中
	if cf.insert(fp, i1) {
		return true
	}
	// 获取位置 i2
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	// 将 fingerprint 插入到 Filter 的 buckets 数组中
	if cf.insert(fp, i2) {
		return true
	}
	// 插入失败，那么进行循环插入踢出元素
	return cf.reinsert(fp, randi(i1, i2))
}
func (cf *Filter) insert(fp fingerprint, i uint) bool {
	// 获取 buckets 中的槽位进行插入
	if cf.bukects[i].insert(fp) {
		// Filter 中元素个数+1
		cf.count++
		return true
	}
	return false
}
func (b *bukect) insert(fp fingerprint) bool {
	//遍历槽位的4个元素，如果为空则插入
	for i, tfp := range b {
		if tfp == nullFp {
			b[i] = fp
			return true
		}
	}
	return false
}

// 循环提出插入
// reinsert 方法随机获取槽位 i1、i2 中的某个位置进行抢占，然后将老元素踢出并循环重复插入
func (cf *Filter) reinsert(fp fingerprint, i uint) bool {
	//默认循环500次
	//这里会最大循环 500 次获取槽位信息。因为每个槽位最多可以存放 4 个元素，
	//所以使用 rand 随机从 4 个位置中取一个元素踢出，然后将当次循环的元素插入，
	//再获取被踢出元素的另一个槽位信息，再调用 insert 进行插入
	for k := 0; k < maxCuckooCount; k++ {
		//随机从槽位中选取一个元素
		j := rand.Intn(buckectSize)
		oldfp := fp
		//获取槽位中的值
		fp = cf.bukects[i][j]
		//将当前循环的值插入
		cf.bukects[i][j] = oldfp
		//获取另一个槽位
		i = getAltIndex(fp, i, cf.bucketPow)
		if cf.insert(fp, i) {
			return true
		}
	}
	return true
}

// 查询数据
func (cf *Filter) Lookup(data []byte) bool {
	// 获取槽位 i1 以及指纹信息
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	// 遍历槽位中 4 个位置，查看有没有相同元素
	if cf.bukects[i1].getFingerprintIndex(fp) > -1 {
		return true
	}
	// 获取另一个槽位 i2
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	// 遍历槽位 i2 中 4 个位置，查看有没有相同元素
	return cf.bukects[i2].getFingerprintIndex(fp) > -1
}

func (b *bukect) getFingerprintIndex(fp fingerprint) int {
	for i, tfp := range b {
		if tfp == fp {
			return i
		}
	}
	return -1
}

// 删除数据，只是抹掉该槽位上的指纹信息
func (cf Filter) Delete(data []byte) bool {
	// 获取槽位 i1 以及指纹信息
	i1, fp := getIndexAndFingerprint(data, cf.bucketPow)
	// 尝试删除指纹信息
	if cf.delete(fp, i1) {
		return true
	}
	// 获取槽位 i2
	i2 := getAltIndex(fp, i1, cf.bucketPow)
	// 尝试删除指纹信息
	return cf.delete(fp, i2)
}
func (cf Filter) delete(fp fingerprint, i uint) bool {
	// 遍历槽位 4个元素，尝试删除指纹信息
	if cf.bukects[i].delete(fp) {
		if cf.count > 0 {
			cf.count--
		}
		return true
	}
	return false
}
func (b bukect) delete(fp fingerprint) bool {
	for i, tfp := range b {
		// 指纹信息相同，将此槽位置空
		if tfp == fp {
			b[i] = nullFp
			return true
		}
	}
	return false

}
func getNextPow2(n uint) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}
func getIndexAndFingerprint(data []byte, bucketPow uint) (uint, fingerprint) {
	// 将 data 进行hash
	hash := metro.Hash64(data, 1337)
	// 取 hash 的指纹信息
	fp := getFingerprint(hash)
	// 取 hash 高32位，对 hash 的高32位进行取与获取槽位 i1
	//masks[bucketPow] 获取的二进制结果全是 1 ，用来取 hash 的低位的值。
	i1 := uint(hash>>32) & masks[bucketPow]
	return i1, fingerprint(fp)
}

// 取 hash 的指纹信息
func getFingerprint(hash uint64) byte {
	fp := byte(hash%255 + 1)
	return fp
}

// getAltIndex 中获取槽位是通过使用 altHash 来获取指纹信息的 hash 值，然后取异或后返回槽位值
// 由于异或的特性，所以传入的不管是槽位 i1，还是槽位 i2 都可以返回对应的另一个槽位
func getAltIndex(fp fingerprint, i uint, bucketPow uint) uint {
	mask := masks[bucketPow]
	hash := altHash[fp] & mask
	return (i & mask) ^ hash
}
func randi(i1, i2 uint) uint {
	if rand.Intn(2) == 0 {
		return i1
	}
	return i2
}
