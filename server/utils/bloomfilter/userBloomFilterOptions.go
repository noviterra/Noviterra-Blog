package bloomfilter

type UserBloomFilter struct {
	bf *BloomFilter
}

func NewUserBloomFilter(n uint, p float64) (*UserBloomFilter, error) {
	bf, err := NewBloomFilter(n, p)
	if err != nil {
		return nil, err
	}
	return &UserBloomFilter{
		bf: bf,
	}, nil
}

// InitUserBloomFilter 启动时需要调用一次 将库中的用户ID插入到布隆过滤器中
func (ubf *UserBloomFilter) InitUserBloomFilter(userID []string) {

}

func (ubf *UserBloomFilter) AddUsers(data string) {
	if len(data) <= 0 {
		return
	}
	in := []byte(data)
	ubf.bf.AddBloomFilterElem(in)
}

func (ubf *UserBloomFilter) IsExistUser(data string) bool {
	if len(data) <= 0 {
		return false
	}
	in := []byte(data)
	exist := ubf.bf.IsExistElem(in)
	return exist
}
