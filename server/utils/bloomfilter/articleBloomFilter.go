package bloomfilter

type ArticleBloomFilter struct {
	bf *BloomFilter
}

func NewArticleBloomFilter(n uint, p float64) (*ArticleBloomFilter, error) {
	bf, err := NewBloomFilter(n, p)
	if err != nil {
		return nil, err
	}
	return &ArticleBloomFilter{
		bf: bf,
	}, nil
}

// InitArticleBloomFilter 启动时需要调用一次 将数据写入到布隆过滤器中
func (abf *ArticleBloomFilter) InitArticleBloomFilter(articleID []string) {

}

func (abf *ArticleBloomFilter) AddArticle(data string) {
	if len(data) <= 0 {
		return
	}
	in := []byte(data)
	abf.bf.AddBloomFilterElem(in)
}

func (abf *ArticleBloomFilter) IsExistArticle(data string) bool {
	if len(data) <= 0 {
		return false
	}
	in := []byte(data)
	exist := abf.bf.IsExistElem(in)
	return exist
}
