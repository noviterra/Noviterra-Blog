package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"regexp"
	"strings"
)

var (
	htmlTagPattern   = regexp.MustCompile(`(?s)<[^>]*>`)
	spacePattern     = regexp.MustCompile(`[ \t\f\v]+`)
	blankLinePattern = regexp.MustCompile(`\n{3,}`)
)

type Chunk struct {
	Index     int
	Content   string
	CharCount int
}

type Chunker struct {
	DefaultChunkSize    int // 每块的字符数
	DefaultChunkOverlap int // 相邻块重叠字符数
}

func NewChunker() *Chunker {
	return &Chunker{
		DefaultChunkOverlap: 600,
		DefaultChunkSize:    100,
	}
}

// Normalize 文本清洗
func (c *Chunker) Normalize(content string) string {
	text := strings.TrimSpace(content)
	if text == "" {
		return ""
	}
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		text = htmlTagPattern.ReplaceAllString(text, " ")
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = html.UnescapeString(text)
	text = spacePattern.ReplaceAllString(text, " ")
	text = blankLinePattern.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// Split 文本切片
func (c *Chunker) Split(content string, size, overlap int) []Chunk {
	text := c.Normalize(content)
	if text == "" {
		return nil
	}
	if size <= 0 {
		size = c.DefaultChunkSize
	}
	if overlap <= 0 {
		overlap = c.DefaultChunkOverlap
	}

	runes := []rune(text)
	if len(runes) <= size {
		return []Chunk{
			{
				Index:     0,
				Content:   text,
				CharCount: len(runes),
			},
		}
	}

	chunks := make([]Chunk, 0, len(runes)/size+1)
	start := 0
	for start < len(runes) {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		} else {
			end = c.pickBoundary(runes, size, end, size)
		}

		piece := strings.TrimSpace(string(runes[start:end]))
		if piece != "" {
			chunks = append(chunks, Chunk{
				Index:     len(chunks),
				Content:   piece,
				CharCount: len([]rune(piece)),
			})
		}

		if end >= len(runes) {
			break
		}
		next := end - overlap
		if next <= start {
			next = start + 1
		}
		start = next
	}
	return chunks
}

// pickBoundary 尽量在句子边界就结束
func (c *Chunker) pickBoundary(runes []rune, start, end, size int) int {
	lower := end - size/5
	if lower < start {
		lower = start + 1
	}
	for i := end - 1; i >= lower; i-- {
		switch runes[i] {
		case '\n', '。', '！', '？', '；', '.', '!', '?', ';':
			return i + 1
		}
	}
	return end
}

// EmbedInput 构造Embedding输入 把标题和正文拼起来
func (c *Chunker) EmbedInput(title, content string) string {
	t := strings.TrimSpace(title)
	if t == "" {
		return content
	}
	runes := []rune(t)
	if len(runes) > 80 {
		t = string(runes[:80])
	}
	return t + "\n" + content
}

// SourchHash 内容指纹 文档去除等
func (c *Chunker) SourchHash(title, content string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(title) + "\x00" + c.Normalize(content)))
	return hex.EncodeToString(sum[:])
}

// Snippet 生成短摘要  作为搜索结果的预览文本
func (c *Chunker) Snippet(content string, maxRunes int) string {
	text := strings.TrimSpace(content)
	if maxRunes <= 0 {
		maxRunes = 160
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}
