package bloomfilter

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/spaolacci/murmur3"
)

type BloomFilter struct {
	bitArry *big.Int     // 比特数组
	m       uint         // 数组大小
	k       uint         // 哈希数量
	mu      sync.RWMutex // 读写锁
}

// NewBloomFilter
// @param n 预计存入的元素数量
// @param p 预期的错误率
func NewBloomFilter(n uint, p float64) (*BloomFilter, error) {
	if n == 0 || p <= 0 || p >= 1 {
		return nil, fmt.Errorf("布隆过滤器参数异常 初始化实例失败")
	}

	// 计算最优数组长度m与最优哈希数量k
	m := getOptimalBitArrayLength(n, p)
	k := getOptimalHashFunction(n, m)

	return &BloomFilter{
		bitArry: big.NewInt(0),
		m:       m,
		k:       k,
		mu:      sync.RWMutex{},
	}, nil
}

// getOptimalBitArrayLength 依据公式: m = -n * ln(p) / (ln(2) * ln(2))
func getOptimalBitArrayLength(n uint, p float64) uint {
	// Ceil 向上取整
	m := math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2))
	if m == 0 {
		m = 1
	}
	return uint(m)
}

// getOptimalHashFunction 依据公式: k = m * ln(2) / n
func getOptimalHashFunction(n, m uint) uint {
	// Round 四舍五入
	k := math.Round(float64(m) * math.Ln2 / float64(n))
	if k < 1 {
		k = 1
	}
	return uint(k)
}

// getHashPosition 得到哈希元素的下标
func (bf *BloomFilter) getHashPostition(data []byte) []uint {
	positions := make([]uint, 0, bf.k) // 将一个元素分片为k个哈希元素

	// 使用murmur3双哈希生成2个独立的哈希值 增大扰乱因子
	hash1, hash2 := murmur3.Sum128(data)

	for i := uint(0); i < bf.k; i++ {
		combined := hash1 + uint64(i)*hash2
		// 取模得到比特位下标
		position := uint(combined % uint64(bf.m))
		positions = append(positions, position)
	}

	return positions
}

// AddBloomFilterElem 添加元素到布隆过滤器中
func (bf *BloomFilter) AddBloomFilterElem(data []byte) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if len(data) == 0 {
		return
	}

	// todo: 记录每个元素的插入频率
	position := bf.getHashPostition(data)
	for _, pos := range position {
		bf.bitArry.SetBit(bf.bitArry, int(pos), 1)
	}
}

// IsExistElem 判断元素是否存在
func (bf *BloomFilter) IsExistElem(data []byte) bool {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if len(data) <= 0 {
		return false
	}

	positions := bf.getHashPostition(data)
	for _, pos := range positions {
		if bf.bitArry.Bit(int(pos)) == 0 {
			return false
		}
	}
	return true
}
