package bloomfilter

import "Noviterra/server/utils/common"

type ArticleBloomFilterParam struct {
	N uint
	P float64
}

type ArticleBloomFilter struct {
	bf              *BloomFilter
	typesConversion *common.TypesConversion
}

func NewArticleBloomFilter(param ArticleBloomFilterParam,
	typesConversion *common.TypesConversion) (*ArticleBloomFilter, error) {
	bf, err := NewBloomFilter(param.N, param.P)
	if err != nil {
		return nil, err
	}
	return &ArticleBloomFilter{
		bf:              bf,
		typesConversion: typesConversion,
	}, nil
}

// InitArticleBloomFilter 启动时需要调用一次 将数据写入到布隆过滤器中
func (abf *ArticleBloomFilter) InitArticleBloomFilter(articleID []string) {

}

func (abf *ArticleBloomFilter) AddArticle(data int64) {
	if data <= 0 {
		return
	}
	in := abf.typesConversion.Int64ConversionBytes(data)
	abf.bf.AddBloomFilterElem(in)
}

func (abf *ArticleBloomFilter) IsExistArticle(data int64) bool {
	if data <= 0 {
		return false
	}
	in := abf.typesConversion.Int64ConversionBytes(data)
	exist := abf.bf.IsExistElem(in)
	return exist
}
