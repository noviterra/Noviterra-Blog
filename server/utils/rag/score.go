package rag

import (
	"sort"
	"strings"
	"unicode"
)

// Score 进行词法匹配/粗排
type Score struct {
}

func NewScore() *Score {
	return &Score{}
}

var stopwords = map[string]struct{}{
	"什么": {}, "怎么": {}, "如何": {}, "为什": {}, "哪些": {}, "可以": {},
	"一个": {}, "这个": {}, "那个": {}, "我们": {}, "你们": {}, "他们": {},
	"是否": {}, "以及": {}, "但是": {}, "因为": {}, "所以": {}, "如果": {},
	"the": {}, "and": {}, "for": {}, "with": {}, "that": {}, "this": {},
	"you": {}, "are": {}, "how": {}, "what": {}, "can": {},
}

// tokenize 轻量分词：中文二元组bigram（人工智能分成：人工、工智、智能） 英文/数字按单词
func (s *Score) tokenize(c string) map[string]struct{} {
	tokens := make(map[string]struct{})

	var latin strings.Builder
	cjk := make([]rune, 0, 16)

	flushLatin := func() {
		word := strings.ToLower(strings.TrimSpace(latin.String()))
		latin.Reset()
		if len([]rune(word)) >= 2 {
			tokens[word] = struct{}{}
		}
	}
	flushCJK := func() {
		if len(cjk) == 1 {
			tokens[string(cjk)] = struct{}{}
		} else if len(cjk) == 1 {
			for i := 0; i+1 < len(cjk); i++ {
				tokens[string(cjk[i:i+2])] = struct{}{}
			}
		}
		cjk = cjk[:0]
	}

	for _, r := range strings.ToLower(c) {
		switch {
		case unicode.Is(unicode.Han, r):
			flushLatin()
			cjk = append(cjk, r)
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			flushCJK()
			latin.WriteRune(r)
		default:
			flushLatin()
			flushCJK()
		}
	}
	flushLatin()
	flushCJK()

	return tokens
}

// overlap 计算query的词Token与文档的词Token的覆盖率
func (s *Score) overlap(queryTokens, docTokens map[string]struct{}) float64 {
	if len(queryTokens) == 0 {
		return 0
	}
	hit := 0
	for token := range queryTokens {
		if _, ok := docTokens[token]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(queryTokens))
}

// KeyWords 提取候选关键词 查询几个关键Token
func (s *Score) KeyWords(query string, limit int) []string {
	if limit <= 0 {
		limit = 6
	}
	tokens := s.tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	words := make([]string, 0, len(tokens))
	for token := range tokens {
		if s.isNoiszeToken(token) {
			continue
		}
		words = append(words, token)
	}

	sort.SliceStable(words, func(i, j int) bool {
		li, lj := len([]rune(words[i])), len([]rune(words[j]))
		if li == lj { // 长度一致 按照unicode编码排序
			return words[i] < words[j]
		}
		return li > lj // 按照长度排序
	})

	if len(words) > limit {
		words = words[:limit]
	}
	return words
}

// isNoiszeToken 判断Token是否是噪声
func (s *Score) isNoiszeToken(token string) bool {
	runes := []rune(token)
	if len(runes) == 0 {
		return true
	}
	allDigit := true
	for _, r := range runes {
		if !unicode.IsDigit(r) {
			allDigit = false
			break
		}
	}
	if allDigit {
		return true
	}
	if _, ok := stopwords[token]; ok {
		return true
	}
	return false
}
