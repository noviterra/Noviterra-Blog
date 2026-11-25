package rag

import "context"

// Embedder 向量化能力抽象 由agentcore.Engine实现
type Embedder interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Model(ctx context.Context) string
}
